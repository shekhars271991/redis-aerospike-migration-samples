# Database Migration Strategy: Redis → Aerospike

**Generated:** 2025-11-10 11:25 IST  
**Project:** Quiz Application - Redis to Aerospike Migration  
**Analysis Scope:** quiz-app/ directory

---

## 1. Overview

### Current Architecture

The Quiz Application currently uses Redis as its primary datastore with a well-architected design:

**Architecture Highlights:**
- **Clean Abstraction:** Database operations abstracted through `DBService` interface
- **Single Responsibility:** All Redis logic contained in `RedisService` implementation
- **8 Interface Methods:** CreateUser, GetUser, UpdateScore, GetLeaderboard, StartGameSession, EndGameSession, GetGameSession, HealthCheck

**Redis Usage Statistics:**
- **Files with Redis Usage:** 1 service file (redis_service.go)
- **Total Lines of Redis Code:** 288 lines
- **Data Structures Used:** 3 types (Hashes, Sorted Sets, Strings with TTL)
- **Operations:** CRUD, atomic transactions, TTL management
- **Complexity:** Medium (well-structured but uses multiple data types)

### Redis Data Structures Identified

| Data Structure | Operations Used | Use Case | Complexity |
|----------------|----------------|----------|------------|
| **Redis Hashes** | HMSet, HGetAll, HGet, HSet, HIncrBy | User profile storage | Medium |
| **Redis Sorted Sets** | ZAdd, ZRevRangeWithScores | Leaderboard rankings | High |
| **Redis Strings (TTL)** | Set (with expiration), Get, Del | Game sessions | Low |
| **Redis Keys** | Exists, Del | Key existence checks | Low |
| **Redis Pipeline** | Pipeline, Exec | Atomic operations | Medium |

### Migration Approach

This migration will:
1. **Preserve Interface:** Maintain `DBService` interface - no changes to handlers/routers
2. **Parallel Implementation:** Create `AerospikeService` alongside existing `RedisService`
3. **Phased Migration:** 5 distinct phases with testing at each step
4. **Zero Downtime:** Support dual-write mode during transition
5. **Rollback Ready:** Each phase can be independently reverted

**Estimated Complexity:** Medium  
**Estimated Duration:** 30-40 hours over 2-3 weeks  
**Risk Level:** Low (excellent interface abstraction reduces risk)

---

## 2. Impacted Files Summary

### A. Files to be Updated

| File Path | Change Type | Description of Required Update |
|-----------|-------------|-------------------------------|
| `quiz-app/go.mod` | Modify | Add Aerospike Go client dependency: `github.com/aerospike/aerospike-client-go/v7` |
| `quiz-app/internal/config/config.go` | Modify | Add `AerospikeConfig` struct with Host, Port, Namespace. Add environment variable parsing for Aerospike settings. |
| `quiz-app/cmd/server/main.go` | Modify | Update database initialization to support both `redis` and `aerospike` via `DB_TYPE` config. Add conditional service instantiation. |
| `quiz-app/docker-compose.yml` | Modify | Add Aerospike service definition. Configure health checks and networking. Keep Redis service for parallel operation during migration. |
| `quiz-app/Makefile` | Modify | Add Aerospike targets: `aerospike-run`, `aerospike-stop`, `aerospike-logs`. Update `dev-setup` to support both databases. |

### B. Files to be Created

| File Path | Purpose | Description |
|-----------|---------|-------------|
| `quiz-app/internal/service/aerospike_service.go` | Service Implementation | New Aerospike implementation of `DBService` interface. Implements all 8 methods with equivalent Aerospike operations. |
| `quiz-app/internal/service/aerospike_service_test.go` | Testing | Comprehensive test suite for Aerospike service. Tests all CRUD operations, leaderboard functionality, session management, and error handling. |
| `quiz-app/internal/service/utils.go` | Utility Functions | Shared utility functions for ID generation (`generateSessionID`) and data conversion helpers used by both services. |
| `quiz-app/scripts/migrate-redis-to-aerospike.go` | Data Migration | Script to migrate existing user data, sessions, and leaderboard entries from Redis to Aerospike. Supports batch processing and validation. |
| `quiz-app/scripts/validate-migration.go` | Validation | Compares data between Redis and Aerospike to verify migration integrity. Generates report with statistics and discrepancies. |
| `quiz-app/aerospike.conf` | Configuration | Aerospike server configuration for local development. Defines namespace `quiz`, sets, and storage parameters. |

