#!/bin/bash

echo "🧪 Testing Container Restart Logic"
echo "=================================="

echo "Current Docker containers:"
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo
echo "Testing Redis container restart logic..."
cd quiz-app
make redis-run
echo

echo "Current containers after Redis start:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo
echo "Testing Redis restart (should restart existing container)..."
make redis-run

echo
echo "Final container status:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
