#!/bin/bash

# API Demo Script for Quiz/Game Leaderboard Service
# This script demonstrates all the API endpoints

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

echo "=== Quiz/Game Leaderboard API Demo ==="
echo "Base URL: $BASE_URL"
echo

# Check health
echo "1. Health Check"
curl -s "$BASE_URL/health" | jq '.' 2>/dev/null || curl -s "$BASE_URL/health"
echo -e "\n"

# Create users
echo "2. Creating Users"
echo "Creating user1 (Alice)..."
curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "name": "Alice Johnson",
    "email": "alice@example.com"
  }' | jq '.' 2>/dev/null || curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "name": "Alice Johnson",
    "email": "alice@example.com"
  }'

echo "Creating user2 (Bob)..."
curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user2",
    "name": "Bob Smith",
    "email": "bob@example.com"
  }' | jq '.' 2>/dev/null || curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user2",
    "name": "Bob Smith",
    "email": "bob@example.com"
  }'

echo "Creating user3 (Charlie)..."
curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user3",
    "name": "Charlie Brown",
    "email": "charlie@example.com"
  }' | jq '.' 2>/dev/null || curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user3",
    "name": "Charlie Brown",
    "email": "charlie@example.com"
  }'
echo -e "\n"

# Get user details
echo "3. Getting User Details"
echo "Getting user1 details..."
curl -s "$API_BASE/users/user1" | jq '.' 2>/dev/null || curl -s "$API_BASE/users/user1"
echo -e "\n"

# Simulate games for multiple users
echo "4. Simulating Games"

# User1 games
echo "User1 (Alice) playing games..."
SESSION1=$(curl -s -X POST "$API_BASE/game/start" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1"}' | jq -r '.session_id' 2>/dev/null)

if [ "$SESSION1" != "null" ] && [ -n "$SESSION1" ]; then
  echo "Session started: $SESSION1"
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION1\", \"score\": 1500}" | jq '.' 2>/dev/null || \
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION1\", \"score\": 1500}"
fi

# User1 second game
SESSION1_2=$(curl -s -X POST "$API_BASE/game/start" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1"}' | jq -r '.session_id' 2>/dev/null)

if [ "$SESSION1_2" != "null" ] && [ -n "$SESSION1_2" ]; then
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION1_2\", \"score\": 2100}" | jq '.' 2>/dev/null || \
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION1_2\", \"score\": 2100}"
fi

# User2 games
echo "User2 (Bob) playing games..."
SESSION2=$(curl -s -X POST "$API_BASE/game/start" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user2"}' | jq -r '.session_id' 2>/dev/null)

if [ "$SESSION2" != "null" ] && [ -n "$SESSION2" ]; then
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION2\", \"score\": 1800}" | jq '.' 2>/dev/null || \
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION2\", \"score\": 1800}"
fi

# User3 games
echo "User3 (Charlie) playing games..."
SESSION3=$(curl -s -X POST "$API_BASE/game/start" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user3"}' | jq -r '.session_id' 2>/dev/null)

if [ "$SESSION3" != "null" ] && [ -n "$SESSION3" ]; then
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION3\", \"score\": 900}" | jq '.' 2>/dev/null || \
  curl -s -X POST "$API_BASE/game/finish" \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION3\", \"score\": 900}"
fi

echo -e "\n"

# Get updated user details
echo "5. Updated User Details"
echo "User1 after games:"
curl -s "$API_BASE/users/user1" | jq '.' 2>/dev/null || curl -s "$API_BASE/users/user1"
echo -e "\n"

# Get leaderboard
echo "6. Leaderboard"
echo "Top 5 players:"
curl -s "$API_BASE/leaderboard?top=5" | jq '.' 2>/dev/null || curl -s "$API_BASE/leaderboard?top=5"
echo -e "\n"

echo "=== Demo Complete ==="
echo "Try these commands manually:"
echo "  curl $API_BASE/leaderboard"
echo "  curl $API_BASE/users/user1"
echo "  curl $BASE_URL/health"
