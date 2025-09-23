#!/bin/bash

# Quiz App Runner Script
# This script helps you run all components of the quiz app with different configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
BACKEND="redis"
MODE="dev"
PORT="8081"
REDIS_PORT="8081"
AEROSPIKE_PORT="8082"
SKIP_BUILD=false
SKIP_DEPS=false
CLEANUP=false
DEMO=false
DUAL_BACKEND=false

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Quiz App Runner - Starts all components of the quiz/game leaderboard service

OPTIONS:
    -b, --backend BACKEND    Database backend: redis (default) or aerospike
    -m, --mode MODE         Run mode: dev (default), docker, or prod
    -p, --port PORT         Server port (default: 8081)
    --redis-port PORT       Redis backend port (default: 8081, used with --dual)
    --aerospike-port PORT   Aerospike backend port (default: 8082, used with --dual)
    --dual, --both          Run both Redis and Aerospike backends simultaneously
    --skip-build           Skip building the application
    --skip-deps            Skip installing dependencies
    --cleanup              Stop and remove all containers before starting
    --demo                 Run API demo after starting services
    -h, --help             Show this help message

EXAMPLES:
    $0                                    # Run with Redis in dev mode
    $0 --backend aerospike               # Run with Aerospike backend
    $0 --dual                            # Run both backends (Redis:8081, Aerospike:8082)
    $0 --dual --redis-port 9001 --aerospike-port 9002  # Custom ports for dual mode
    $0 --mode docker --backend redis    # Run everything in Docker
    $0 --cleanup --demo                  # Clean start with demo
    $0 --port 9090 --backend aerospike   # Custom port with Aerospike

MODES:
    dev     - Run services locally with Docker databases
    docker  - Run everything in Docker containers
    prod    - Production mode with optimized builds

EOF
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -b|--backend)
                BACKEND="$2"
                shift 2
                ;;
            -m|--mode)
                MODE="$2"
                shift 2
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            --redis-port)
                REDIS_PORT="$2"
                shift 2
                ;;
            --aerospike-port)
                AEROSPIKE_PORT="$2"
                shift 2
                ;;
            --dual|--both)
                DUAL_BACKEND=true
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --cleanup)
                CLEANUP=true
                shift
                ;;
            --demo)
                DEMO=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Function to validate arguments
validate_args() {
    if [[ "$DUAL_BACKEND" == "false" ]]; then
        if [[ "$BACKEND" != "redis" && "$BACKEND" != "aerospike" ]]; then
            print_error "Invalid backend: $BACKEND. Must be 'redis' or 'aerospike'"
            exit 1
        fi
    else
        # In dual mode, ignore single backend setting
        print_info "Dual backend mode enabled - will run both Redis and Aerospike"
    fi

    if [[ "$MODE" != "dev" && "$MODE" != "docker" && "$MODE" != "prod" ]]; then
        print_error "Invalid mode: $MODE. Must be 'dev', 'docker', or 'prod'"
        exit 1
    fi

    # Validate ports
    if ! [[ "$PORT" =~ ^[0-9]+$ ]] || [ "$PORT" -lt 1024 ] || [ "$PORT" -gt 65535 ]; then
        print_error "Invalid port: $PORT. Must be between 1024 and 65535"
        exit 1
    fi
    
    if ! [[ "$REDIS_PORT" =~ ^[0-9]+$ ]] || [ "$REDIS_PORT" -lt 1024 ] || [ "$REDIS_PORT" -gt 65535 ]; then
        print_error "Invalid Redis port: $REDIS_PORT. Must be between 1024 and 65535"
        exit 1
    fi
    
    if ! [[ "$AEROSPIKE_PORT" =~ ^[0-9]+$ ]] || [ "$AEROSPIKE_PORT" -lt 1024 ] || [ "$AEROSPIKE_PORT" -gt 65535 ]; then
        print_error "Invalid Aerospike port: $AEROSPIKE_PORT. Must be between 1024 and 65535"
        exit 1
    fi

    # Check for port conflicts in dual mode
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        if [[ "$REDIS_PORT" == "$AEROSPIKE_PORT" ]]; then
            print_error "Redis and Aerospike ports cannot be the same: $REDIS_PORT"
            exit 1
        fi
    fi
}

