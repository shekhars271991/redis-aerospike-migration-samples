# Redis → Aerospike Data Modeling Guidelines
This document is auto-generated using a Cursor command workflow.
It identifies Redis usage in the codebase and provides placeholders
to define equivalent Aerospike data models.

---
## 1. Redis Data Access Summary
*(Automatically generated below)*

### Files with Redis Usage:

1. **quiz-app/internal/service/redis_service.go** (Primary Redis service implementation)
   - Line 23: `client *redis.Client` - Redis client field
   - Line 29: `rdb := redis.NewClient(&redis.Options{...})` - Client initialization
   - Line 76: `r.client.ZAdd(r.ctx, leaderboard, redis.Z{...})` - Sorted set operations
   - Line 132: `pipe := r.client.Pipeline()` - Pipeline operations
   - Line 140: `pipe.ZAdd(r.ctx, leaderboard, redis.Z{...})` - Pipeline sorted set
   - Line 226: `if err == redis.Nil` - Redis nil check
   - Line 262: `if err == redis.Nil` - Redis nil check

2. **quiz-app/cmd/server/main.go** (Service initialization)
   - Lines 29-31: Redis configuration usage (`cfg.Redis.Addr`, `cfg.Redis.Password`, `cfg.Redis.DB`)

### Redis Commands Identified:
- **Hash Operations**: `HMSet`, `HGetAll`, `HGet`, `HSet`, `HIncrBy`
- **Sorted Set Operations**: `ZAdd`, `ZRevRangeWithScores`
- **String Operations**: `Set`, `Get`
- **Key Operations**: `Exists`, `Del`
- **Pipeline Operations**: Used for atomic updates

---
## 2. Redis Key Usages

### Key Patterns and Operations:

#### 1. User Hash Keys (`user:` prefix)
- **File**: quiz-app/internal/service/redis_service.go
- **Pattern**: `userPrefix + userID` → `"user:" + userID`
- **Operations**:
  - Line 43: `userKey := userPrefix + userID` - Key construction
  - Line 46: `r.client.Exists(r.ctx, userKey)` - Check existence
  - Line 70: `r.client.HMSet(r.ctx, userKey, userData)` - Create user hash
  - Line 91: `r.client.HGetAll(r.ctx, userKey)` - Get all user data
  - Line 123: `r.client.Exists(r.ctx, userKey)` - Check existence
  - Line 135: `pipe.HIncrBy(r.ctx, userKey, "games_played", 1)` - Increment counter
  - Line 136: `pipe.HSet(r.ctx, userKey, "last_score", score)` - Update field
  - Line 137: `pipe.HSet(r.ctx, userKey, "updated_at", time.Now().Unix())` - Update timestamp
  - Line 168: `r.client.HGet(r.ctx, userPrefix+userID, "name")` - Get specific field

#### 2. Session String Keys (`session:` prefix)
- **File**: quiz-app/internal/service/redis_service.go
- **Pattern**: `sessionPrefix + sessionID` → `"session:" + sessionID`
- **Operations**:
  - Line 197: `sessionKey := sessionPrefix + sessionID` - Key construction
  - Line 212: `r.client.Set(r.ctx, sessionKey, sessionData, time.Hour)` - Create with TTL
  - Line 225: `r.client.Get(r.ctx, sessionKey)` - Retrieve session
  - Line 249: `r.client.Del(r.ctx, sessionKey)` - Delete session
  - Line 261: `r.client.Get(r.ctx, sessionKey)` - Retrieve session

#### 3. Leaderboard Sorted Set (fixed key)
- **File**: quiz-app/internal/service/redis_service.go
- **Key**: `"leaderboard"` (constant)
- **Operations**:
  - Line 76: `r.client.ZAdd(r.ctx, leaderboard, redis.Z{...})` - Add user with score 0
  - Line 140: `pipe.ZAdd(r.ctx, leaderboard, redis.Z{...})` - Update user score
  - Line 156: `r.client.ZRevRangeWithScores(r.ctx, leaderboard, 0, int64(topN-1))` - Get top N

---
## 3. Aerospike Mapping Templates

### Redis Key: `user:{userID}` (Hash)
**Current Redis Structure:**
```
Hash Fields:
- user_id: string
- name: string  
- email: string
- games_played: integer
- last_score: integer
- created_at: unix timestamp
- updated_at: unix timestamp
```

