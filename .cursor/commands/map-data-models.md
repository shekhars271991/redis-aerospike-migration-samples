# Output Specification for data-modeling-guidelines.md
# ------------------------------------------------------------
# The output document must follow this exact structure.
# Do not add or remove any sections beyond what is defined below.
# Maintain all headings and markdown formatting exactly as written.\

# Read through migration-strategy.md to make sure you understand the full picture of the task

# Expected Markdown Document Structure

# 1. HEADER SECTION
# ------------------------------------------------------------
# Title:
#   # Redis to Aerospike Data Modeling Guidelines
#
#
# Divider:
#   ---

# 2. SECTION 1: Redis Data Access and Key Usages
# ------------------------------------------------------------
# Heading:
#   ## 1. Redis Data Access and Key Usages
#
# Purpose:
#   This section provides a consolidated view of all Redis interactions
#   in the codebase, including the file location, command type,
#   key arguments, and the surrounding code snippet used for context.
#
# Content:
#   A markdown table with the following columns:
#     | File Path | Line Number | Redis Command | Key Argument | Code Snippet |
#
# Column definitions:
#   - File Path: Relative path of the file containing the Redis call.
#   - Line Number: Line number where the Redis call is found.
#   - Redis Command: The Redis operation used (e.g., get, set, hset, zadd).
#   - Key Argument: The Redis key or expression passed as the first argument.
#   - Code Snippet: The actual line or short code block where the Redis call appears.
#
# Each row represents one Redis call in the codebase.
# Do not summarize, group, or interpret entries.
# Include both static and dynamic key usages exactly as found.
#
# Example:
#
# ## 1. Redis Data Access and Key Usages
#
# | File Path | Line Number | Redis Command | Key Argument | Code Snippet |
# |------------|-------------|---------------|--------------|--------------|
# | src/cache/userStore.js | 24 | redis.set | "user:123" | `redis.set("user:123", userObj)` |
# | app/session.py | 88 | redis.get | f"session:{user_id}" | `session = redis.get(f"session:{user_id}")` |
#
# Notes:
#   - Code snippets should be fenced with backticks for readability.
#   - Do not attempt to normalize variable-based keys.
#   - The table must include all Redis commands found in the scan.