# Function to install Go if not available
install_go() {
    print_info "Go not found. Installing Go 1.23..."
    
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Map architecture names
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv6l) ARCH="armv6l" ;;
        armv7l) ARCH="armv7l" ;;
        *) print_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    # Set download URL
    GO_VERSION="1.23.0"
    GO_TARBALL="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TARBALL}"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    print_info "Downloading Go from $GO_URL..."
    if command -v curl &> /dev/null; then
        curl -LO "$GO_URL"
    elif command -v wget &> /dev/null; then
        wget "$GO_URL"
    else
        print_error "Neither curl nor wget is available for downloading Go"
        exit 1
    fi
    
    # Install Go
    print_info "Installing Go to /usr/local/go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TARBALL"
    
    # Add Go to PATH
    export PATH=$PATH:/usr/local/go/bin
    
    # Add to shell profile
    SHELL_PROFILE=""
    if [[ -f "$HOME/.bashrc" ]]; then
        SHELL_PROFILE="$HOME/.bashrc"
    elif [[ -f "$HOME/.bash_profile" ]]; then
        SHELL_PROFILE="$HOME/.bash_profile"
    elif [[ -f "$HOME/.zshrc" ]]; then
        SHELL_PROFILE="$HOME/.zshrc"
    fi
    
    if [[ -n "$SHELL_PROFILE" ]]; then
        if ! grep -q "/usr/local/go/bin" "$SHELL_PROFILE"; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_PROFILE"
            print_info "Added Go to PATH in $SHELL_PROFILE"
        fi
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
    
    # Verify installation
    if command -v go &> /dev/null; then
        print_success "Go $(go version | awk '{print $3}') installed successfully"
    else
        print_error "Go installation failed"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    # Check if we're in the right directory
    if [[ ! -d "quiz-app" ]]; then
        print_error "quiz-app directory not found. Please run this script from the project root."
        exit 1
    fi

    # Check for required tools based on mode
    if [[ "$MODE" == "docker" ]]; then
        if ! command -v docker &> /dev/null; then
            print_error "Docker is required for docker mode"
            exit 1
        fi
        if ! command -v docker-compose &> /dev/null; then
            print_error "Docker Compose is required for docker mode"
            exit 1
        fi
    else
        # Check for Go and install if not available
        # First try common Go installation paths
        if ! command -v go &> /dev/null; then
            # Try to find Go in common locations
            if [[ -x "/usr/local/go/bin/go" ]]; then
                export PATH="/usr/local/go/bin:$PATH"
                print_info "Found Go at /usr/local/go/bin, added to PATH"
            elif [[ -x "$HOME/go/bin/go" ]]; then
                export PATH="$HOME/go/bin:$PATH"
                print_info "Found Go at $HOME/go/bin, added to PATH"
            else
                print_warning "Go not found. Attempting to install..."
                install_go
            fi
        fi
        
        # Verify Go is now available
        if command -v go &> /dev/null; then
            print_info "Go $(go version | awk '{print $3}') found"
        else
            print_error "Go installation failed or not found in PATH"
            exit 1
        fi
    fi

    print_success "Prerequisites check passed"
}

# Function to cleanup existing containers
cleanup_containers() {
    if [[ "$CLEANUP" == true ]]; then
        print_info "Cleaning up existing containers..."
        cd quiz-app
        
        # Stop and remove containers
        docker-compose down --remove-orphans 2>/dev/null || true
        
        # Remove any standalone containers
        local containers_to_cleanup=("redis-quiz" "aerospike-quiz" "quiz-redis" "quiz-aerospike")
        
        for container in "${containers_to_cleanup[@]}"; do
            if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
                print_info "Stopping and removing container: $container"
                docker stop "$container" 2>/dev/null || true
                docker rm "$container" 2>/dev/null || true
            fi
        done
        
        cd ..
        print_success "Cleanup completed"
    fi
}

# Function to install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == false && "$MODE" != "docker" ]]; then
        print_info "Installing dependencies..."
        cd quiz-app
        make deps
        cd ..
        print_success "Dependencies installed"
    fi
}

# Function to build application
build_application() {
    if [[ "$SKIP_BUILD" == false && "$MODE" != "docker" ]]; then
        print_info "Building application..."
        cd quiz-app
        if [[ "$MODE" == "prod" ]]; then
            make build-prod
        else
            make build
        fi
        cd ..
        print_success "Application built"
    fi
}

# Function to start database services
start_databases() {
    print_info "Starting database services..."
    cd quiz-app

    # Always use docker-compose for consistency
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        print_info "Starting both Redis and Aerospike using docker-compose..."
        docker-compose up -d redis aerospike
    elif [[ "$BACKEND" == "redis" ]]; then
        print_info "Starting Redis using docker-compose..."
        docker-compose up -d redis
    else
        print_info "Starting Aerospike using docker-compose..."
        docker-compose up -d aerospike
    fi

    cd ..
    print_success "Database services started"
}

