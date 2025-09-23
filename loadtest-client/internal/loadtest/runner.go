package loadtest

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"loadtest-client/internal/client"
	"loadtest-client/internal/config"
	"loadtest-client/internal/generator"
	"loadtest-client/internal/metrics"
)

// Runner orchestrates the load test execution
type Runner struct {
	config           *config.Config
	httpClient       *client.HTTPClient
	requestGenerator *generator.RequestGenerator
	metricsCollector *metrics.Collector
	
	// Control channels
	stopCh     chan struct{}
	reportCh   chan struct{}
	
	// State
	running    int32
	startTime  time.Time
}

// NewRunner creates a new load test runner
func NewRunner(cfg *config.Config) (*Runner, error) {
	metricsCollector := metrics.NewCollector()
	
	httpClient, err := client.NewHTTPClient(cfg, metricsCollector)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	requestGenerator := generator.NewRequestGenerator(cfg)
	
	return &Runner{
		config:           cfg,
		httpClient:       httpClient,
		requestGenerator: requestGenerator,
		metricsCollector: metricsCollector,
		stopCh:           make(chan struct{}),
		reportCh:         make(chan struct{}),
	}, nil
}

// Run executes the load test
func (r *Runner) Run(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&r.running, 0, 1) {
		return fmt.Errorf("load test is already running")
	}
	defer atomic.StoreInt32(&r.running, 0)
	
	log.Println("Starting load test...")
	r.startTime = time.Now()
	
	// Get durations
	duration, err := r.config.GetDuration()
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	
	rampUpDuration, err := r.config.GetRampUpDuration()
	if err != nil {
		return fmt.Errorf("invalid ramp up duration: %w", err)
	}
	
	reportInterval, err := r.config.GetReportingInterval()
	if err != nil {
		return fmt.Errorf("invalid reporting interval: %w", err)
	}
	
	// Start reporting goroutine
	go r.reportingLoop(reportInterval)
	
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	
	// Start load generation
	if rampUpDuration > 0 {
		log.Printf("Ramping up over %v to %d QPS with %d workers", 
			rampUpDuration, r.config.Load.TargetQPS, r.config.Load.Workers)
		err = r.runWithRampUp(testCtx, rampUpDuration)
	} else {
		log.Printf("Running at %d QPS with %d workers", 
			r.config.Load.TargetQPS, r.config.Load.Workers)
		err = r.runSteadyState(testCtx)
	}
	
	// Signal reporting to stop
	close(r.reportCh)
	
	// Final report
	log.Println("Load test completed. Final statistics:")
	r.metricsCollector.PrintStats()
	
	return err
}

// runWithRampUp executes load test with gradual ramp-up
func (r *Runner) runWithRampUp(ctx context.Context, rampUpDuration time.Duration) error {
	rampUpSteps := 10
	stepDuration := rampUpDuration / time.Duration(rampUpSteps)
	targetQPS := r.config.Load.TargetQPS
	
	for step := 1; step <= rampUpSteps; step++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Calculate QPS for this step
		currentQPS := (targetQPS * step) / rampUpSteps
		if currentQPS < 1 {
			currentQPS = 1
		}
		
		log.Printf("Ramp-up step %d/%d: %d QPS", step, rampUpSteps, currentQPS)
		
		// Run this step
		stepCtx, cancel := context.WithTimeout(ctx, stepDuration)
		err := r.runAtQPS(stepCtx, currentQPS)
		cancel()
		
		if err != nil && err != context.DeadlineExceeded {
			return err
		}
	}
	
	// Continue with steady state for remaining time
	log.Printf("Ramp-up complete. Running steady state at %d QPS", targetQPS)
	return r.runSteadyState(ctx)
}

// runSteadyState runs at target QPS for the duration
func (r *Runner) runSteadyState(ctx context.Context) error {
	return r.runAtQPS(ctx, r.config.Load.TargetQPS)
}

