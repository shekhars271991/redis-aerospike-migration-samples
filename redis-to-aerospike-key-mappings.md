# Redis to Aerospike Key Mappings

## Quiz Application Data Structure Mappings

Based on the analysis of the quiz application codebase, here are the specific Redis key patterns and their corresponding Aerospike mappings:

---

## 1. User Profiles

### Redis
```
Key Pattern: user:<userID>
Data Type: Hash (HMSET)
Example: user:john123
Fields: {
  "user_id": "john123",
  "name": "John Doe", 
  "email": "john@example.com",
  "games_played": 5,
  "last_score": 850,
  "created_at": 1234567890,
  "updated_at": 1234567900
}
```

### Aerospike
```
Namespace: quiz
Set: users
Key: john123
Bins: [user_id, name, email, games_played, last_score, created_at, updated_at]
Example Record: {
  user_id: "john123",
  name: "John Doe",
  email: "john@example.com", 
  games_played: 5,
  last_score: 850,
  created_at: 1234567890,
  updated_at: 1234567900
}
```

---

## 2. Game Sessions

### Redis
```
Key Pattern: session:<sessionID>
Data Type: String with TTL (SET with EXPIRE)
Example: session:abc123def456
Value: JSON string with 1 hour TTL
Content: {
  "session_id": "abc123def456",
  "user_id": "john123",
  "start_time": 1234567890,
  "active": true
}
```

### Aerospike
```
Namespace: quiz
Set: users (reusing users set)
Key: session:abc123def456
Bins: [session_id, user_id, start_time, active]
Example Record: {
  session_id: "abc123def456",
  user_id: "john123", 
  start_time: 1234567890,
  active: true
}
TTL: Application-level expiration check (1 hour)
```

---

## 3. Global Leaderboard

### Redis
```
Key: leaderboard
Data Type: Sorted Set (ZADD, ZREVRANGEWITHSCORES)
Members: userID strings
Scores: user scores (float64)
Example: {
  "john123": 950.0,
  "jane456": 850.0,
  "bob789": 750.0
}
```

### Aerospike
```
Namespace: quiz
Set: leaderboard
Key: global
Bins: [scores]
Example Record: {
  scores: [
    [-950, "john123", "John Doe"],    // Negative score for desc order
    [-850, "jane456", "Jane Smith"],
    [-750, "bob789", "Bob Wilson"]
  ]
}
```

---

## Alternative Leaderboard Implementation (Current Aerospike Code)

### Aerospike (Alternative)
```
Namespace: quiz
Set: users (storing in users set)
Key: leaderboard:top
Bins: [top_users, updated]
Example Record: {
  top_users: [],  // Empty initially, populated via ordered list operations
  updated: 1234567890
}
```

---

## Summary of Mappings

| Redis Pattern | Redis Type | Aerospike Namespace | Aerospike Set | Aerospike Key | Aerospike Bins |
|---------------|------------|-------------------|---------------|---------------|----------------|
| `user:<userID>` | Hash | quiz | users | `<userID>` | [user_id, name, email, games_played, last_score, created_at, updated_at] |
| `session:<sessionID>` | String+TTL | quiz | users | `session:<sessionID>` | [session_id, user_id, start_time, active] |
| `leaderboard` | Sorted Set | quiz | leaderboard | global | [scores] |

---

## Key Design Principles Applied

### 1. Key Transformation
- **Redis**: Prefix-based keys (`user:123`, `session:abc`)
- **Aerospike**: Clean keys without prefixes, using namespace/set for organization

### 2. Data Structure Mapping
- **Hash → Multi-bin Record**: Each Redis hash field becomes an Aerospike bin
- **String → Single/Multi-bin Record**: JSON strings decomposed into multiple bins
- **Sorted Set → Ordered List**: Redis ZSET becomes Aerospike ordered list with composite entries

### 3. TTL Handling
- **Redis**: Native TTL support (`SETEX`, `EXPIRE`)
- **Aerospike**: Application-level expiration logic due to storage configuration

### 4. Namespace Organization
- **Primary Namespace**: `quiz` for all application data
- **Set Separation**: `users` for user data and sessions, `leaderboard` for ranking data

---

## Implementation Notes

### Atomic Operations
- **Redis Pipelines** → **Aerospike Batch Operations**
- **Redis Transactions** → **Aerospike Multi-operation Transactions**

### Data Consistency
- Maintain data types during migration (integers stay integers)
- Preserve relationships between records
- Handle concurrent access patterns

### Performance Considerations
- Create secondary indexes for non-primary key queries
- Use batch operations for bulk data operations
- Leverage Aerospike's native ordered lists for leaderboard functionality
