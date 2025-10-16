# Load Test Client

A high-performance load testing client designed to generate massive loads (1M+ QPS) for the quiz/game leaderboard API. Built for running on large AWS instances with configurable throughput distribution and real-time metrics reporting.

## Features

- **Extreme Performance**: Capable of generating 1M+ QPS on large AWS instances
- **Configurable Load Distribution**: Weighted API endpoint selection
- **Real-time Metrics**: Live statistics every 10 seconds (configurable)
- **Gradual Ramp-up**: Smooth load increase to target QPS
- **Connection Pooling**: Optimized HTTP client with connection reuse
- **Memory Efficient**: Object pooling and minimal allocations
- **Detailed Analytics**: Response time percentiles, error rates, endpoint statistics
- **Flexible Configuration**: YAML-based configuration with CLI overrides

## Quick Start

### Build and Run

```bash
# Build the client
make build

# Run with default configuration (1K QPS for 60s)
make run

# Quick test with CLI arguments
make quick-test

# High load test (100K QPS)
make run-high-load

# Extreme stress test (1M QPS) - USE WITH CAUTION!
make run-stress
```

### Command Line Usage

```bash
# Basic usage
./bin/loadtest --url http://localhost:8080 --duration 60s --qps 1000 --workers 100

# With configuration file
./bin/loadtest --config configs/high-load.yaml

# Override config values
./bin/loadtest --config configs/default.yaml --qps 50000 --workers 1000
```

## Configuration

### Configuration Files

The client supports YAML configuration files with three pre-built profiles:

- **`configs/default.yaml`**: Balanced load test (1K QPS, 60s)
- **`configs/high-load.yaml`**: High throughput test (100K QPS, 5min)
- **`configs/stress-test.yaml`**: Extreme stress test (1M QPS, 2min)

### Configuration Structure

```yaml
target:
  base_url: "http://localhost:8080"
  timeout: "30s"

load:
  duration: "60s"           # Test duration
  workers: 100              # Concurrent workers
  ramp_up_duration: "10s"   # Gradual increase time
  target_qps: 1000          # Target queries per second
  max_concurrency: 1000     # Maximum concurrent requests

apis:
  - name: "create_user"
    method: "POST"
    path: "/api/v1/users"
    weight: 0.1             # 10% of traffic
    body: |
      {
        "user_id": "{{.UserID}}",
        "name": "{{.Name}}",
        "email": "{{.Email}}"
      }
    headers:
      Content-Type: "application/json"

reporting:
  interval: "10s"           # Report every 10 seconds
  output_file: "results.json"
  detailed_stats: true

http:
  max_idle_conns: 1000      # Connection pool size
  max_conns_per_host: 1000  # Connections per host
  idle_conn_timeout: "90s"
```

### Template Variables

The client supports template variables in request bodies and paths:

- `{{.UserID}}`: Generated user ID
- `{{.SessionID}}`: Generated session ID  
- `{{.Name}}`: Random user name
- `{{.Email}}`: Random email address
- `{{.Score}}`: Random game score
- `{{.TopN}}`: Random leaderboard size (5, 10, 20, 50, 100)

## Performance Optimization

### For Maximum Throughput (1M+ QPS)

1. **Use Large AWS Instances**: 
   - c5.24xlarge (96 vCPUs, 192 GB RAM)
   - c6i.32xlarge (128 vCPUs, 256 GB RAM)

2. **System Tuning**:
   ```bash
   # Increase file descriptor limits
   ulimit -n 1000000
   
   # Tune network settings
   echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
   echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
   echo 'net.core.netdev_max_backlog = 65535' >> /etc/sysctl.conf
   sysctl -p
   ```

3. **Configuration Optimizations**:
   - High worker count (2000-5000)
   - Large connection pools (10000+)
   - Disable compression for CPU savings
   - Short timeouts (5-10s)
   - Minimal request/response bodies

4. **Go Runtime Tuning**:
   ```bash
   export GOMAXPROCS=96  # Match CPU count
   export GOGC=100       # Tune GC frequency
   ```

### Memory Usage

The client is designed for minimal memory usage:
- Object pooling for HTTP requests/responses
- Efficient template processing
- Circular buffers for metrics
- Connection reuse

Expected memory usage: ~1-2GB for 1M QPS with 5000 workers.