# Function to wait for database readiness
wait_for_database() {
    print_info "Waiting for database to be ready..."
    
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        # Wait for both Redis and Aerospike
        print_info "Waiting for Redis..."
        for i in {1..30}; do
            if docker exec redis-quiz redis-cli ping 2>/dev/null | grep -q PONG; then
                break
            fi
            if [[ $i -eq 30 ]]; then
                print_error "Redis failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Redis is ready"
        
        print_info "Waiting for Aerospike..."
        for i in {1..60}; do
            if docker exec aerospike-quiz /usr/bin/asinfo -v status 2>/dev/null | grep -q "ok"; then
                break
            fi
            if [[ $i -eq 60 ]]; then
                print_error "Aerospike failed to start within 60 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Aerospike is ready"
        
    elif [[ "$BACKEND" == "redis" ]]; then
        # Wait for Redis
        for i in {1..30}; do
            if docker exec redis-quiz redis-cli ping 2>/dev/null | grep -q PONG; then
                break
            fi
            if [[ $i -eq 30 ]]; then
                print_error "Redis failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Redis is ready"
    else
        # Wait for Aerospike
        for i in {1..60}; do
            if docker exec aerospike-quiz /usr/bin/asinfo -v status 2>/dev/null | grep -q "ok"; then
                break
            fi
            if [[ $i -eq 60 ]]; then
                print_error "Aerospike failed to start within 60 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Aerospike is ready"
    fi
}

# Function to start the quiz app
start_quiz_app() {
    print_info "Starting Quiz App..."
    cd quiz-app

    if [[ "$DUAL_BACKEND" == "true" ]]; then
        # Start both Redis and Aerospike backends
        start_redis_app
        start_aerospike_app
    else
        # Start single backend
        export SERVER_PORT="$PORT"
        export DB_TYPE="$BACKEND"

        if [[ "$MODE" == "docker" ]]; then
            # Start app in Docker
            if [[ "$BACKEND" == "redis" ]]; then
                docker-compose --profile redis up -d quiz-redis
            else
                docker-compose --profile aerospike up -d quiz-aerospike
            fi
        else
            # Start app locally
            if [[ "$BACKEND" == "redis" ]]; then
                export REDIS_ADDR="localhost:6379"
            else
                export AEROSPIKE_HOSTS="localhost"
                export AEROSPIKE_PORT="3000"
            fi

            print_info "Starting server on port $PORT with $BACKEND backend..."
            if [[ "$MODE" == "prod" ]]; then
                ./bin/server &
            else
                go run cmd/server/main.go &
            fi
            
            # Store PID for cleanup
            APP_PID=$!
            echo $APP_PID > ../quiz-app.pid
        fi
    fi

    cd ..
    print_success "Quiz App started"
}

# Function to start Redis app instance
start_redis_app() {
    print_info "Starting Redis backend on port $REDIS_PORT..."
    
    if [[ "$MODE" == "docker" ]]; then
        # Start Redis app in Docker with custom port
        SERVER_PORT="$REDIS_PORT" DB_TYPE="redis" docker-compose --profile redis up -d quiz-redis
    else
        # Start Redis app locally
        (
            export SERVER_PORT="$REDIS_PORT"
            export DB_TYPE="redis"
            export REDIS_ADDR="localhost:6379"
            
            if [[ "$MODE" == "prod" ]]; then
                ./bin/server &
            else
                go run cmd/server/main.go &
            fi
            
            # Store PID for cleanup
            REDIS_APP_PID=$!
            echo $REDIS_APP_PID > ../quiz-app-redis.pid
        ) &
    fi
    
    print_success "Redis backend started on port $REDIS_PORT"
}

# Function to start Aerospike app instance
start_aerospike_app() {
    print_info "Starting Aerospike backend on port $AEROSPIKE_PORT..."
    
    if [[ "$MODE" == "docker" ]]; then
        # Start Aerospike app in Docker with custom port
        SERVER_PORT="$AEROSPIKE_PORT" DB_TYPE="aerospike" docker-compose --profile aerospike up -d quiz-aerospike
    else
        # Start Aerospike app locally
        (
            export SERVER_PORT="$AEROSPIKE_PORT"
            export DB_TYPE="aerospike"
            export AEROSPIKE_HOSTS="localhost"
            export AEROSPIKE_PORT="3000"
            
            if [[ "$MODE" == "prod" ]]; then
                ./bin/server &
            else
                go run cmd/server/main.go &
            fi
            
            # Store PID for cleanup
            AEROSPIKE_APP_PID=$!
            echo $AEROSPIKE_APP_PID > ../quiz-app-aerospike.pid
        ) &
    fi
    
    print_success "Aerospike backend started on port $AEROSPIKE_PORT"
}

