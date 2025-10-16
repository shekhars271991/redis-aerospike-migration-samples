.PHONY: build run test clean deps install

# Build the load test client
build:
	go build -o bin/loadtest cmd/loadtest/main.go

# Build optimized for production/high-performance
build-prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/loadtest cmd/loadtest/main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run with default configuration
run:
	./bin/loadtest --config configs/default.yaml

# Run high load test (100K QPS)
run-high-load:
	./bin/loadtest --config configs/high-load.yaml

# Run stress test (1M QPS) - BE CAREFUL!
run-stress:
	./bin/loadtest --config configs/stress-test.yaml

# Quick test with command line args
quick-test:
	./bin/loadtest --url http://localhost:8080 --duration 30s --qps 1000 --workers 50

# Test against Redis backend
test-redis:
	./bin/loadtest --url http://localhost:8080 --duration 60s --qps 5000 --workers 200

# Generate sample configuration
generate-config:
	./bin/loadtest --config /dev/null --help

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Install the binary to GOPATH/bin
install: build-prod
	cp bin/loadtest $(GOPATH)/bin/

# Development setup
dev-setup: deps build
	@echo "Load test client ready!"
	@echo "Run 'make run' to start with default config"
	@echo "Run 'make run-high-load' for high throughput testing"

# Docker build
docker-build:
	docker build -t loadtest-client .

# Performance profiling build
build-profile:
	go build -race -o bin/loadtest-profile cmd/loadtest/main.go

# Benchmark test
benchmark:
	go test -bench=. -benchmem ./...