### C. Optional Files (Manual Review)

| File Path | Reason for Review | Suggested Action |
|-----------|-------------------|------------------|
| `quiz-app/internal/service/db_service.go` | Interface Definition | Verify interface is sufficient for Aerospike. Current 8 methods are compatible. No changes needed. |
| `quiz-app/internal/models/user.go` | Data Models | Review data structures for Aerospike bin compatibility. Ensure field types match Aerospike native types. |
| `quiz-app/internal/service/redis_service_test.go` | Existing Tests | Review test patterns and coverage. Use as reference for Aerospike tests. Do not modify (will be deprecated after migration). |
| `quiz-app/internal/handlers/handlers.go` | HTTP Handlers | Verify handlers only use `DBService` interface. Should require zero changes if interface abstraction is clean. |
| `quiz-app/README.md` | Documentation | Update with Aerospike setup instructions, environment variables, and deployment notes. |

---

## 3. Phased Migration Plan with Testing Strategy

### Migration Approach

This migration follows a phased approach where each phase:
1. **Pre-Migration Tests**: Establish baseline behavior with Redis
2. **Implementation**: Build Aerospike equivalent functionality  
3. **Post-Migration Tests**: Verify Aerospike matches Redis behavior
4. **Validation Tests**: Compare outputs, performance, and data integrity
5. **Rollback Plan**: Clear path to revert if issues arise

Each phase is independently testable and can be rolled back without affecting other phases.

---

### Phase 1: Foundation & Infrastructure Setup

**Objective:** Set up Aerospike infrastructure and create service skeleton

**Pre-Migration Tests:**
- [ ] Run existing Redis test suite and capture baseline results
- [ ] Document current Redis performance metrics (latency, throughput)
- [ ] Verify all 8 interface methods are covered by tests
- [ ] Capture sample Redis data for validation

**Implementation Steps:**
1. Add Aerospike dependency to `go.mod`: `go get github.com/aerospike/aerospike-client-go/v7`
2. Create `AerospikeConfig` struct in `config.go` with Host, Port, Namespace fields
3. Add Aerospike environment variable parsing (AEROSPIKE_HOST, AEROSPIKE_PORT, AEROSPIKE_NAMESPACE)
4. Add Aerospike service to `docker-compose.yml` with health checks
5. Create `aerospike_service.go` with basic struct and constructor
6. Implement `HealthCheck()` method
7. Update `main.go` to support `DB_TYPE=aerospike`

**Post-Migration Tests:**
- [ ] Test Aerospike connection establishment
- [ ] Verify HealthCheck() returns success
- [ ] Validate configuration parsing from environment
- [ ] Test Aerospike container starts and is healthy

**Validation Tests:**
- [ ] Aerospike service initializes without errors
- [ ] Configuration correctly loaded
- [ ] Can switch between Redis and Aerospike via DB_TYPE

**Success Criteria:**
- Aerospike service compiles and initializes
- Health check passes consistently
- No impact on existing Redis functionality

**Rollback Plan:** Remove Aerospike config, delete aerospike_service.go, revert to Redis-only

---

### Phase 2: User Management (Hash Operations)

**Objective:** Migrate user CRUD operations from Redis Hashes to Aerospike Records

**Redis Operations:** `HMSet`, `HGetAll`, `HGet`, `Exists`  
**Aerospike Mapping:** Hash → Multi-bin Record