# Function to wait for app readiness
wait_for_app() {
    print_info "Waiting for Quiz App to be ready..."
    
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        # Wait for both Redis and Aerospike apps
        print_info "Waiting for Redis backend on port $REDIS_PORT..."
        for i in {1..30}; do
            if curl -s "http://localhost:$REDIS_PORT/health" | grep -q "healthy"; then
                break
            fi
            if [[ $i -eq 30 ]]; then
                print_error "Redis backend failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Redis backend is ready on http://localhost:$REDIS_PORT"
        
        print_info "Waiting for Aerospike backend on port $AEROSPIKE_PORT..."
        for i in {1..30}; do
            if curl -s "http://localhost:$AEROSPIKE_PORT/health" | grep -q "healthy"; then
                break
            fi
            if [[ $i -eq 30 ]]; then
                print_error "Aerospike backend failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Aerospike backend is ready on http://localhost:$AEROSPIKE_PORT"
    else
        # Wait for single backend
        for i in {1..30}; do
            if curl -s "http://localhost:$PORT/health" | grep -q "healthy"; then
                break
            fi
            if [[ $i -eq 30 ]]; then
                print_error "Quiz App failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
        print_success "Quiz App is ready and healthy"
    fi
}

# Function to run API demo
run_demo() {
    if [[ "$DEMO" == true ]]; then
        print_info "Running API demo..."
        sleep 2  # Give the app a moment to fully initialize
        
        cd quiz-app
        if [[ -f "examples/api_demo.sh" ]]; then
            # Modify the demo script to use the correct port
            sed "s/localhost:8080/localhost:$PORT/g" examples/api_demo.sh > /tmp/api_demo_custom.sh
            chmod +x /tmp/api_demo_custom.sh
            /tmp/api_demo_custom.sh
            rm /tmp/api_demo_custom.sh
        else
            print_warning "API demo script not found"
        fi
        cd ..
    fi
}

# Function to display status and usage info
show_status() {
    print_success "🎉 Quiz App is running successfully!"
    echo
    
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        echo "Configuration:"
        echo "  Backend: DUAL (Redis + Aerospike)"
        echo "  Mode: $MODE"
        echo "  Redis Port: $REDIS_PORT"
        echo "  Aerospike Port: $AEROSPIKE_PORT"
        echo
        echo "Endpoints:"
        echo "  Redis Backend:"
        echo "    Health Check: http://localhost:$REDIS_PORT/health"
        echo "    API Base: http://localhost:$REDIS_PORT/api/v1"
        echo "  Aerospike Backend:"
        echo "    Health Check: http://localhost:$AEROSPIKE_PORT/health"
        echo "    API Base: http://localhost:$AEROSPIKE_PORT/api/v1"
        echo
        echo "Quick Test Commands:"
        echo "  # Test Redis backend:"
        echo "  curl http://localhost:$REDIS_PORT/health"
        echo "  curl http://localhost:$REDIS_PORT/api/v1/leaderboard"
        echo "  # Test Aerospike backend:"
        echo "  curl http://localhost:$AEROSPIKE_PORT/health"
        echo "  curl http://localhost:$AEROSPIKE_PORT/api/v1/leaderboard"
        echo
        echo "Performance Comparison:"
        echo "  time curl -X POST http://localhost:$REDIS_PORT/api/v1/users/test1/score -d '{\"score\":100}'"
        echo "  time curl -X POST http://localhost:$AEROSPIKE_PORT/api/v1/users/test1/score -d '{\"score\":100}'"
    else
        echo "Configuration:"
        echo "  Backend: $BACKEND"
        echo "  Mode: $MODE"
        echo "  Port: $PORT"
        echo
        echo "Endpoints:"
        echo "  Health Check: http://localhost:$PORT/health"
        echo "  API Base: http://localhost:$PORT/api/v1"
        echo
        echo "Quick Test Commands:"
        echo "  curl http://localhost:$PORT/health"
        echo "  curl http://localhost:$PORT/api/v1/leaderboard"
    fi
    
    echo
    echo "To stop the application:"
    echo "  Press Ctrl+C in this terminal"
    echo
    echo "Logs:"
    if [[ "$MODE" == "docker" ]]; then
        echo "  docker-compose -f quiz-app/docker-compose.yml logs -f"
    else
        echo "  Check console output above"
    fi
}