**Aerospike Mapping:**
- **Namespace**: `quiz_app`
- **Set**: `users`
- **Record Key**: `{userID}` (direct mapping from Redis key suffix)
- **Bins**:
  - `user_id`: string
  - `name`: string
  - `email`: string
  - `games_played`: integer
  - `last_score`: integer  
  - `created_at`: integer (unix timestamp)
  - `updated_at`: integer (unix timestamp)

**Notes:**
- Direct 1:1 mapping from Redis hash fields to Aerospike bins
- Record key removes the "user:" prefix for cleaner Aerospike design
- All data types remain consistent between Redis and Aerospike

---

### Redis Key: `session:{sessionID}` (String with JSON)
**Current Redis Structure:**
```
JSON String containing:
{
  "SessionID": string,
  "UserID": string, 
  "StartTime": unix timestamp,
  "Active": boolean
}
```

**Aerospike Mapping:**
- **Namespace**: `quiz_app`
- **Set**: `sessions`
- **Record Key**: `{sessionID}` (direct mapping from Redis key suffix)
- **Bins**:
  - `session_id`: string
  - `user_id`: string
  - `start_time`: integer (unix timestamp)
  - `active`: boolean

**TTL Considerations:**
- Redis: 1 hour expiration via `time.Hour`
- Aerospike: Set record TTL to 3600 seconds (1 hour) to match Redis behavior

**Notes:**
- Decompose JSON string into individual bins for better query performance
- Record key removes the "session:" prefix
- TTL behavior preserved from Redis implementation

---

### Redis Key: `leaderboard` (Sorted Set)
**Current Redis Structure:**
```
Sorted Set with:
- Members: userID strings
- Scores: user scores (float64)
```

**Aerospike Mapping:**
- **Namespace**: `quiz_app`
- **Set**: `leaderboard`
- **Record Key**: `"global"` (single record for global leaderboard)
- **Bins**:
  - `entries`: List of Maps, where each map contains:
    - `user_id`: string
    - `score`: integer
  - `last_updated`: integer (unix timestamp)

**Alternative Approach (for better scalability):**
- **Set**: `leaderboard_entries`
- **Record Key**: `{userID}`
- **Bins**:
  - `user_id`: string
  - `score`: integer
  - `updated_at`: integer (unix timestamp)

**Notes:**
- Primary approach stores all leaderboard data in a single record (simpler queries)
- Alternative approach uses individual records per user (better for large leaderboards)
- Consider data size limits when choosing approach
- May need custom sorting logic in application layer (Aerospike doesn't have native sorted sets)

---
## 4. Migration Strategy Summary

### Data Migration Approach:
1. **Users**: Direct hash-to-record mapping with bin-level field mapping
2. **Sessions**: JSON string decomposition into individual bins with TTL preservation
3. **Leaderboard**: Sorted set to list/map structure with application-level sorting

### Key Considerations:
- **Atomic Operations**: Redis pipelines → Aerospike batch operations
- **TTL Management**: Session expiration behavior must be preserved
- **Query Patterns**: Leaderboard queries will require application-level sorting
- **Data Types**: All primitive types map directly between systems

---
## 5. Validation Checklist

### Pre-Migration Validation:
- [ ] All Redis call sites identified and documented
- [ ] Each Redis key pattern mapped to corresponding Aerospike model
- [ ] Bin definitions match existing Redis field names and types
- [ ] TTL requirements identified and mapped
- [ ] Query patterns analyzed for compatibility

### Data Model Validation:
- [ ] **User Records**: Hash fields → Aerospike bins mapping verified
- [ ] **Session Records**: JSON decomposition strategy approved
- [ ] **Leaderboard**: Sorted set replacement strategy chosen and documented
- [ ] Namespace and set naming conventions established
- [ ] Record key generation logic defined

### Implementation Validation:
- [ ] No structural changes introduced unless explicitly approved
- [ ] Atomic operation patterns (pipelines) have Aerospike equivalents
- [ ] Error handling patterns (Redis Nil checks) adapted for Aerospike
- [ ] Connection and client initialization patterns updated

### Testing Requirements:
- [ ] Data migration scripts tested with sample data
- [ ] Query performance benchmarked against Redis baseline  
- [ ] TTL behavior validated for session management
- [ ] Leaderboard sorting performance measured
- [ ] Concurrent access patterns tested

### Documentation Review:
- [ ] Data modeling guidelines reviewed by architecture team
- [ ] Migration strategy approved by data modeling owner
- [ ] Rollback procedures documented
- [ ] Monitoring and alerting strategy defined

### Final Approval:
- [ ] Technical review completed
- [ ] Business stakeholder sign-off obtained
- [ ] Migration timeline and rollout plan approved