**Pre-Migration Tests:**
- [ ] Run Redis user creation tests (capture 100 sample users)
- [ ] Test user retrieval with various data types
- [ ] Verify hash field type conversions (string → int for games_played, last_score)
- [ ] Test error handling (duplicate user, user not found)
- [ ] Document performance: user creation, retrieval (avg latency)

**Implementation Steps:**
1. Implement `CreateUser()` in `AerospikeService`
   - Map Redis key `user:{userID}` → Aerospike key (namespace="quiz", set="users", key=userID)
   - Map each hash field to Aerospike bin
   - Handle type conversions
2. Implement `GetUser()` in `AerospikeService`
   - Retrieve all bins from record
   - Convert bin types back to expected format
   - Handle missing user error
3. Create unit tests in `aerospike_service_test.go`

**Post-Migration Tests:**
- [ ] Test `CreateUser()` creates record correctly
- [ ] Verify all fields stored with correct types
- [ ] Test `GetUser()` retrieves data matching input
- [ ] Test type conversion accuracy (string ↔ int)
- [ ] Compare output structure: Redis vs Aerospike (must be identical)
- [ ] Test error handling (duplicate user returns error)
- [ ] Test user not found scenario

**Validation Tests:**
- [ ] Create same 100 users in both Redis and Aerospike
- [ ] Compare retrieved data (must match 100%)
- [ ] Performance comparison: 1000 user create/get operations
- [ ] Verify no data loss in type conversions

**Success Criteria:**
- All user CRUD tests pass
- Data structure parity with Redis
- Performance within 20% of Redis
- No data loss or corruption

**Rollback Plan:** Continue using Redis for user operations, remove Aerospike user methods

---

### Phase 3: Leaderboard (Sorted Set Operations)

**Objective:** Migrate leaderboard functionality from Redis Sorted Set to Aerospike

**Redis Operations:** `ZAdd`, `ZRevRangeWithScores`  
**Aerospike Mapping:** Sorted Set → Secondary Index on score + Query (recommended) OR Ordered List

**Pre-Migration Tests:**
- [ ] Test Redis `ZAdd` with various scores (0, negative, large numbers)
- [ ] Test `ZRevRangeWithScores` for top-N queries (top 10, 100, 1000)
- [ ] Verify score ordering is correct
- [ ] Test leaderboard with 100, 1000, 10000 users
- [ ] Document performance: leaderboard query latency

**Implementation Steps:**
1. Design leaderboard data structure (recommend: Secondary Index approach)
2. Implement `UpdateScore()` in `AerospikeService`
   - Update user record with new score
   - Aerospike secondary index will auto-update
3. Implement `GetLeaderboard()` in `AerospikeService`
   - Query users ordered by score descending
   - Limit to top N results
   - Fetch user names via batch operation
4. Handle atomic score + leaderboard update

**Post-Migration Tests:**
- [ ] Test score updates in Aerospike
- [ ] Verify `GetLeaderboard()` returns correctly sorted results
- [ ] Test with duplicate scores (tie-breaking)
- [ ] Test leaderboard with 10, 100, 1000, 10000 users
- [ ] Compare top-10 results: Redis vs Aerospike (must match)
- [ ] Test after multiple score updates for same user

**Validation Tests:**
- [ ] Insert same 1000 users with random scores in both systems
- [ ] Compare top 100 leaderboards (order must match exactly)
- [ ] Performance comparison: leaderboard queries (various N values)
- [ ] Test score updates and verify re-ranking works
- [ ] Stress test: 10K concurrent score updates

**Success Criteria:**
- Leaderboard ordering matches Redis exactly
- Performance within acceptable range (< 2x Redis latency)
- Handles large leaderboards (10K+ users) efficiently
- Atomic score + leaderboard updates work correctly

**Rollback Plan:** Continue using Redis sorted set for leaderboard

---

### Phase 4: Session Management (TTL & Expiration)