# 3. SECTION 2: Aerospike Mapping Template
# ------------------------------------------------------------
# Heading:
#   ## 2. Aerospike Mapping Template
#
# Purpose:
#   Define how Redis data types are mapped to Aerospike data models.
#   Use Aerospike Ordered Lists for Redis structures with multiple elements
#   (Lists, Sets, Sorted Sets).
#   Pub/Sub and unknown patterns should be skipped entirely.
#
# General Rules:
#   - Use one Aerospike record per Redis key.
#   - Keep field and bin names concise and aligned with Redis field names.
#   - Maintain one-to-one logical correspondence with Redis structures.
#   - Avoid schema redesign or introducing new entities.
#
# ---------------------------------------------------------------------
# Template 1: Redis String (includes TTL)
# ---------------------------------------------------------------------
# Description:
#   Simple key-value pairs (redis.set, redis.get, redis.setex, redis.expire)
#
# Example:
#   Redis Command:
#     redis.set("user:123", "ACTIVE")
#
#   Aerospike Mapping:
#     ### Redis Key: "user:123"
#     Aerospike Mapping:
#     - Namespace: users
#     - Set: user_status
#     - Record Key: userId
#     - Bins:
#       - status: ACTIVE
#     Notes:
#     - Store as a single-bin record.
#     - If Redis TTL is present (via setex or expire), record it in the Notes section.
#     - TTL is handled natively by Aerospike at the record level.
#
# ---------------------------------------------------------------------
# Template 2: Redis Hash
# ---------------------------------------------------------------------
# Description:
#   Dictionary-like structure (redis.hset, redis.hmget)
#
# Example:
#   Redis Command:
#     redis.hset("user:123", "email", "test@example.com")
#
#   Aerospike Mapping:
#     ### Redis Key: "user:123"
#     Aerospike Mapping:
#     - Namespace: users
#     - Set: profile
#     - Record Key: userId
#     - Bins:
#       - email: test@example.com
#       - name: <value>
#       - phone: <value>
#     Notes:
#     - Each Redis hash field maps to a separate bin.
#     - All fields remain under the same record.
#
# ---------------------------------------------------------------------
# Template 3: Redis List
# ---------------------------------------------------------------------
# Description:
#   Ordered sequence of elements (redis.lpush, redis.rpush, redis.lrange)
#   Use Aerospike Ordered Lists for maintaining rank-based order.
#
# Example:
#   Redis Command:
#     redis.lpush("notifications:123", "alert1")
#
#   Aerospike Mapping:
#     ### Redis Key: "notifications:123"
#     Aerospike Mapping:
#     - Namespace: users
#     - Set: notifications
#     - Record Key: userId
#     - Bins:
#       - alerts (Ordered List): ["alert1", "alert2", "alert3"]
#     Notes:
#     - Ordered Lists auto-sort on insertion and maintain rank order.
#     - Common operations: append, insert, get_by_rank, remove_by_rank.
#     - Limitations:
#       * Lists are bound by max-record-size.
#       * Disk-heavy workloads can slow writes.
#       * Lua UDFs cannot modify list bins.
#
# ---------------------------------------------------------------------
# Template 4: Redis Set
# ---------------------------------------------------------------------
# Description:
#   Unordered collection of unique values (redis.sadd, redis.smembers)
#   Represent as Aerospike Ordered List for deterministic ordering.
#
# Example:
#   Redis Command:
#     redis.sadd("user_roles:123", "admin")
#
#   Aerospike Mapping:
#     ### Redis Key: "user_roles:123"
#     Aerospike Mapping:
#     - Namespace: users
#     - Set: roles
#     - Record Key: userId
#     - Bins:
#       - roles (Ordered List): ["admin", "editor", "viewer"]
#     Notes:
#     - Use Ordered List to ensure stable ordering.
#     - Enforce uniqueness in application logic.
#     - Limitations:
#       * Large sets increase record size.
#       * Lists are limited by Aerospike’s max-record-size.
#
# ---------------------------------------------------------------------
# Template 5: Redis Sorted Set
# ---------------------------------------------------------------------
# Description:
#   Ordered collection with scores (redis.zadd, redis.zrange)
#   Use Aerospike Ordered List of maps (value, score).
#
# Example:
#   Redis Command:
#     redis.zadd("leaderboard:game1", 200, "player123")
#
#   Aerospike Mapping:
#     ### Redis Key: "leaderboard:game1"
#     Aerospike Mapping:
#     - Namespace: games
#     - Set: leaderboard
#     - Record Key: gameId
#     - Bins:
#       - scores (Ordered List): [
#           {"player": "player123", "score": 200},
#           {"player": "player456", "score": 150}
#         ]
#     Notes:
#     - Store as Ordered List sorted by score.
#     - Aerospike’s list ordering ensures rank consistency.
#     - Limitations:
#       * Mixed types ordered by type then value.
#       * Lua UDFs cannot update Ordered Lists.
#
# ---------------------------------------------------------------------
# Template 6: Redis JSON Data
# ---------------------------------------------------------------------
# Description:
#   RedisJSON data type for storing structured JSON documents.
#   Map directly to Aerospike Map and Ordered List bins.
#
# Example:
#   Redis Command:
#     JSON.SET "user:123" $ '{"name": "John", "age": 30, "roles": ["admin", "editor"]}'
#
#   Aerospike Mapping:
#     ### Redis Key: "user:123"
#     Aerospike Mapping:
#     - Namespace: users
#     - Set: profile_json
#     - Record Key: userId
#     - Bins:
#       - data (Map):
#           name: "John"
#           age: 30
#           roles (Ordered List): ["admin", "editor"]
#     Notes:
#     - Represent JSON objects as Aerospike Maps.
#     - Represent JSON arrays as Aerospike Ordered Lists.
#     - Nested JSON objects can be stored as nested Maps.
#     - Limitations:
#       * Record size limited by max-record-size.
#       * No schema enforcement for nested fields.
#       * Lua UDFs cannot modify deeply nested list/map structures.
#
# ---------------------------------------------------------------------
# Template 7: Redis Pub/Sub or Stream Data (Skip Mapping)
# ---------------------------------------------------------------------
# Description:
#   Transient or streaming data (redis.publish, redis.subscribe).
#   These are not persisted and should not be mapped.
#
# Example:
#   Redis Command:
#     redis.publish("alerts", "system_down")
#
#   Action:
#     - Do not create Aerospike mappings for Pub/Sub or stream keys.
#     - Mark these as "Skipped (Pub/Sub)" in documentation.
#
# ---------------------------------------------------------------------
# Template 8: Generic or Unknown Structure (Skip Mapping)
# ---------------------------------------------------------------------
# Description:
#   Keys that do not match recognized Redis data structures.
#
# Example:
#   Redis Command:
#     redis.customCommand("someKey", "value")
#
#   Action:
#     - Skip Aerospike mapping generation.
#     - Mark as "Skipped (Unrecognized Pattern)" for manual review.
#
# ---------------------------------------------------------------------
# End of Aerospike Mapping Templates
# ---------------------------------------------------------------------

# 4. SECTION 4: Validation Checklist
# ------------------------------------------------------------
# Heading:
#   ## 4. Validation Checklist
#
# Content:
#   The checklist must contain exactly the following five items:
#
#   - [ ] All Redis call sites captured
#   - [ ] Each Redis key mapped to an Aerospike model
#   - [ ] Bin definitions match existing Redis fields
#   - [ ] No structural changes unless approved
#   - [ ] Document reviewed by data modeling owner
#
# Example:
#
# ---
# ## 4. Validation Checklist
# - [ ] All Redis call sites captured
# - [ ] Each Redis key mapped to an Aerospike model
# - [ ] Bin definitions match existing Redis fields
# - [ ] No structural changes unless approved
# - [ ] Document reviewed by data modeling owner

# Formatting Rules
# ------------------------------------------------------------
# 1. All sections and headings must appear in the defined order.
# 2. Use markdown tables where specified.
# 3. Use fenced code snippets with backticks for Redis code examples.
# 4. No additional commentary or metadata should be added.
# 5. Keep heading levels and indentation consistent.
# 6. Always use the file name: data-modeling-guidelines.md
# 7. When regenerated, existing sections should be replaced in place.
# 8. Do not add timestamps, version info, or author information.