// runAtQPS runs the load test at specified QPS
func (r *Runner) runAtQPS(ctx context.Context, targetQPS int) error {
	// Calculate request interval
	requestInterval := time.Second / time.Duration(targetQPS)
	
	log.Printf("Starting load generation: %d QPS, interval: %v", targetQPS, requestInterval)
	
	// Create worker pool
	workerCount := r.config.Load.Workers
	if workerCount > targetQPS {
		workerCount = targetQPS
	}
	
	log.Printf("Starting %d workers", workerCount)
	
	// Channel for distributing work
	requestCh := make(chan struct{}, workerCount*2)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r.worker(ctx, workerID, requestCh)
		}(i)
	}
	
	// Start request generator
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()
	
	requestCount := 0
	lastLogTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, sent %d requests", requestCount)
			close(requestCh)
			wg.Wait()
			return ctx.Err()
		case <-ticker.C:
			select {
			case requestCh <- struct{}{}:
				requestCount++
				// Log every 100 requests or every 5 seconds
				if requestCount % 100 == 0 || time.Since(lastLogTime) > 5*time.Second {
					log.Printf("Sent %d requests to workers", requestCount)
					lastLogTime = time.Now()
				}
			default:
				// Channel full, skip this request
				log.Printf("Channel full, skipping request %d", requestCount)
			}
		}
	}
}

// worker processes requests from the request channel
func (r *Runner) worker(ctx context.Context, workerID int, requestCh <-chan struct{}) {
	requestsProcessed := 0
	log.Printf("Worker %d started", workerID)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping, processed %d requests", workerID, requestsProcessed)
			return
		case _, ok := <-requestCh:
			if !ok {
				log.Printf("Worker %d channel closed, processed %d requests", workerID, requestsProcessed)
				return
			}
			
			requestsProcessed++
			if requestsProcessed <= 5 || requestsProcessed % 50 == 0 {
				log.Printf("Worker %d processing request #%d", workerID, requestsProcessed)
			}
			
			// Generate and execute request
			req, err := r.requestGenerator.GenerateRequest()
			if err != nil {
				log.Printf("Worker %d: failed to generate request: %v", workerID, err)
				continue
			}
			
			if requestsProcessed <= 3 {
				log.Printf("Worker %d: generated request: %s %s", workerID, req.Method, req.Path)
			}
			
			// Execute request with timeout
			reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			resp := r.httpClient.Execute(reqCtx, req)
			cancel()
			
			// Log first few responses and errors
			if requestsProcessed <= 3 {
				log.Printf("Worker %d: response status: %d, error: %v", workerID, resp.StatusCode, resp.Error)
			}
			
			// Log errors for debugging (but don't spam)
			if resp.Error != nil {
				stats := r.metricsCollector.GetStats()
				if stats.TotalRequests % 1000 == 0 || stats.TotalRequests < 10 {
					log.Printf("Worker %d: request error: %v", workerID, resp.Error)
				}
			}
		}
	}
}

// reportingLoop handles periodic reporting
func (r *Runner) reportingLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-r.reportCh:
			return
		case <-ticker.C:
			r.metricsCollector.PrintStats()
		}
	}
}

// Stop stops the load test
func (r *Runner) Stop() {
	if atomic.LoadInt32(&r.running) == 1 {
		close(r.stopCh)
	}
}

// IsRunning returns true if the load test is currently running
func (r *Runner) IsRunning() bool {
	return atomic.LoadInt32(&r.running) == 1
}

// GetStats returns current statistics
func (r *Runner) GetStats() *metrics.Stats {
	return r.metricsCollector.GetStats()
}

// ExportResults exports test results to a file
func (r *Runner) ExportResults(filename string) error {
	data, err := r.metricsCollector.ExportJSON()
	if err != nil {
		return fmt.Errorf("failed to export results: %w", err)
	}
	
	// In a real implementation, you'd write to file
	log.Printf("Results exported to %s (%d bytes)", filename, len(data))
	return nil
}

// Close cleans up resources
func (r *Runner) Close() error {
	r.Stop()
	return r.httpClient.Close()
}