**Objective:** Migrate game sessions from Redis Strings with TTL to Aerospike Records with TTL

**Redis Operations:** `Set` (with expiration), `Get`, `Del`  
**Aerospike Mapping:** String with TTL → Record with TTL (WritePolicy)

**Pre-Migration Tests:**
- [ ] Test Redis session creation with 1-hour TTL
- [ ] Verify session retrieval before expiration
- [ ] Verify session expires and returns error after TTL
- [ ] Test session deletion (EndGameSession)
- [ ] Document TTL behavior and timing

**Implementation Steps:**
1. Implement `StartGameSession()` in `AerospikeService`
   - Create session record with TTL using WritePolicy
   - Store JSON data in single bin
2. Implement `GetGameSession()` in `AerospikeService`
   - Retrieve session record
   - Unmarshal JSON data
   - Handle expired session (returns nil error from Aerospike)
3. Implement `EndGameSession()` in `AerospikeService`
   - Delete session record
   - Update user score (call UpdateScore)

**Post-Migration Tests:**
- [ ] Test session creation with TTL
- [ ] Verify session data retrieval is correct
- [ ] Test session expiration (wait for TTL + 10 seconds, verify deleted)
- [ ] Test session end and cleanup
- [ ] Compare session data structure: Redis vs Aerospike
- [ ] Test concurrent session operations

**Validation Tests:**
- [ ] Create 100 sessions in both systems
- [ ] Verify all sessions expire after TTL (allow ±5 second margin)
- [ ] Test session workflow: create → get → end → verify deleted
- [ ] Performance comparison: session CRUD operations

**Success Criteria:**
- Session TTL behavior matches Redis
- Session data integrity maintained
- Expired sessions automatically cleaned up
- No memory leaks from expired sessions

**Rollback Plan:** Continue using Redis for session management

---

### Phase 5: Atomic Operations & Data Migration

**Objective:** Ensure atomic updates work correctly and migrate existing data

**Redis Operations:** `Pipeline`, `Exec`  
**Aerospike Mapping:** Pipeline → Batch operations for reads, Operate() for single-record atomicity

**5A: Atomic Operations**

**Pre-Migration Tests:**
- [ ] Test Redis pipeline in `UpdateScore` (HIncrBy + HSet + ZAdd)
- [ ] Verify atomicity: all operations succeed or all fail
- [ ] Test concurrent updates to same user
- [ ] Document consistency guarantees

**Implementation Steps:**
1. Refactor `UpdateScore()` to use Aerospike `Operate()`
   - Single atomic operation updating multiple bins + leaderboard
2. Test transaction-like behavior
3. Handle partial failure scenarios

**Post-Migration Tests:**
- [ ] Test atomic user stats + score update
- [ ] Verify all fields updated or none (atomicity)
- [ ] Test error handling and rollback behavior
- [ ] Test 100 concurrent updates to same user

**Validation Tests:**
- [ ] Concurrent stress test: 1000 simultaneous score updates
- [ ] Verify final state matches expected (no lost updates)
- [ ] Compare Redis pipeline vs Aerospike Operate() consistency

**5B: Data Migration**

**Pre-Migration Tests:**
- [ ] Take Redis data snapshot (SAVE command)
- [ ] Count total users: `DBSIZE`
- [ ] Count leaderboard entries: `ZCARD leaderboard`
- [ ] Count active sessions: `KEYS session:*`
- [ ] Document data integrity checksums

**Implementation Steps:**
1. Create `scripts/migrate-redis-to-aerospike.go`
   - Connect to both Redis and Aerospike
   - Migrate users in batches of 100
   - Migrate leaderboard entries
   - Migrate active sessions (preserve TTL)
2. Implement progress tracking and logging
3. Add data validation checks
4. Run migration in staging environment

**Post-Migration Tests:**
- [ ] Verify user count matches (Redis count == Aerospike count)
- [ ] Verify leaderboard order preserved
- [ ] Verify active sessions migrated with correct TTL
- [ ] Random sampling: compare 100 users in both systems
- [ ] Validate no data corruption

