# Quiz App API - Postman Collection Guide

This guide explains how to use the comprehensive Postman collection for testing the Quiz/Game Leaderboard API.

## 📁 Files Included

- **`Quiz-App-API.postman_collection.json`** - Complete API collection with all endpoints
- **`Quiz-App-Environment.postman_environment.json`** - Environment variables for easy testing
- **`POSTMAN_GUIDE.md`** - This guide

## 🚀 Quick Setup

### 1. Import Collection and Environment

1. **Open Postman**
2. **Import Collection**:
   - Click "Import" → "Upload Files"
   - Select `Quiz-App-API.postman_collection.json`
3. **Import Environment**:
   - Click "Import" → "Upload Files" 
   - Select `Quiz-App-Environment.postman_environment.json`
4. **Select Environment**:
   - Click the environment dropdown (top right)
   - Select "Quiz App Environment"

### 2. Start the Quiz App

Make sure the Quiz App is running:
```bash
./run-quiz-app.sh
```

The app should be available at `http://localhost:8080`

## 📋 Collection Structure

### 🏥 **Health Check**
- **Health Check** - Verify service is running and database is connected

### 👥 **User Management**
- **Create User** - Register a new user in the system
- **Get User** - Retrieve user details and statistics
- **Create Test User (Alice)** - Pre-configured test user
- **Create Test User (Bob)** - Another pre-configured test user

### 🎮 **Game Management**
- **Start Game Session** - Begin a new game for a user
- **Get Game Session** - Check session details and status
- **Finish Game Session** - Complete game and update score

### 🏆 **Leaderboard**
- **Get Leaderboard (Default)** - Top 10 players
- **Get Top 5 Leaderboard** - Top 5 players
- **Get Top 20 Leaderboard** - Top 20 players

### 🔄 **Complete Game Flow**
- **End-to-end workflow** demonstrating full game lifecycle:
  1. Create Player
  2. Start Game
  3. Check Session
  4. Finish Game with Score
  5. Check Updated User Stats
  6. Check Leaderboard

### 🔙 **Legacy Endpoints**
- Backward compatibility endpoints without `/api/v1` prefix

## 🎯 Testing Scenarios

### **Scenario 1: Basic API Testing**

1. **Health Check** - Ensure service is running
2. **Create User** - Register a new player
3. **Get User** - Verify user was created
4. **Get Leaderboard** - Check current standings

### **Scenario 2: Complete Game Flow**

Run the "Complete Game Flow" folder in sequence:
- Creates a user with random data
- Starts a game session
- Verifies session is active
- Finishes game with a score
- Checks updated user statistics
- Verifies user appears in leaderboard

### **Scenario 3: Multi-User Leaderboard**

1. **Create Test User (Alice)**
2. **Create Test User (Bob)**
3. **Start Game** for Alice → **Finish Game** with high score
4. **Start Game** for Bob → **Finish Game** with different score
5. **Get Leaderboard** - See both users ranked by score

## 🔧 Environment Variables

The environment includes these variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `base_url` | Main API endpoint | `http://localhost:8080` |
| `user_id` | Auto-generated user ID | (dynamic) |
| `session_id` | Auto-generated session ID | (dynamic) |
| `alice_id` | Test user Alice ID | `alice123` |
| `bob_id` | Test user Bob ID | `bob456` |

## 🧪 Automated Testing Features

### **Pre-request Scripts**
- Generate unique user IDs using timestamps
- Set up test data automatically
- Initialize environment variables

### **Test Scripts**
- Validate response status codes
- Check response structure and data types
- Store response data for subsequent requests
- Verify business logic (e.g., scores are updated)

### **Global Tests**
- Response time validation (< 1000ms)
- Content-Type verification
- Basic response structure checks

## 📊 Example API Responses

### **User Creation**
```json
{
  "message": "User created successfully",
  "user_id": "alice123"
}
```

### **Game Session Start**
```json
{
  "session_id": "session_1234567890_987654",
  "user_id": "alice123",
  "message": "Game session started"
}
```

### **Leaderboard**
```json
{
  "leaderboard": [
    {
      "user_id": "alice123",
      "name": "Alice Johnson",
      "score": 2500
    },
    {
      "user_id": "bob456", 
      "name": "Bob Smith",
      "score": 1800
    }
  ],
  "count": 2
}
```

## 🚀 Advanced Usage

### **Running Collection with Newman**

Install Newman (Postman CLI):
```bash
npm install -g newman
```

Run the collection:
```bash
# Basic run
newman run Quiz-App-API.postman_collection.json -e Quiz-App-Environment.postman_environment.json

# With detailed reporting
newman run Quiz-App-API.postman_collection.json \
  -e Quiz-App-Environment.postman_environment.json \
  --reporters cli,html \
  --reporter-html-export results.html
```

### **Load Testing with Newman**

Run multiple iterations:
```bash
newman run Quiz-App-API.postman_collection.json \
  -e Quiz-App-Environment.postman_environment.json \
  -n 100 \
  --delay-request 100
```

### **CI/CD Integration**

Use in GitHub Actions, Jenkins, etc.:
```yaml
- name: Run API Tests
  run: |
    newman run Quiz-App-API.postman_collection.json \
      -e Quiz-App-Environment.postman_environment.json \
      --reporters junit \
      --reporter-junit-export results.xml
```

## 🔍 Troubleshooting

### **Common Issues**

1. **Connection Refused**
   - Ensure Quiz App is running: `./run-quiz-app.sh`
   - Check the `base_url` in environment variables

2. **User Already Exists**
   - Use the dynamic `{{$randomUUID}}` in user_id field
   - Or manually change user IDs in requests

3. **Session Not Found**
   - Ensure you run "Start Game" before "Finish Game"
   - Check that session_id is properly stored in environment

4. **Empty Leaderboard**
   - Create users and finish games with scores first
   - Leaderboard only shows users with scores > 0

### **Debug Tips**

- Enable **Postman Console** (View → Show Postman Console)
- Check **Tests** tab results for detailed validation
- Use **Environment** tab to verify variable values
- Check **Response** body and headers for error details

## 📈 Performance Testing

While this collection is great for functional testing, for performance testing use the dedicated load test client:

```bash
./run-loadtest.sh --config default
```

The Postman collection is perfect for:
- ✅ API validation and functional testing
- ✅ Integration testing
- ✅ Manual exploration and debugging
- ✅ CI/CD automated testing

For high-volume performance testing (1K+ QPS), use the dedicated load test client.

## 🎉 Happy Testing!

This collection provides comprehensive coverage of all Quiz App APIs with automated testing and easy-to-use workflows. Perfect for development, testing, and integration scenarios!
