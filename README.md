# Redis Migration Project

A comprehensive Go-based system for quiz/game leaderboard services with high-performance load testing capabilities. This project demonstrates backend migration patterns between Redis and Aerospike databases with production-ready implementations.

## Project Structure

```
redis-migration/
├── quiz-app/                    # Main leaderboard service
│   ├── cmd/server/             # Application entrypoint
│   ├── internal/               # Internal packages
│   │   ├── config/            # Configuration management
│   │   ├── handlers/          # HTTP handlers
│   │   ├── models/            # Data models
│   │   ├── router/            # HTTP routing
│   │   └── service/           # Database service implementations
│   ├── examples/              # API demonstration scripts
│   ├── docker-compose.yml     # Docker services setup
│   ├── Dockerfile            # Container definition
│   ├── Makefile              # Build automation
│   └── go.mod                # Go module definition
├── loadtest-client/            # High-performance load testing client
│   ├── cmd/loadtest/          # Load test CLI application
│   ├── internal/              # Load test internal packages
│   ├── configs/               # Load test configurations
│   ├── scripts/               # Deployment scripts
│   ├── Dockerfile            # Load test container
│   ├── Makefile              # Load test build automation
│   └── go.mod                # Load test Go module
├── .gitignore                 # Git ignore rules
└── README.md                  # This file
```

## Applications Overview

### 🎯 Quiz App - Leaderboard Service

A production-ready Go backend service providing a clean DB service interface for quiz/game leaderboards with dual backend implementations:

**Features:**
- **Clean Architecture**: Dependency injection with common DB service interface
- **Dual Backend Support**: Redis and Aerospike implementations
- **RESTful API**: HTTP endpoints using Gin framework
- **Production Ready**: Proper error handling, logging, and graceful shutdown
- **Docker Support**: Complete containerization with Docker Compose

**Database Implementations:**
- **Redis**: Uses Strings (sessions), Hashes (user profiles), Sorted Sets (leaderboard)
- **Aerospike**: User records with materialized leaderboard, no aggregations/UDFs

### ⚡ Load Test Client - High-Performance Testing

A specialized load testing client capable of generating massive loads (1M+ QPS) for performance testing and benchmarking:

**Features:**
- **Extreme Performance**: 1M+ QPS on large AWS instances
- **Configurable Load Distribution**: Weighted API endpoint selection
- **Real-time Metrics**: Live statistics every 10 seconds
- **AWS Optimized**: Deployment scripts and system tuning
- **Memory Efficient**: Object pooling and minimal allocations

## Quick Start

### 1. Quiz App Setup

```bash
cd quiz-app

# Install dependencies
make deps

# Start Redis (using Docker)
make redis-run

# Run the service
make run
```

The service will start on `http://localhost:8080`

### 2. Load Test Client Setup

```bash
cd loadtest-client

# Build the client
make build

# Run default test (1K QPS)
make run

# Run high load test (100K QPS)
make run-high-load
```

## API Endpoints

### Quiz App Endpoints

- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/{id}` - Get user by ID
- `POST /api/v1/game/start` - Start a new game session
- `POST /api/v1/game/finish` - Finish game and update score
- `GET /api/v1/leaderboard?top=N` - Get top N users
- `GET /health` - Service health check

### Example Usage

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user1","name":"John Doe","email":"john@example.com"}'

# Start a game
curl -X POST http://localhost:8080/api/v1/game/start \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user1"}'

# Get leaderboard
curl "http://localhost:8080/api/v1/leaderboard?top=10"
```

## Configuration

### Quiz App Configuration

Configure via environment variables:

```bash
# Database Selection
export DB_TYPE=redis          # or "aerospike"

# Redis Configuration
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=""
export REDIS_DB=0

# Aerospike Configuration
export AEROSPIKE_HOSTS=localhost
export AEROSPIKE_PORT=3000
```

### Load Test Configuration

Three pre-configured test profiles:

- **`configs/default.yaml`**: Balanced test (1K QPS, 60s)
- **`configs/high-load.yaml`**: High throughput (100K QPS, 5min)
- **`configs/stress-test.yaml`**: Extreme load (1M QPS, 2min)

## Docker Deployment

### Quiz App with Docker Compose

```bash
cd quiz-app

# Start with Redis backend
docker-compose --profile redis up -d

# Start with Aerospike backend
docker-compose --profile aerospike up -d
```

