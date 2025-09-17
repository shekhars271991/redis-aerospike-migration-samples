.PHONY: build run test clean deps redis-run aerospike-run docker-up docker-down

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run with Redis backend (default)
run:
	DB_TYPE=redis go run cmd/server/main.go

# Run with Aerospike backend
run-aerospike:
	DB_TYPE=aerospike go run cmd/server/main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Start Redis using Docker
redis-run:
	docker run --name redis-quiz -p 6379:6379 -d redis:7-alpine

# Start Aerospike using Docker
aerospike-run:
	docker run --name aerospike-quiz -p 3000:3000 -d aerospike/aerospike-server:8.0.0.0

# Start both databases using Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Development setup
dev-setup: deps redis-run
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server with Redis"
	@echo "Run 'make run-aerospike' to start with Aerospike (make sure to run 'make aerospike-run' first)"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go
