# Redis to Aerospike Data Modeling Guidelines

---

## 1. Redis Data Access and Key Usages

| File Path | Line Number | Redis Command | Key Argument | Code Snippet |
|-----------|-------------|---------------|--------------|--------------|
| quiz-app/internal/service/redis_service.go | 46 | Exists | `userPrefix + userID` | `exists, err := r.client.Exists(r.ctx, userKey).Result()` |
| quiz-app/internal/service/redis_service.go | 70 | HMSet | `userPrefix + userID` | `err = r.client.HMSet(r.ctx, userKey, userData).Err()` |
| quiz-app/internal/service/redis_service.go | 76 | ZAdd | `leaderboard` | `err = r.client.ZAdd(r.ctx, leaderboard, redis.Z{Score: 0, Member: userID}).Err()` |
| quiz-app/internal/service/redis_service.go | 91 | HGetAll | `userPrefix + userID` | `result, err := r.client.HGetAll(r.ctx, userKey).Result()` |
| quiz-app/internal/service/redis_service.go | 123 | Exists | `userPrefix + userID` | `exists, err := r.client.Exists(r.ctx, userKey).Result()` |
| quiz-app/internal/service/redis_service.go | 132 | Pipeline | N/A | `pipe := r.client.Pipeline()` |
| quiz-app/internal/service/redis_service.go | 135 | HIncrBy | `userPrefix + userID` | `pipe.HIncrBy(r.ctx, userKey, "games_played", 1)` |
| quiz-app/internal/service/redis_service.go | 136 | HSet | `userPrefix + userID` | `pipe.HSet(r.ctx, userKey, "last_score", score)` |
| quiz-app/internal/service/redis_service.go | 137 | HSet | `userPrefix + userID` | `pipe.HSet(r.ctx, userKey, "updated_at", time.Now().Unix())` |
| quiz-app/internal/service/redis_service.go | 140 | ZAdd | `leaderboard` | `pipe.ZAdd(r.ctx, leaderboard, redis.Z{Score: float64(score), Member: userID})` |
| quiz-app/internal/service/redis_service.go | 145 | Exec | N/A | `_, err = pipe.Exec(r.ctx)` |
| quiz-app/internal/service/redis_service.go | 156 | ZRevRangeWithScores | `leaderboard` | `results, err := r.client.ZRevRangeWithScores(r.ctx, leaderboard, 0, int64(topN-1)).Result()` |
| quiz-app/internal/service/redis_service.go | 168 | HGet | `userPrefix + userID` | `userName, err := r.client.HGet(r.ctx, userPrefix+userID, "name").Result()` |
| quiz-app/internal/service/redis_service.go | 188 | Exists | `userPrefix + userID` | `exists, err := r.client.Exists(r.ctx, userKey).Result()` |
| quiz-app/internal/service/redis_service.go | 212 | Set | `sessionPrefix + sessionID` | `err = r.client.Set(r.ctx, sessionKey, sessionData, time.Hour).Err()` |
| quiz-app/internal/service/redis_service.go | 225 | Get | `sessionPrefix + sessionID` | `sessionData, err := r.client.Get(r.ctx, sessionKey).Result()` |
| quiz-app/internal/service/redis_service.go | 249 | Del | `sessionPrefix + sessionID` | `err = r.client.Del(r.ctx, sessionKey).Err()` |
| quiz-app/internal/service/redis_service.go | 261 | Get | `sessionPrefix + sessionID` | `sessionData, err := r.client.Get(r.ctx, sessionKey).Result()` |

---

## 2. Aerospike Mapping Template

### Redis Key: "user:{userID}"

**Data Structure:** Redis Hash

**Redis Commands:**
- `HMSet` (line 70): Store user profile data
- `HGetAll` (line 91): Retrieve all user fields
- `HGet` (line 168): Retrieve specific field (name)
- `HIncrBy` (line 135): Increment games_played counter
- `HSet` (lines 136, 137): Update last_score and updated_at

**Fields Stored:**
- user_id (string)
- name (string)
- email (string)
- games_played (integer)
- last_score (integer)
- created_at (unix timestamp)
- updated_at (unix timestamp)

**Aerospike Mapping:**
- **Namespace:** quiz
- **Set:** users
- **Record Key:** userID (string)
- **Bins:**
  - user_id: string
  - name: string
  - email: string
  - games_played: integer (atomic counter)
  - last_score: integer
  - created_at: integer (unix timestamp)
  - updated_at: integer (unix timestamp)

**Notes:**
- Each Redis hash field maps to a separate Aerospike bin
- Type conversions: Redis stores all as strings, Aerospike uses native integer types
- Use `Operate()` with `AddOp()` for atomic counter increment (games_played)
- No TTL required for user records (persistent data)
- Record size: ~200 bytes (well within 1MB limit)

---

### Redis Key: "leaderboard"

**Data Structure:** Redis Sorted Set

**Redis Commands:**
- `ZAdd` (lines 76, 140): Add/update user score in leaderboard
- `ZRevRangeWithScores` (line 156): Retrieve top N users by score (descending)

**Use Case:**
- Global leaderboard with real-time rankings
- Stores userID as member, score as floating-point value
- Automatically handles duplicates (updates score)
- Returns sorted results in O(log(N) + M) time

