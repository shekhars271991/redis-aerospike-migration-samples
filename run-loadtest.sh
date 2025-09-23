#!/bin/bash

# Load Test Client Runner Script
# This script helps you run the high-performance load test client with different configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
CONFIG="default"
TARGET_URL="http://localhost:8080"
QPS=""
DURATION=""
WORKERS=""
OUTPUT_FILE=""
SKIP_BUILD=false
SKIP_DEPS=false
CONTINUOUS=false

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

Load Test Client Runner - High-performance load testing for quiz/game leaderboard API

OPTIONS:
    -c, --config CONFIG     Configuration profile: default, high-load, or stress-test
    -u, --url URL          Target API base URL (default: http://localhost:8080)
    -q, --qps QPS          Target queries per second (overrides config)
    -d, --duration TIME    Test duration (overrides config, e.g., 60s, 5m)
    -w, --workers COUNT    Number of worker goroutines (overrides config)
    -o, --output FILE      Output file for results (overrides config)
    --skip-build          Skip building the load test client
    --skip-deps           Skip installing dependencies
    --continuous          Run continuous tests (restart after completion)
    -h, --help            Show this help message

CONFIGURATION PROFILES:
    default       - Balanced test: 1K QPS for 60s (good for development)
    high-load     - High throughput: 100K QPS for 5min (production simulation)
    stress-test   - Extreme load: 1M QPS for 2min (WARNING: Can overwhelm systems)

EXAMPLES:
    $0                                          # Run default config (1K QPS)
    $0 --config high-load                      # Run high load test (100K QPS)
    $0 --config stress-test                    # Run stress test (1M QPS - BE CAREFUL!)
    $0 --url http://localhost:8081             # Test different endpoint
    $0 --qps 5000 --duration 30s --workers 200 # Custom parameters
    $0 --config default --continuous           # Continuous testing
    $0 --output my_results.json                # Custom output file

PERFORMANCE NOTES:
    - For 100K+ QPS: Use large AWS instances (c5.24xlarge or larger)
    - For 1M QPS: Use c6i.32xlarge with system tuning
    - Monitor target service capacity to avoid overwhelming it
    - Start with 'default' config, then scale up gradually

EOF
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--config)
                CONFIG="$2"
                shift 2
                ;;
            -u|--url)
                TARGET_URL="$2"
                shift 2
                ;;
            -q|--qps)
                QPS="$2"
                shift 2
                ;;
            -d|--duration)
                DURATION="$2"
                shift 2
                ;;
            -w|--workers)
                WORKERS="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --continuous)
                CONTINUOUS=true
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
    if [[ "$CONFIG" != "default" && "$CONFIG" != "high-load" && "$CONFIG" != "stress-test" ]]; then
        print_error "Invalid config: $CONFIG. Must be 'default', 'high-load', or 'stress-test'"
        exit 1
    fi

    # Validate URL format
    if [[ ! "$TARGET_URL" =~ ^https?:// ]]; then
        print_error "Invalid URL format: $TARGET_URL. Must start with http:// or https://"
        exit 1
    fi

    # Validate numeric parameters if provided
    if [[ -n "$QPS" && ! "$QPS" =~ ^[0-9]+$ ]]; then
        print_error "Invalid QPS: $QPS. Must be a positive integer"
        exit 1
    fi

    if [[ -n "$WORKERS" && ! "$WORKERS" =~ ^[0-9]+$ ]]; then
        print_error "Invalid workers count: $WORKERS. Must be a positive integer"
        exit 1
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
    if [[ ! -d "loadtest-client" ]]; then
        print_error "loadtest-client directory not found. Please run this script from the project root."
        exit 1
    fi

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

    # Check if target URL is reachable
    print_info "Checking target URL connectivity..."
    if ! curl -s --connect-timeout 5 "$TARGET_URL/health" > /dev/null 2>&1; then
        print_warning "Target URL $TARGET_URL may not be reachable"
        print_warning "Make sure the Quiz App is running before starting load tests"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "Target URL is reachable"
    fi

    print_success "Prerequisites check passed"
}

# Function to install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == false ]]; then
        print_info "Installing dependencies..."
        cd loadtest-client
        make deps
        cd ..
        print_success "Dependencies installed"
    fi
}

# Function to build application
build_application() {
    if [[ "$SKIP_BUILD" == false ]]; then
        print_info "Building load test client..."
        cd loadtest-client
        make build-prod
        cd ..
        print_success "Load test client built"
    fi
}

# Function to prepare configuration
prepare_config() {
    print_info "Preparing configuration..."
    
    # Build command line arguments
    CMD_ARGS="--config configs/${CONFIG}.yaml --url $TARGET_URL"
    
    # Add overrides if specified
    if [[ -n "$QPS" ]]; then
        CMD_ARGS="$CMD_ARGS --qps $QPS"
    fi
    
    if [[ -n "$DURATION" ]]; then
        CMD_ARGS="$CMD_ARGS --duration $DURATION"
    fi
    
    if [[ -n "$WORKERS" ]]; then
        CMD_ARGS="$CMD_ARGS --workers $WORKERS"
    fi
    
    if [[ -n "$OUTPUT_FILE" ]]; then
        CMD_ARGS="$CMD_ARGS --output $OUTPUT_FILE"
    fi
    
    print_success "Configuration prepared"
}