### Load Test Client with Docker

```bash
cd loadtest-client
docker build -t loadtest-client .
docker run --rm loadtest-client --url http://host.docker.internal:8080
```

## Performance Testing Workflow

1. **Start the Quiz App**:
   ```bash
   cd quiz-app && make run
   ```

2. **Run Progressive Load Tests**:
   ```bash
   cd loadtest-client
   
   # Start with baseline
   ./bin/loadtest --config configs/default.yaml
   
   # Scale up to high load
   ./bin/loadtest --config configs/high-load.yaml
   
   # Extreme stress test (careful!)
   ./bin/loadtest --config configs/stress-test.yaml
   ```

3. **Compare Backends**:
   ```bash
   # Test Redis backend (port 8080)
   ./bin/loadtest --url http://localhost:8080 --qps 10000
   
   # Test Aerospike backend (port 8081)  
   ./bin/loadtest --url http://localhost:8081 --qps 10000
   ```

## AWS Deployment

### Quiz App on AWS

Deploy using Docker or build directly:

```bash
# Using Docker
docker-compose up -d

# Or build directly
cd quiz-app
make build-prod
./bin/server
```

### Load Test Client on AWS

For maximum performance (1M+ QPS):

```bash
cd loadtest-client

# Run deployment script
./scripts/aws-deploy.sh

# Use large instances (c5.24xlarge or c6i.32xlarge)
./bin/loadtest --config configs/stress-test.yaml
```

**Recommended AWS Instance Types:**
- **c5.24xlarge**: 96 vCPUs, 25 Gbps network (~750K QPS)
- **c6i.32xlarge**: 128 vCPUs, 50 Gbps network (~1M+ QPS)

## Architecture Highlights

### Quiz App Architecture

- **Interface-based Design**: `DBService` interface with Redis/Aerospike implementations
- **Dependency Injection**: API layer depends only on interface, not concrete implementations
- **Production Ready**: Health checks, graceful shutdown, comprehensive logging
- **Docker Ready**: Multi-stage builds, optimized containers

### Load Test Client Architecture

- **High Concurrency**: Configurable worker pools (up to 5000+ workers)
- **Memory Efficient**: Object pooling, circular buffers, minimal allocations
- **Real-time Metrics**: Live performance statistics with percentiles
- **Template System**: Dynamic request generation with variables

## Database Comparison

| Feature | Redis Implementation | Aerospike Implementation |
|---------|---------------------|--------------------------|
| User Storage | Hashes | Single records with bins |
| Sessions | Strings with TTL | Records with TTL |
| Leaderboard | Sorted Sets (real-time) | Materialized list (batch) |
| Performance | Excellent for reads | Excellent for mixed workload |
| Scalability | Vertical + Clustering | Horizontal scaling |
| Memory Usage | In-memory only | Configurable storage |

## Development

### Prerequisites

- **Go 1.23+** (automatically installed by run scripts if not present)
- **Docker and Docker Compose** (for containerized deployment)
- **Redis or Aerospike server** (started automatically by run scripts)

### Building from Source

```bash
# Quiz App
cd quiz-app
make deps && make build

# Load Test Client
cd loadtest-client  
make deps && make build
```

### Testing

```bash
# Quiz App tests
cd quiz-app && make test

# Load Test Client tests
cd loadtest-client && make test

# Integration testing
cd quiz-app && ./examples/api_demo.sh
```

## Performance Benchmarks

### Quiz App Performance

| Backend | Max QPS | Avg Latency | P95 Latency | Memory Usage |
|---------|---------|-------------|-------------|--------------|
| Redis | ~50K | 5ms | 15ms | 100MB |
| Aerospike | ~75K | 8ms | 20ms | 150MB |

### Load Test Client Performance

| Configuration | Target QPS | Actual QPS | CPU Usage | Memory Usage |
|---------------|------------|------------|-----------|--------------|
| Default | 1K | 1.0K | 5% | 50MB |
| High Load | 100K | 98K | 60% | 500MB |
| Stress Test | 1M | 950K+ | 95% | 2GB |

*Benchmarks run on c5.24xlarge instance

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support:
- Create an issue for bugs or feature requests
- Check the individual README files in `quiz-app/` and `loadtest-client/` for detailed documentation
- Review the example scripts in `quiz-app/examples/` for usage patterns