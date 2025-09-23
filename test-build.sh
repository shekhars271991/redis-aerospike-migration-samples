#!/bin/bash

echo "Testing Go compilation..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Go not found, trying to add to PATH..."
    export PATH=$PATH:/usr/local/go/bin
fi

if ! command -v go &> /dev/null; then
    echo "Go still not found. Please install Go 1.23+"
    exit 1
fi

echo "Go version: $(go version)"

# Test compilation
cd quiz-app
echo "Building Quiz App..."
go mod tidy
go build -o bin/server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Quiz App builds successfully!"
else
    echo "❌ Quiz App build failed"
    exit 1
fi

cd ../loadtest-client
echo "Building Load Test Client..."
go mod tidy
go build -o bin/loadtest cmd/loadtest/main.go

if [ $? -eq 0 ]; then
    echo "✅ Load Test Client builds successfully!"
else
    echo "❌ Load Test Client build failed"
    exit 1
fi

echo "🎉 All builds successful!"