**Validation Tests:**
- [ ] Run `scripts/validate-migration.go` comparing all records
- [ ] Check for data loss (expect 0 missing records)
- [ ] Verify field types and values match
- [ ] Full data integrity report

**Success Criteria:**
- 100% data migration success
- No data corruption or loss
- Atomic operations work as expected
- Migration completes in reasonable time (< 1 hour for 10K users)

**Rollback Plan:** Keep Redis data intact during migration, can revert anytime

---

## 4. Technical Notes

### Redis Data Structures Discovered

The codebase uses 3 primary Redis data structures:
1. **Hashes (3 occurrences):** User profile data storage
2. **Sorted Sets (1 occurrence):** Leaderboard with score-based ranking
3. **Strings with TTL (1 occurrence):** Temporary session data

### High-Level Aerospike Equivalents

| Redis Structure | Aerospike Equivalent | Complexity |
|-----------------|---------------------|------------|
| Hash (user data) | Multi-bin Record | Low |
| Sorted Set (leaderboard) | Secondary Index + Query OR Ordered List | High |
| String with TTL (sessions) | Single-bin Record with WritePolicy TTL | Low |
| Pipeline (atomic ops) | Operate() for single-record atomicity | Medium |

### Key Architectural Differences

**Type System:**
- Redis stores everything as strings, requires manual type conversion
- Aerospike has native types (int, string, list, map) - cleaner type handling

**Record Size:**
- Redis: No hard limit per key (practical limit ~512MB)
- Aerospike: Default 1MB per record (configurable to 8MB max)
- Impact: Session JSON data should stay well under 1MB (current size ~200 bytes)

**Data Structure Equivalents:**
- Hashes map directly to multi-bin records (1:1)
- Sorted Sets require design choice (secondary index vs ordered list)
- Lists/Sets use Aerospike Ordered Lists

**TTL and Expiration:**
- Redis: Key-level TTL, managed per key
- Aerospike: Record-level TTL, set via WritePolicy
- Both support automatic expiration

**Atomic Operations:**
- Redis: MULTI/EXEC for transactions, PIPELINE for batching
- Aerospike: Operate() for single-record atomicity, Batch operations for multiple records
- Note: Aerospike doesn't have multi-record transactions

### Operational Considerations

**Connection Management:**
- Redis: Manual connection pooling via go-redis client
- Aerospike: Built-in connection pooling, configured via ClientPolicy

**Performance Characteristics:**
- Aerospike typically faster for large-scale reads/writes (SSD-optimized)
- Sorted set queries may need optimization vs Redis native sorted sets
- Session operations should have similar performance (both in-memory)

**Migration Complexity:** Medium
- ✅ Pros: Clean interface abstraction, well-structured code
- ⚠️ Challenges: Sorted set migration requires careful design, type conversions needed

### Recommendations

1. **Run Detailed Data Modeling:** Execute `/map-data-models` command for granular Redis-to-Aerospike mappings
2. **Performance Baselines:** Capture current Redis performance before migration
3. **Leaderboard Strategy:** Recommend Secondary Index approach for scalability
4. **Phased Rollout:** Use dual-write mode to validate Aerospike in production before cutover
5. **Monitoring:** Set up Aerospike monitoring (latency, throughput, errors) before migration

### Next Steps

1. ✅ **Review this strategy document** - Validate approach and phases
2. ⏭️ **Execute `/map-data-models`** - Generate detailed data model mappings
3. ⏭️ **Set up Aerospike locally** - Install and configure for development
4. ⏭️ **Begin Phase 1** - Infrastructure setup and service skeleton
5. ⏭️ **Execute `/plan-iterate-go`** - Generate all implementation tasks and sync to Jira

---

**Document Status:** ✅ Ready for Review  
**Next Command:** `/map-data-models` for detailed data mappings

