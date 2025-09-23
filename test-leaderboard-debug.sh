#!/bin/bash

echo "🔍 Debugging Leaderboard Issue"
echo "================================"

BASE_URL="http://localhost:8080"

echo "1. Creating a test user..."
USER_RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"username":"debuguser","email":"debug@test.com"}')
echo "User creation response: $USER_RESPONSE"

USER_ID=$(echo $USER_RESPONSE | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
echo "Extracted user_id: $USER_ID"

if [ -z "$USER_ID" ]; then
    echo "❌ Failed to create user or extract user_id"
    exit 1
fi

echo
echo "2. Starting a game session..."
SESSION_RESPONSE=$(curl -s -X POST $BASE_URL/sessions/start \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\"}")
echo "Session response: $SESSION_RESPONSE"

SESSION_ID=$(echo $SESSION_RESPONSE | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo "Extracted session_id: $SESSION_ID"

if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to create session or extract session_id"
    exit 1
fi

echo
echo "3. Finishing game with score 1000..."
FINISH_RESPONSE=$(curl -s -X POST $BASE_URL/sessions/finish \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"score\":1000}")
echo "Finish response: $FINISH_RESPONSE"

echo
echo "4. Checking user stats..."
USER_STATS=$(curl -s -X GET $BASE_URL/users/$USER_ID)
echo "User stats: $USER_STATS"

echo
echo "5. Checking leaderboard..."
LEADERBOARD=$(curl -s -X GET $BASE_URL/leaderboard)
echo "Leaderboard: $LEADERBOARD"

echo
echo "6. Checking leaderboard top 5..."
LEADERBOARD_TOP5=$(curl -s -X GET "$BASE_URL/leaderboard?top=5")
echo "Leaderboard top 5: $LEADERBOARD_TOP5"

echo
echo "🏁 Debug test completed!"