## Metrics and Reporting

### Real-time Console Output

Every 10 seconds (configurable), the client outputs:

```
=== Load Test Statistics ===
Duration: 30s
Total Requests: 30000
Total Errors: 15
Requests/sec: 1000.50
Error Rate: 0.05%
Avg Response Time: 25.3ms
Min Response Time: 5.2ms
Max Response Time: 150.8ms
P50 Response Time: 22.1ms
P95 Response Time: 45.6ms
P99 Response Time: 89.2ms

=== Endpoint Statistics ===
POST /api/v1/users:
  Requests: 3000
  Errors: 5
  Avg Response: 35.2ms
  P95 Response: 67.8ms
  Status Codes: 201:2995 500:5

GET /api/v1/users/{{.UserID}}:
  Requests: 6000
  Errors: 2
  Avg Response: 18.7ms
  P95 Response: 32.1ms
  Status Codes: 200:5998 404:2
```

### JSON Export

Results are automatically exported to JSON for analysis:

```json
{
  "start_time": "2024-01-15T10:30:00Z",
  "duration": "60s",
  "total_requests": 60000,
  "total_errors": 45,
  "requests_per_second": 1000.75,
  "error_rate": 0.075,
  "avg_response_time": "24.5ms",
  "p95_response_time": "48.2ms",
  "endpoints": {
    "POST /api/v1/users": {
      "request_count": 6000,
      "error_count": 12,
      "avg_response_time": "32.1ms"
    }
  }
}
```

## Testing Scenarios

### 1. Baseline Performance Test
```bash
./bin/loadtest --config configs/default.yaml
```
- 1K QPS for 60 seconds
- Balanced API distribution
- Good for initial performance assessment

### 2. High Load Test
```bash
./bin/loadtest --config configs/high-load.yaml
```
- 100K QPS for 5 minutes
- Read-heavy workload
- Suitable for production load simulation

### 3. Stress Test
```bash
./bin/loadtest --config configs/stress-test.yaml
```
- 1M QPS for 2 minutes
- Maximum system stress
- **WARNING**: Can overwhelm systems

### 4. Custom Endpoint Testing
```bash
# Test Redis backend
make test-redis
```

## AWS Deployment

### Recommended Instance Types

For high-performance load testing:

| Instance Type | vCPUs | Memory | Network | Max QPS* |
|---------------|-------|---------|---------|----------|
| c5.2xlarge    | 8     | 16 GB   | 10 Gbps | ~50K     |
| c5.9xlarge    | 36    | 72 GB   | 10 Gbps | ~200K    |
| c5.18xlarge   | 72    | 144 GB  | 25 Gbps | ~500K    |
| c5.24xlarge   | 96    | 192 GB  | 25 Gbps | ~750K    |
| c6i.32xlarge  | 128   | 256 GB  | 50 Gbps | ~1M+     |

*Estimated maximum QPS depends on target service performance

### Deployment Script

```bash
#!/bin/bash
# AWS deployment script

# Update system
sudo yum update -y

# Install Go 1.23
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone <repository>
cd loadtest-client
make deps
make build-prod

# System tuning
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=65535
ulimit -n 1000000

# Run high load test
./bin/loadtest --config configs/high-load.yaml
```

## Troubleshooting

### Common Issues

1. **"Too many open files" error**
   ```bash
   ulimit -n 1000000
   ```

2. **Low throughput on high QPS**
   - Increase worker count
   - Increase connection pool size
   - Check target service capacity

3. **High memory usage**
   - Reduce worker count
   - Enable compression
   - Tune GOGC value

4. **Network bottlenecks**
   - Use instances with higher network bandwidth
   - Test from multiple instances
   - Monitor network utilization

### Performance Monitoring

```bash
# Monitor during test
htop                    # CPU and memory usage
iftop                   # Network usage  
ss -tuln               # Connection states
./bin/loadtest --help  # Built-in metrics
```

## Development

### Building from Source

```bash
git clone <repository>
cd loadtest-client
make deps
make build
```

### Testing

```bash
make test
go test -race ./...
go test -bench=. ./...
```

### Contributing

1. Fork the repository
2. Create feature branch
3. Add tests
4. Submit pull request

## License

MIT License - see LICENSE file for details.
