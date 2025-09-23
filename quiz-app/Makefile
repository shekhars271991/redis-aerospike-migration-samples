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

# Start Redis using docker-compose
redis-run:
	@echo "Starting Redis using docker-compose..."
	docker-compose up -d redis

# Start Aerospike using docker-compose
aerospike-run:
	@echo "Starting Aerospike using docker-compose..."
	docker-compose up -d aerospike

# Start both databases using Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Stop specific services
redis-stop:
	@echo "Stopping Redis using docker-compose..."
	docker-compose stop redis

aerospike-stop:
	@echo "Stopping Aerospike using docker-compose..."
	docker-compose stop aerospike

# Restart specific services
redis-restart:
	@echo "Restarting Redis using docker-compose..."
	docker-compose restart redis

aerospike-restart:
	@echo "Restarting Aerospike using docker-compose..."
	docker-compose restart aerospike

# Show container status
status:
	@echo "Container Status:"
	docker-compose ps

# View logs
redis-logs:
	docker-compose logs -f redis

aerospike-logs:
	docker-compose logs -f aerospike

logs:
	docker-compose logs -f

# Development setup
dev-setup: deps redis-run
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server with Redis"
	@echo "Run 'make run-aerospike' to start with Aerospike (make sure to run 'make aerospike-run' first)"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go
