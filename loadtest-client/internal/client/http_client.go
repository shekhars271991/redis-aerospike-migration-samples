package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"loadtest-client/internal/config"
	"loadtest-client/internal/metrics"
)

// HTTPClient is a high-performance HTTP client for load testing
type HTTPClient struct {
	client   *http.Client
	baseURL  string
	metrics  *metrics.Collector
	reqPool  sync.Pool
	respPool sync.Pool
}

// NewHTTPClient creates a new HTTP client optimized for high throughput
func NewHTTPClient(cfg *config.Config, metricsCollector *metrics.Collector) (*HTTPClient, error) {
	timeout, err := cfg.GetTargetTimeout()
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	idleTimeout, err := cfg.GetIdleConnTimeout()
	if err != nil {
		return nil, fmt.Errorf("invalid idle timeout: %w", err)
	}

	transport := &http.Transport{
		MaxIdleConns:        cfg.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.HTTP.MaxConnsPerHost,
		IdleConnTimeout:     idleTimeout,
		DisableCompression:  cfg.HTTP.DisableCompression,
		DisableKeepAlives:   cfg.HTTP.DisableKeepAlives,
		// Optimize for high throughput
		WriteBufferSize:     32 * 1024, // 32KB
		ReadBufferSize:      32 * 1024, // 32KB
		ForceAttemptHTTP2:   true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	httpClient := &HTTPClient{
		client:  client,
		baseURL: cfg.Target.BaseURL,
		metrics: metricsCollector,
	}

	// Initialize object pools for memory efficiency
	httpClient.reqPool = sync.Pool{
		New: func() interface{} {
			return &http.Request{}
		},
	}

	httpClient.respPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024) // Pre-allocate 1KB buffer
		},
	}

	return httpClient, nil
}

// Request represents an HTTP request to be made
type Request struct {
	Method  string
	Path    string
	Body    []byte
	Headers map[string]string
}

// Response represents the HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Duration   time.Duration
	Error      error
}

// Execute performs an HTTP request and returns the response
func (c *HTTPClient) Execute(ctx context.Context, req *Request) *Response {
	start := time.Now()
	
	// Create full URL
	url := c.baseURL + req.Path
	
	// Create HTTP request
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return &Response{
			Error:    err,
			Duration: time.Since(start),
		}
	}
	
	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	
	// Set default headers for better performance
	httpReq.Header.Set("User-Agent", "LoadTestClient/1.0")
	httpReq.Header.Set("Accept", "*/*")
	
	// Execute request
	resp, err := c.client.Do(httpReq)
	duration := time.Since(start)
	
	if err != nil {
		c.metrics.RecordRequest(req.Method, req.Path, 0, duration, err)
		return &Response{
			Error:    err,
			Duration: duration,
		}
	}
	
	defer resp.Body.Close()
	
	// Read response body efficiently
	bodyBytes := c.respPool.Get().([]byte)
	bodyBytes = bodyBytes[:0] // Reset slice but keep capacity
	
	// Read body in chunks for memory efficiency
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			c.respPool.Put(bodyBytes)
			c.metrics.RecordRequest(req.Method, req.Path, resp.StatusCode, duration, err)
			return &Response{
				StatusCode: resp.StatusCode,
				Error:      err,
				Duration:   duration,
			}
		}
	}
	
	// Create response
	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       make([]byte, len(bodyBytes)),
		Duration:   duration,
	}
	
	copy(response.Body, bodyBytes)
	c.respPool.Put(bodyBytes)
	
	// Record metrics
	c.metrics.RecordRequest(req.Method, req.Path, resp.StatusCode, duration, nil)
	
	return response
}

// ExecuteBatch executes multiple requests concurrently
func (c *HTTPClient) ExecuteBatch(ctx context.Context, requests []*Request) []*Response {
	responses := make([]*Response, len(requests))
	var wg sync.WaitGroup
	
	for i, req := range requests {
		wg.Add(1)
		go func(idx int, request *Request) {
			defer wg.Done()
			responses[idx] = c.Execute(ctx, request)
		}(i, req)
	}
	
	wg.Wait()
	return responses
}

// Close closes the HTTP client and cleans up resources
func (c *HTTPClient) Close() error {
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// GetStats returns client statistics
func (c *HTTPClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"base_url": c.baseURL,
	}
}
