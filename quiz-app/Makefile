.PHONY: build run test clean deps redis-run docker-up docker-down

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run with Redis backend
run:
	DB_TYPE=redis go run cmd/server/main.go

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

# Start databases using Docker Compose
docker-up:
	docker-compose up -d redis

# Stop Docker containers
docker-down:
	docker-compose down

# Stop Redis service
redis-stop:
	@echo "Stopping Redis using docker-compose..."
	docker-compose stop redis

# Restart Redis service
redis-restart:
	@echo "Restarting Redis using docker-compose..."
	docker-compose restart redis

# Show container status
status:
	@echo "Container Status:"
	docker-compose ps

# View logs
redis-logs:
	docker-compose logs -f redis

logs:
	docker-compose logs -f

# Development setup
dev-setup: deps redis-run
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server with Redis"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go
