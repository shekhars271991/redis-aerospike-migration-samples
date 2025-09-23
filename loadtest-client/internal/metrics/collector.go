package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Collector collects and aggregates metrics from load testing
type Collector struct {
	mu              sync.RWMutex
	startTime       time.Time
	requests        map[string]*EndpointMetrics
	totalRequests   int64
	totalErrors     int64
	totalDuration   int64 // nanoseconds
	responseTimes   []time.Duration
	responseTimeMu  sync.Mutex
	maxResponseTime time.Duration
	minResponseTime time.Duration
}

// EndpointMetrics tracks metrics for a specific endpoint
type EndpointMetrics struct {
	Method           string            `json:"method"`
	Path             string            `json:"path"`
	RequestCount     int64             `json:"request_count"`
	ErrorCount       int64             `json:"error_count"`
	TotalDuration    int64             `json:"total_duration_ns"`
	StatusCodes      map[int]int64     `json:"status_codes"`
	ResponseTimes    []time.Duration   `json:"-"`
	AvgResponseTime  time.Duration     `json:"avg_response_time"`
	MinResponseTime  time.Duration     `json:"min_response_time"`
	MaxResponseTime  time.Duration     `json:"max_response_time"`
	P50ResponseTime  time.Duration     `json:"p50_response_time"`
	P95ResponseTime  time.Duration     `json:"p95_response_time"`
	P99ResponseTime  time.Duration     `json:"p99_response_time"`
}

// Stats represents overall statistics
type Stats struct {
	StartTime         time.Time                    `json:"start_time"`
	Duration          time.Duration                `json:"duration"`
	TotalRequests     int64                        `json:"total_requests"`
	TotalErrors       int64                        `json:"total_errors"`
	RequestsPerSecond float64                      `json:"requests_per_second"`
	ErrorRate         float64                      `json:"error_rate"`
	AvgResponseTime   time.Duration                `json:"avg_response_time"`
	MinResponseTime   time.Duration                `json:"min_response_time"`
	MaxResponseTime   time.Duration                `json:"max_response_time"`
	P50ResponseTime   time.Duration                `json:"p50_response_time"`
	P95ResponseTime   time.Duration                `json:"p95_response_time"`
	P99ResponseTime   time.Duration                `json:"p99_response_time"`
	Endpoints         map[string]*EndpointMetrics  `json:"endpoints"`
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		startTime:       time.Now(),
		requests:        make(map[string]*EndpointMetrics),
		minResponseTime: time.Hour, // Initialize with high value
	}
}

// RecordRequest records a request and its metrics
func (c *Collector) RecordRequest(method, path string, statusCode int, duration time.Duration, err error) {
	key := fmt.Sprintf("%s %s", method, path)
	
	// Update atomic counters
	atomic.AddInt64(&c.totalRequests, 1)
	atomic.AddInt64(&c.totalDuration, int64(duration))
	
	if err != nil {
		atomic.AddInt64(&c.totalErrors, 1)
	}
	
	// Update response time tracking
	c.responseTimeMu.Lock()
	if duration > c.maxResponseTime {
		c.maxResponseTime = duration
	}
	if duration < c.minResponseTime {
		c.minResponseTime = duration
	}
	c.responseTimes = append(c.responseTimes, duration)
	c.responseTimeMu.Unlock()
	
	// Update endpoint-specific metrics
	c.mu.Lock()
	defer c.mu.Unlock()
	
	endpoint, exists := c.requests[key]
	if !exists {
		endpoint = &EndpointMetrics{
			Method:          method,
			Path:            path,
			StatusCodes:     make(map[int]int64),
			ResponseTimes:   make([]time.Duration, 0),
			MinResponseTime: time.Hour,
		}
		c.requests[key] = endpoint
	}
	
	endpoint.RequestCount++
	endpoint.TotalDuration += int64(duration)
	endpoint.ResponseTimes = append(endpoint.ResponseTimes, duration)
	
	if err != nil {
		endpoint.ErrorCount++
	} else {
		endpoint.StatusCodes[statusCode]++
	}
	
	// Update endpoint response time stats
	if duration > endpoint.MaxResponseTime {
		endpoint.MaxResponseTime = duration
	}
	if duration < endpoint.MinResponseTime {
		endpoint.MinResponseTime = duration
	}
}

