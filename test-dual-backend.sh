#!/bin/bash

# Test script for dual backend functionality
# This script demonstrates how to test both Redis and Aerospike backends simultaneously

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Default ports for dual backend
REDIS_PORT=8081
AEROSPIKE_PORT=8082

echo "🧪 Dual Backend Test Script"
echo "==========================="
echo

print_info "This script will test both Redis and Aerospike backends running simultaneously"
print_info "Redis backend: http://localhost:$REDIS_PORT"
print_info "Aerospike backend: http://localhost:$AEROSPIKE_PORT"
echo

# Function to test a backend
test_backend() {
    local backend_name=$1
    local port=$2
    local base_url="http://localhost:$port"
    
    print_info "Testing $backend_name backend on port $port..."
    
    # Test health endpoint
    if curl -s "$base_url/health" | grep -q "healthy"; then
        print_success "$backend_name health check passed"
    else
        print_error "$backend_name health check failed"
        return 1
    fi
    
    # Create a test user
    print_info "Creating test user for $backend_name..."
    if curl -s -X POST "$base_url/api/v1/users" \
        -H "Content-Type: application/json" \
        -d '{"user_id":"test_'$backend_name'","name":"Test User '$backend_name'","email":"test@example.com"}' | grep -q "success"; then
        print_success "User created successfully in $backend_name"
    else
        print_warning "User creation failed or user already exists in $backend_name"
    fi
    
    # Update user score
    print_info "Updating score for $backend_name..."
    local score=$((RANDOM % 1000 + 100))
    if curl -s -X POST "$base_url/api/v1/users/test_$backend_name/score" \
        -H "Content-Type: application/json" \
        -d '{"score":'$score'}' | grep -q "success"; then
        print_success "Score updated successfully in $backend_name (Score: $score)"
    else
        print_error "Score update failed in $backend_name"
        return 1
    fi
    
    # Get leaderboard
    print_info "Getting leaderboard from $backend_name..."
    local leaderboard=$(curl -s "$base_url/api/v1/leaderboard")
    if echo "$leaderboard" | grep -q "test_$backend_name"; then
        print_success "Leaderboard retrieved successfully from $backend_name"
        echo "  Leaderboard preview: $(echo "$leaderboard" | jq -r '.data[0] // "No data"' 2>/dev/null || echo "Raw: $leaderboard")"
    else
        print_warning "User not found in leaderboard for $backend_name"
    fi
    
    echo
}

# Function to compare performance
compare_performance() {
    print_info "Running performance comparison..."
    
    local redis_url="http://localhost:$REDIS_PORT/api/v1/users/perf_test/score"
    local aerospike_url="http://localhost:$AEROSPIKE_PORT/api/v1/users/perf_test/score"
    
    print_info "Testing Redis performance..."
    local redis_time=$(time (
        for i in {1..10}; do
            curl -s -X POST "$redis_url" \
                -H "Content-Type: application/json" \
                -d '{"score":'$((RANDOM % 1000))'}' > /dev/null
        done
    ) 2>&1 | grep real | awk '{print $2}')
    
    print_info "Testing Aerospike performance..."
    local aerospike_time=$(time (
        for i in {1..10}; do
            curl -s -X POST "$aerospike_url" \
                -H "Content-Type: application/json" \
                -d '{"score":'$((RANDOM % 1000))'}' > /dev/null
        done
    ) 2>&1 | grep real | awk '{print $2}')
    
    print_success "Performance Results:"
    echo "  Redis: $redis_time (10 score updates)"
    echo "  Aerospike: $aerospike_time (10 score updates)"
    echo
}

# Main test execution
main() {
    # Check if both backends are running
    if ! curl -s "http://localhost:$REDIS_PORT/health" > /dev/null; then
        print_error "Redis backend not responding on port $REDIS_PORT"
        print_info "Please start the dual backend mode first:"
        print_info "  ./run-quiz-app.sh --dual"
        exit 1
    fi
    
    if ! curl -s "http://localhost:$AEROSPIKE_PORT/health" > /dev/null; then
        print_error "Aerospike backend not responding on port $AEROSPIKE_PORT"
        print_info "Please start the dual backend mode first:"
        print_info "  ./run-quiz-app.sh --dual"
        exit 1
    fi
    
    print_success "Both backends are running!"
    echo
    
    # Test Redis backend
    test_backend "Redis" $REDIS_PORT
    
    # Test Aerospike backend
    test_backend "Aerospike" $AEROSPIKE_PORT
    
    # Compare performance
    compare_performance
    
    print_success "🎉 Dual backend test completed!"
    print_info "Both Redis and Aerospike backends are working correctly"
    echo
    print_info "You can now compare:"
    print_info "  - API responses between backends"
    print_info "  - Performance characteristics"
    print_info "  - Data consistency"
    print_info "  - Leaderboard implementations (Redis sorted sets vs Aerospike ordered lists)"
}

# Run main function
main "$@"