**Aerospike Mapping (Recommended: Secondary Index Approach):**
- **Namespace:** quiz
- **Set:** users (reuse existing user records)
- **Secondary Index:**
  - Index Name: idx_user_score
  - Bin: last_score
  - Type: NUMERIC
- **Query:** Query users ordered by last_score descending, limit N

**Alternative Aerospike Mapping (Ordered List Approach):**
- **Namespace:** quiz
- **Set:** leaderboard
- **Record Key:** "global_leaderboard" (single record)
- **Bins:**
  - scores (Ordered List): [
      {"user": "user1", "score": 200},
      {"user": "user2", "score": 150},
      ...
    ]

**Recommended Approach: Secondary Index**

**Rationale:**
- Scalability: Secondary index approach scales better for large leaderboards (10K+ users)
- Flexibility: Easy to query top N without loading entire dataset
- Maintenance: No need to maintain separate leaderboard record
- Consistency: Score updates automatically reflected in leaderboard queries

**Implementation Details:**
1. Create secondary index on `last_score` bin:
   ```
   CREATE INDEX idx_user_score ON quiz.users (last_score) NUMERIC
   ```
2. Query leaderboard:
   ```go
   stmt := as.NewStatement("quiz", "users")
   stmt.SetFilter(as.NewRangeFilter("last_score", 0, 999999))
   stmt.Addfilter(as.NewRangeFilter("last_score", as.DESCENDING))
   recordset, err := client.Query(nil, stmt)
   ```
3. For each record, extract userID, name, and score

**Notes:**
- Secondary index queries may be slower than Redis sorted sets for very large datasets
- Consider caching top N results if query performance is critical
- Ordered List approach has 1MB record size limitation (can store ~5000 entries)
- Use batch reads to fetch user names efficiently

---

### Redis Key: "session:{sessionID}"

**Data Structure:** Redis String (JSON-serialized data with TTL)

**Redis Commands:**
- `Set` (line 212): Create session with 1-hour expiration
- `Get` (lines 225, 261): Retrieve session data
- `Del` (line 249): Delete session after game ends

**Data Stored (JSON):**
```json
{
  "session_id": "abc123...",
  "user_id": "user1",
  "start_time": 1699612800,
  "active": true
}
```

**TTL:** 1 hour (3600 seconds)

**Aerospike Mapping:**
- **Namespace:** quiz
- **Set:** sessions
- **Record Key:** sessionID (string)
- **Bins:**
  - session_id: string
  - user_id: string
  - start_time: integer (unix timestamp)
  - active: boolean (or integer 1/0)
- **TTL:** 3600 seconds (1 hour)

**Implementation Details:**

Use `WritePolicy` to set TTL at record creation:
```go
writePolicy := as.NewWritePolicy(0, 3600) // 3600 seconds TTL
key, _ := as.NewKey("quiz", "sessions", sessionID)
bins := as.BinMap{
    "session_id": sessionID,
    "user_id": userID,
    "start_time": startTime,
    "active": true,
}
err := client.Put(writePolicy, key, bins)
```

**Notes:**
- TTL is set at record level, not bin level
- Aerospike automatically expires and removes records after TTL
- No need to store JSON string; use individual bins for structured data
- Use `Get()` to retrieve session; returns error if expired/not found
- Use `Delete()` to manually remove session (equivalent to Redis DEL)
- Record size: ~100 bytes (well within limits)
- Active sessions are temporary; expired sessions auto-cleaned by Aerospike

---

## 3. Atomic Operations Mapping

### Redis Pipeline (lines 132-145)

**Use Case:** Atomic update of user stats and leaderboard

**Redis Implementation:**
```go
pipe := r.client.Pipeline()
pipe.HIncrBy(r.ctx, userKey, "games_played", 1)
pipe.HSet(r.ctx, userKey, "last_score", score)
pipe.HSet(r.ctx, userKey, "updated_at", time.Now().Unix())
pipe.ZAdd(r.ctx, leaderboard, redis.Z{Score: float64(score), Member: userID})
_, err = pipe.Exec(r.ctx)
```

**Operations:**
1. Increment `games_played` counter
2. Update `last_score` field
3. Update `updated_at` timestamp
4. Update leaderboard score

**Aerospike Mapping:**

Use `Operate()` for single-record atomic operations:

```go
key, _ := as.NewKey("quiz", "users", userID)
ops := []*as.Operation{
    as.AddOp(as.NewBin("games_played", 1)),           // Atomic increment
    as.PutOp(as.NewBin("last_score", score)),          // Update score
    as.PutOp(as.NewBin("updated_at", time.Now().Unix())), // Update timestamp
}
record, err := client.Operate(nil, key, ops...)
```

**Notes:**
- Aerospike `Operate()` ensures all operations on a **single record** are atomic
- Leaderboard update is automatic (secondary index approach) or requires separate operation (ordered list approach)
- **Important:** Aerospike does NOT support multi-record transactions
- If leaderboard is a separate record (ordered list), atomicity across user + leaderboard is NOT guaranteed
- **Recommendation:** Use secondary index approach to avoid multi-record atomicity issue

---

## 4. Validation Checklist

- [ ] All Redis call sites captured
- [ ] Each Redis key mapped to an Aerospike model
- [ ] Bin definitions match existing Redis fields
- [ ] No structural changes unless approved
- [ ] Document reviewed by data modeling owner

