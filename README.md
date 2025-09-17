# Quiz/Game Leaderboard Service

A Go backend service that provides a clean DB service interface for a quiz/game leaderboard with two interchangeable backends: Redis and Aerospike.

## Features

- **Clean Architecture**: Uses dependency injection with a common DB service interface
- **Dual Backend Support**: Redis and Aerospike implementations
- **RESTful API**: HTTP endpoints using Gin framework
- **Production Ready**: Proper error handling, logging, and graceful shutdown
- **Docker Support**: Complete containerization with Docker Compose

## Architecture

### Database Implementations

#### Redis Backend
- **Strings**: Game sessions with TTL
- **Hashes**: User profiles with structured data
- **Sorted Sets**: Real-time leaderboard with automatic sorting

#### Aerospike Backend
- **User Records**: Single record per user in `users` set with bins: `{name, email, games_played, last_score}`
- **Materialized Leaderboard**: Maintains a `leaderboard:top` record with list of top N users
- **Game Sessions**: Temporary session records with TTL

### API Endpoints

#### User Management
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/{id}` - Get user by ID

#### Game Management
- `POST /api/v1/game/start` - Start a new game session
- `POST /api/v1/game/finish` - Finish game and update score
- `GET /api/v1/game/session/{id}` - Get game session details

#### Leaderboard
- `GET /api/v1/leaderboard?top=N` - Get top N users (default: 10, max: 100)

#### Health Check
- `GET /health` - Service health check

## Quick Start

### Prerequisites
- Go 1.23+
- Docker and Docker Compose (optional)
- Redis or Aerospike server

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd redis-migration
```

2. Install dependencies:
```bash
make deps
```

3. Start Redis (using Docker):
```bash
make redis-run
```

4. Run the service:
```bash
make run
```

The service will start on `http://localhost:8080`

### Using Aerospike

1. Start Aerospike:
```bash
make aerospike-run
```

2. Run with Aerospike backend:
```bash
make run-aerospike
```

### Docker Compose

Start everything with Docker:

```bash
# With Redis backend
docker-compose --profile redis up -d

# With Aerospike backend  
docker-compose --profile aerospike up -d

# Just the databases
docker-compose up -d redis aerospike
```

## Configuration

Configure the service using environment variables:

### Server Configuration
- `SERVER_HOST` - Server host (default: "0.0.0.0")
- `SERVER_PORT` - Server port (default: "8080")

### Database Selection
- `DB_TYPE` - Database type: "redis" or "aerospike" (default: "redis")

### Redis Configuration
- `REDIS_ADDR` - Redis address (default: "localhost:6379")
- `REDIS_PASSWORD` - Redis password (default: "")
- `REDIS_DB` - Redis database number (default: 0)

### Aerospike Configuration
- `AEROSPIKE_HOSTS` - Comma-separated list of hosts (default: "localhost")
- `AEROSPIKE_PORT` - Aerospike port (default: 3000)

## Usage Examples

### Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "name": "John Doe", 
    "email": "john@example.com"
  }'
```

### Start a Game
```bash
curl -X POST http://localhost:8080/api/v1/game/start \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1"}'
```

### Finish a Game
```bash
curl -X POST http://localhost:8080/api/v1/game/finish \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-uuid-here",
    "score": 1500
  }'
```

### Get Leaderboard
```bash
curl "http://localhost:8080/api/v1/leaderboard?top=5"
```

### Get User Details
```bash
curl "http://localhost:8080/api/v1/users/user1"
```

## Development

### Project Structure
```
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   ├── router/         # HTTP routing
│   └── service/        # Database service implementations
├── docker-compose.yml  # Docker Compose configuration
├── Dockerfile         # Container definition
├── Makefile          # Build and development tasks
└── README.md         # This file
```

### Available Make Commands
- `make build` - Build the application
- `make run` - Run with Redis backend
- `make run-aerospike` - Run with Aerospike backend
- `make test` - Run tests
- `make deps` - Install dependencies
- `make redis-run` - Start Redis container
- `make aerospike-run` - Start Aerospike container
- `make docker-up` - Start all services with Docker
- `make clean` - Clean build artifacts

### Testing

Run the test suite:
```bash
make test
```

### Building for Production

```bash
make build-prod
```

## Performance Considerations

### Redis Backend
- Uses pipelining for atomic operations
- Sorted sets provide O(log N) insertion and range queries
- Session data has automatic TTL expiration

### Aerospike Backend
- Materialized leaderboard for fast queries
- Single-record transactions for consistency
- Configurable TTL for session management

## Monitoring

The service exposes a health check endpoint at `/health` that verifies database connectivity.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.
