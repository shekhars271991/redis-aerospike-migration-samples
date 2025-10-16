# Quiz App - Leaderboard Service

A production-ready Go backend service providing a clean DB service interface for quiz/game leaderboards with Redis backend.

## Features

- **Clean Architecture**: Uses dependency injection with a common DB service interface
- **Redis Backend**: High-performance Redis implementation
- **RESTful API**: HTTP endpoints using Gin framework
- **Production Ready**: Proper error handling, logging, and graceful shutdown
- **Docker Support**: Complete containerization with Docker Compose

## Architecture

### Database Implementation

#### Redis Backend
- **Strings**: Game sessions with TTL (1 hour expiration)
- **Hashes**: User profiles with structured data
- **Sorted Sets**: Real-time leaderboard with automatic sorting

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
- Redis server

### Installation

1. Install dependencies:
```bash
make deps
```

2. Start Redis (using Docker):
```bash
make redis-run
```

3. Run the service:
```bash
make run
```

The service will start on `http://localhost:8080`

### Docker Compose

Start everything with Docker:

```bash
# With Redis backend
docker-compose --profile redis up -d

# Just Redis database
docker-compose up -d redis
```

## Configuration

Configure the service using environment variables:

### Server Configuration
- `SERVER_HOST` - Server host (default: "0.0.0.0")
- `SERVER_PORT` - Server port (default: "8080")

### Database Selection
- `DB_TYPE` - Database type: "redis" (default: "redis")

### Redis Configuration
- `REDIS_ADDR` - Redis address (default: "localhost:6379")
- `REDIS_PASSWORD` - Redis password (default: "")
- `REDIS_DB` - Redis database number (default: 0)

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
├── examples/           # API demonstration scripts
├── docker-compose.yml  # Docker Compose configuration
├── Dockerfile         # Container definition
├── Makefile          # Build and development tasks
└── README.md         # This file
```

### Available Make Commands
- `make build` - Build the application
- `make run` - Run with Redis backend
- `make test` - Run tests
- `make deps` - Install dependencies
- `make redis-run` - Start Redis container
- `make docker-up` - Start all services with Docker
- `make clean` - Clean build artifacts

### Testing

Run the test suite:
```bash
make test
```

Run the API demo:
```bash
./examples/api_demo.sh
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
- Excellent performance for read and write operations

## Monitoring

The service exposes a health check endpoint at `/health` that verifies database connectivity.

## Docker Deployment

### Build and Run with Docker

```bash
# Build image
docker build -t quiz-app .

# Run with Redis
docker run --rm -p 8080:8080 -e DB_TYPE=redis quiz-app
```

### Docker Compose

```bash
# Start with Redis backend
docker-compose --profile redis up -d

# View logs
docker-compose logs -f quiz-redis
```

## API Documentation

### Request/Response Examples

#### Create User
```bash
POST /api/v1/users
Content-Type: application/json

{
  "user_id": "user123",
  "name": "Alice Johnson",
  "email": "alice@example.com"
}
```

Response:
```json
{
  "message": "User created successfully",
  "user_id": "user123"
}
```

#### Start Game
```bash
POST /api/v1/game/start
Content-Type: application/json

{
  "user_id": "user123"
}
```

Response:
```json
{
  "session_id": "session_1234567890_987654",
  "user_id": "user123",
  "message": "Game session started"
}
```

#### Get Leaderboard
```bash
GET /api/v1/leaderboard?top=5
```

Response:
```json
{
  "leaderboard": [
    {
      "user_id": "user123",
      "name": "Alice Johnson",
      "score": 2500
    },
    {
      "user_id": "user456",
      "name": "Bob Smith",
      "score": 2100
    }
  ],
  "count": 2
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.
