#!/bin/bash

echo "🧪 Testing Docker Compose Container Management"
echo "=============================================="

cd quiz-app

echo "Current containers:"
docker-compose ps

echo
echo "Starting Redis with docker-compose..."
docker-compose up -d redis

echo
echo "Waiting 5 seconds..."
sleep 5

echo "Container status after Redis start:"
docker-compose ps

echo
echo "Starting Aerospike with docker-compose..."
docker-compose up -d aerospike

echo
echo "Waiting 10 seconds for Aerospike to start..."
sleep 10

echo "Final container status:"
docker-compose ps

echo
echo "Testing connectivity..."
echo "Redis ping:"
docker exec redis-quiz redis-cli ping || echo "Redis not ready"

echo
echo "Aerospike status:"
docker exec aerospike-quiz /opt/aerospike/bin/asinfo -v status || echo "Aerospike not ready"

echo
echo "Stopping services..."
docker-compose stop

echo "Done!"