# Function to show test configuration
show_test_config() {
    print_info "Load Test Configuration:"
    echo "  Profile: $CONFIG"
    echo "  Target URL: $TARGET_URL"
    if [[ -n "$QPS" ]]; then
        echo "  QPS Override: $QPS"
    fi
    if [[ -n "$DURATION" ]]; then
        echo "  Duration Override: $DURATION"
    fi
    if [[ -n "$WORKERS" ]]; then
        echo "  Workers Override: $WORKERS"
    fi
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "  Output File: $OUTPUT_FILE"
    fi
    echo "  Continuous Mode: $CONTINUOUS"
    echo
}

# Function to show safety warning for high load tests
show_safety_warning() {
    if [[ "$CONFIG" == "stress-test" ]]; then
        print_warning "⚠️  STRESS TEST WARNING ⚠️"
        echo "You are about to run a STRESS TEST that can generate 1M+ QPS!"
        echo "This can overwhelm target systems and cause service disruption."
        echo "Only run this on systems designed to handle extreme loads."
        echo
        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Stress test cancelled"
            exit 0
        fi
    elif [[ "$CONFIG" == "high-load" ]]; then
        print_warning "⚠️  HIGH LOAD TEST WARNING ⚠️"
        echo "You are about to run a HIGH LOAD TEST (100K QPS)."
        echo "Make sure your target system can handle this load."
        echo
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "High load test cancelled"
            exit 0
        fi
    fi
}

# Function to run load test
run_load_test() {
    print_info "Starting load test..."
    cd loadtest-client
    
    # Store PID for cleanup
    ./bin/loadtest $CMD_ARGS &
    LOADTEST_PID=$!
    echo $LOADTEST_PID > ../loadtest.pid
    
    # Wait for the load test to complete
    wait $LOADTEST_PID
    LOADTEST_EXIT_CODE=$?
    
    # Clean up PID file
    rm ../loadtest.pid 2>/dev/null || true
    
    cd ..
    
    if [[ $LOADTEST_EXIT_CODE -eq 0 ]]; then
        print_success "Load test completed successfully"
    else
        print_error "Load test failed with exit code $LOADTEST_EXIT_CODE"
        return $LOADTEST_EXIT_CODE
    fi
}

# Function to run continuous tests
run_continuous_tests() {
    local test_count=1
    
    print_info "Starting continuous load testing mode..."
    print_info "Press Ctrl+C to stop continuous testing"
    
    while true; do
        print_info "Starting test run #$test_count..."
        
        if run_load_test; then
            print_success "Test run #$test_count completed"
        else
            print_error "Test run #$test_count failed"
        fi
        
        test_count=$((test_count + 1))
        
        # Wait a bit between tests
        print_info "Waiting 10 seconds before next test..."
        sleep 10
    done
}

# Function to setup signal handlers
setup_signal_handlers() {
    trap cleanup_and_exit SIGINT SIGTERM
}

# Function to cleanup and exit
cleanup_and_exit() {
    print_info "Stopping load test..."
    
    # Kill the load test process
    if [[ -f "loadtest.pid" ]]; then
        LOADTEST_PID=$(cat loadtest.pid)
        print_info "Stopping load test process (PID: $LOADTEST_PID)..."
        kill -TERM $LOADTEST_PID 2>/dev/null || true
        
        # Wait a moment for graceful shutdown
        sleep 2
        
        # Force kill if still running
        if kill -0 $LOADTEST_PID 2>/dev/null; then
            print_warning "Force killing load test process..."
            kill -KILL $LOADTEST_PID 2>/dev/null || true
        fi
        
        rm loadtest.pid 2>/dev/null || true
    fi
    
    print_success "Load test stopped"
    exit 0
}

# Function to show post-test information
show_post_test_info() {
    echo
    print_success "🎯 Load Test Session Complete!"
    echo
    echo "Results:"
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "  Detailed results saved to: loadtest-client/$OUTPUT_FILE"
    else
        echo "  Check the console output above for detailed statistics"
    fi
    echo
    echo "Next Steps:"
    echo "  - Analyze response times and error rates"
    echo "  - Check target service logs for any issues"
    echo "  - Scale up gradually if testing higher loads"
    echo
    echo "Available Configs:"
    echo "  ./run-loadtest.sh --config default      # 1K QPS (safe)"
    echo "  ./run-loadtest.sh --config high-load    # 100K QPS"
    echo "  ./run-loadtest.sh --config stress-test  # 1M QPS (extreme)"
    echo
}

# Main function
main() {
    echo "⚡ Load Test Client Runner"
    echo "========================="
    
    parse_args "$@"
    validate_args
    check_prerequisites
    install_dependencies
    build_application
    prepare_config
    show_test_config
    show_safety_warning
    
    # Setup signal handlers
    setup_signal_handlers
    
    if [[ "$CONTINUOUS" == true ]]; then
        run_continuous_tests
    else
        run_load_test
        show_post_test_info
    fi
}

# Run main function with all arguments
main "$@"