# Function to setup signal handlers
setup_signal_handlers() {
    if [[ "$MODE" != "docker" ]]; then
        trap cleanup_and_exit SIGINT SIGTERM
    else
        trap cleanup_docker_and_exit SIGINT SIGTERM
    fi
}

# Function to cleanup and exit (for local mode)
cleanup_and_exit() {
    print_info "Shutting down Quiz App..."
    
    # Kill application processes
    if [[ "$DUAL_BACKEND" == "true" ]]; then
        # Stop both Redis and Aerospike app processes
        if [[ -f "quiz-app-redis.pid" ]]; then
            REDIS_APP_PID=$(cat quiz-app-redis.pid)
            print_info "Stopping Redis App process (PID: $REDIS_APP_PID)..."
            kill -TERM $REDIS_APP_PID 2>/dev/null || true
            sleep 1
            if kill -0 $REDIS_APP_PID 2>/dev/null; then
                kill -KILL $REDIS_APP_PID 2>/dev/null || true
            fi
            rm quiz-app-redis.pid 2>/dev/null || true
        fi
        
        if [[ -f "quiz-app-aerospike.pid" ]]; then
            AEROSPIKE_APP_PID=$(cat quiz-app-aerospike.pid)
            print_info "Stopping Aerospike App process (PID: $AEROSPIKE_APP_PID)..."
            kill -TERM $AEROSPIKE_APP_PID 2>/dev/null || true
            sleep 1
            if kill -0 $AEROSPIKE_APP_PID 2>/dev/null; then
                kill -KILL $AEROSPIKE_APP_PID 2>/dev/null || true
            fi
            rm quiz-app-aerospike.pid 2>/dev/null || true
        fi
        
        # Stop both database containers
        print_info "Stopping database containers..."
        cd quiz-app
        print_info "Stopping Redis and Aerospike using docker-compose..."
        docker-compose stop redis aerospike 2>/dev/null || true
        cd ..
    else
        # Stop single application process
        if [[ -f "quiz-app.pid" ]]; then
            APP_PID=$(cat quiz-app.pid)
            print_info "Stopping Quiz App process (PID: $APP_PID)..."
            kill -TERM $APP_PID 2>/dev/null || true
            
            # Wait a moment for graceful shutdown
            sleep 2
            
            # Force kill if still running
            if kill -0 $APP_PID 2>/dev/null; then
                print_warning "Force killing Quiz App process..."
                kill -KILL $APP_PID 2>/dev/null || true
            fi
            
            rm quiz-app.pid 2>/dev/null || true
        fi
        
        # Stop database containers if they were started by this script
        print_info "Stopping database containers..."
        cd quiz-app
        if [[ "$BACKEND" == "redis" ]]; then
            print_info "Stopping Redis using docker-compose..."
            docker-compose stop redis 2>/dev/null || true
        else
            print_info "Stopping Aerospike using docker-compose..."
            docker-compose stop aerospike 2>/dev/null || true
        fi
        cd ..
    fi
    
    print_success "Shutdown complete"
    exit 0
}

# Function to cleanup Docker and exit (for Docker mode)
cleanup_docker_and_exit() {
    print_info "Shutting down Docker services..."
    
    cd quiz-app
    docker-compose down 2>/dev/null || true
    cd ..
    
    print_success "Docker services stopped"
    exit 0
}

# Main function
main() {
    echo "🚀 Quiz App Runner"
    echo "=================="
    
    parse_args "$@"
    validate_args
    check_prerequisites
    cleanup_containers
    install_dependencies
    build_application
    start_databases
    wait_for_database
    start_quiz_app
    wait_for_app
    run_demo
    show_status
    
    # Setup signal handlers for both modes
    setup_signal_handlers
    
    if [[ "$MODE" != "docker" ]]; then
        print_info "Press Ctrl+C to stop the application"
        # Keep the script running and wait for signals
        while true; do
            sleep 1
        done
    else
        print_info "Press Ctrl+C to stop all Docker services"
        # For Docker mode, we can just wait
        while true; do
            sleep 1
        done
    fi
}

# Run main function with all arguments
main "$@"