// GetStats returns current statistics
func (c *Collector) GetStats() *Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	totalRequests := atomic.LoadInt64(&c.totalRequests)
	totalErrors := atomic.LoadInt64(&c.totalErrors)
	totalDuration := atomic.LoadInt64(&c.totalDuration)
	
	duration := time.Since(c.startTime)
	
	stats := &Stats{
		StartTime:     c.startTime,
		Duration:      duration,
		TotalRequests: totalRequests,
		TotalErrors:   totalErrors,
		Endpoints:     make(map[string]*EndpointMetrics),
	}
	
	if totalRequests > 0 {
		stats.RequestsPerSecond = float64(totalRequests) / duration.Seconds()
		stats.ErrorRate = float64(totalErrors) / float64(totalRequests) * 100
		stats.AvgResponseTime = time.Duration(totalDuration / totalRequests)
	}
	
	// Calculate percentiles for overall response times
	c.responseTimeMu.Lock()
	if len(c.responseTimes) > 0 {
		sortedTimes := make([]time.Duration, len(c.responseTimes))
		copy(sortedTimes, c.responseTimes)
		sort.Slice(sortedTimes, func(i, j int) bool {
			return sortedTimes[i] < sortedTimes[j]
		})
		
		stats.MinResponseTime = c.minResponseTime
		stats.MaxResponseTime = c.maxResponseTime
		stats.P50ResponseTime = calculatePercentile(sortedTimes, 50)
		stats.P95ResponseTime = calculatePercentile(sortedTimes, 95)
		stats.P99ResponseTime = calculatePercentile(sortedTimes, 99)
	}
	c.responseTimeMu.Unlock()
	
	// Calculate endpoint-specific stats
	for key, endpoint := range c.requests {
		endpointCopy := &EndpointMetrics{
			Method:        endpoint.Method,
			Path:          endpoint.Path,
			RequestCount:  endpoint.RequestCount,
			ErrorCount:    endpoint.ErrorCount,
			TotalDuration: endpoint.TotalDuration,
			StatusCodes:   make(map[int]int64),
		}
		
		// Copy status codes
		for code, count := range endpoint.StatusCodes {
			endpointCopy.StatusCodes[code] = count
		}
		
		// Calculate endpoint response time stats
		if endpoint.RequestCount > 0 {
			endpointCopy.AvgResponseTime = time.Duration(endpoint.TotalDuration / endpoint.RequestCount)
			endpointCopy.MinResponseTime = endpoint.MinResponseTime
			endpointCopy.MaxResponseTime = endpoint.MaxResponseTime
			
			// Calculate percentiles
			if len(endpoint.ResponseTimes) > 0 {
				sortedTimes := make([]time.Duration, len(endpoint.ResponseTimes))
				copy(sortedTimes, endpoint.ResponseTimes)
				sort.Slice(sortedTimes, func(i, j int) bool {
					return sortedTimes[i] < sortedTimes[j]
				})
				
				endpointCopy.P50ResponseTime = calculatePercentile(sortedTimes, 50)
				endpointCopy.P95ResponseTime = calculatePercentile(sortedTimes, 95)
				endpointCopy.P99ResponseTime = calculatePercentile(sortedTimes, 99)
			}
		}
		
		stats.Endpoints[key] = endpointCopy
	}
	
	return stats
}

// Reset resets all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.startTime = time.Now()
	c.requests = make(map[string]*EndpointMetrics)
	atomic.StoreInt64(&c.totalRequests, 0)
	atomic.StoreInt64(&c.totalErrors, 0)
	atomic.StoreInt64(&c.totalDuration, 0)
	
	c.responseTimeMu.Lock()
	c.responseTimes = c.responseTimes[:0]
	c.maxResponseTime = 0
	c.minResponseTime = time.Hour
	c.responseTimeMu.Unlock()
}

// calculatePercentile calculates the nth percentile from sorted durations
func calculatePercentile(sortedDurations []time.Duration, percentile int) time.Duration {
	if len(sortedDurations) == 0 {
		return 0
	}
	
	index := (percentile * len(sortedDurations)) / 100
	if index >= len(sortedDurations) {
		index = len(sortedDurations) - 1
	}
	
	return sortedDurations[index]
}

// PrintStats prints formatted statistics to console
func (c *Collector) PrintStats() {
	stats := c.GetStats()
	
	fmt.Printf("\n=== Load Test Statistics ===\n")
	fmt.Printf("Duration: %v\n", stats.Duration.Round(time.Second))
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Total Errors: %d\n", stats.TotalErrors)
	fmt.Printf("Requests/sec: %.2f\n", stats.RequestsPerSecond)
	fmt.Printf("Error Rate: %.2f%%\n", stats.ErrorRate)
	fmt.Printf("Avg Response Time: %v\n", stats.AvgResponseTime.Round(time.Microsecond))
	fmt.Printf("Min Response Time: %v\n", stats.MinResponseTime.Round(time.Microsecond))
	fmt.Printf("Max Response Time: %v\n", stats.MaxResponseTime.Round(time.Microsecond))
	fmt.Printf("P50 Response Time: %v\n", stats.P50ResponseTime.Round(time.Microsecond))
	fmt.Printf("P95 Response Time: %v\n", stats.P95ResponseTime.Round(time.Microsecond))
	fmt.Printf("P99 Response Time: %v\n", stats.P99ResponseTime.Round(time.Microsecond))
	
	if len(stats.Endpoints) > 0 {
		fmt.Printf("\n=== Endpoint Statistics ===\n")
		for key, endpoint := range stats.Endpoints {
			fmt.Printf("\n%s:\n", key)
			fmt.Printf("  Requests: %d\n", endpoint.RequestCount)
			fmt.Printf("  Errors: %d\n", endpoint.ErrorCount)
			fmt.Printf("  Avg Response: %v\n", endpoint.AvgResponseTime.Round(time.Microsecond))
			fmt.Printf("  P95 Response: %v\n", endpoint.P95ResponseTime.Round(time.Microsecond))
			
			if len(endpoint.StatusCodes) > 0 {
				fmt.Printf("  Status Codes: ")
				for code, count := range endpoint.StatusCodes {
					fmt.Printf("%d:%d ", code, count)
				}
				fmt.Printf("\n")
			}
		}
	}
	fmt.Printf("============================\n\n")
}

// ExportJSON exports statistics as JSON
func (c *Collector) ExportJSON() ([]byte, error) {
	stats := c.GetStats()
	return json.MarshalIndent(stats, "", "  ")
}
