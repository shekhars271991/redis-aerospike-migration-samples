#!/bin/bash

echo "Testing Load Test Client manually..."

cd loadtest-client

# Test with very simple parameters
echo "Running simple load test: 10 QPS for 10 seconds..."
./bin/loadtest --url http://localhost:8080 --qps 10 --duration 10s --workers 5

echo "Done!"
