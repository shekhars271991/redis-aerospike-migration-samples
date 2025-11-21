# =====================================================================
# migration-strategy-analysis.md
# =====================================================================
# PURPOSE
#   Scan the codebase to detect and classify Redis usage patterns, then
#   generate a concise `migration-strategy.md` describing the code-level
#   migration plan to Aerospike. 
#
#   **Absolutely no field-level or bin-level data modeling** is allowed.
#   All schema work is delegated exclusively to Step 2 (/map-data-models).
#
# OUTPUT
#   A single Markdown file named: `migration-strategy.md`
#
# GLOBAL RULES
#   - Markdown only.
#   - No Aerospike schema/bin/field definitions.
#   - No per-field mappings.
#   - Only code migration + architectural notes.
#   - Total document must remain ≤ 2,000 words.

# ---------------------------------------------------------------------
# REQUIRED SECTIONS (exact headings & order)
# ---------------------------------------------------------------------
# 1. # Database Migration Strategy: Redis → Aerospike
# 2. ## 1. Overview
# 3. ## 2. Impacted Files Summary
# 4. ## 3. Phased Migration Plan (High Level)
# 5. ## 4. Technical Notes (Operational & Architectural)
# 6. ## 5. Risks & Mitigations
# 7. ## Next Steps

# ---------------------------------------------------------------------
# REDIS SCAN RULES
# ---------------------------------------------------------------------
# Detect Redis usage using:
#   - Imports/Clients:
#       "import redis", "from redis", "Redis(", "redis.NewClient",
#       "createClient", "redis.StrictRedis"
#   - Commands:
#       set,get,setex,getset,incr,decr,
#       hset,hget,hgetall,hmset,hincrby,hdel,
#       lpush,rpush,lpop,rpop,lrange,llen,
#       sadd,smembers,srem,sismember,scard,
#       zadd,zrange,zrevrange,zrank,zscore,zincrby,
#       exists,del,expire,ttl,keys,scan,
#       pipeline,multi,exec,watch,
#       publish,subscribe,unsubscribe,
#       xadd,xread,xgroup   # Redis Streams

# ---------------------------------------------------------------------
# REQUIRED SCAN OUTPUT IN "Overview"
# ---------------------------------------------------------------------
# Overview must include:
#   - Total files with Redis usage.
#   - Breakdown by Redis structure (Strings, Hashes, Lists, Sets, Sorted Sets,
#     Streams, TTL, Pipelines, Pub/Sub).
#   - List of all Redis-using files (paths only).
#   - Architectural layer classification (DAO/Service/Cache/etc).
#   - Short migration complexity summary (≤ 6 bullets).

# ---------------------------------------------------------------------
# IMPACTED FILES SUMMARY (Section 2)
# ---------------------------------------------------------------------
# Three tables required:
#
# 1. Files to Update:
#    | File Path | Change Type | Why (one-line) |
#
# 2. Files to Create:
#    | File Path | Purpose | Notes (delegate to /map-data-models if schema) |
#
# 3. Files for Review:
#    | File Path | Reason for Review | Suggested Action |
#
# TOTAL rows across all three tables must be ≤ 12.

# ---------------------------------------------------------------------
# PHASED MIGRATION PLAN (Section 3)
# ---------------------------------------------------------------------
# High-level phases only. No field-level schemas.
# For each phase:
#   - Objective (≤ 40 words)
#   - Triggers (Redis patterns that activated this phase)
#   - Major implementation tasks (file/function-level)
#       * If any task requires schema/bin mapping → MUST include:
#             **DELEGATE: run /map-data-models for field-level mapping.**
#   - Exit criteria (3 bullets)
#
# Required phases:
#   - Foundation & Infrastructure
#   - Conditional phases (Strings, Hashes, Lists, Sets, Sorted Sets,
#     Streams, Pub/Sub, TTL, Pipelines/Atomicity)
#   - Cleanup & Decommission

# ---------------------------------------------------------------------
# TECHNICAL NOTES (Section 4)
# ---------------------------------------------------------------------
# MUST contain the following **three exact subsections**:
#
# 4.1 Unsupported or Limited Redis Features in Aerospike
#     - For each Redis capability detected:
#         * Label as:
#             "Supported"
#             "Supported with caveats"
#             "Not supported"
#         * One-line impact
#         * One-line mitigation
#
# 4.2 Major Architectural Changes Required
#     - For each unsupported/limited feature:
#         * Required architectural changes (e.g., "introduce Kafka for Pub/Sub")
#         * Affected files/components (from Section 2)
#         * Classification:
#             "Code change", "New service", or "Operational change"
#         * One-line migration impact
#
# 4.3 Recommendation Summary
#     - For each limitation, choose ONE:
#         * Replace with Aerospike-native pattern
#         * Implement application-level workaround
#         * Retain Redis for that workload (dual-read/write)
#         * Build an auxiliary service
#       Include one-line justification.
#
# STRICT: No per-field, per-bin, or schema definitions.

# ---------------------------------------------------------------------
# RISKS & MITIGATIONS (Section 5)
# ---------------------------------------------------------------------
# Table required:
#   | Risk | Probability | Impact | Mitigation |
# Max 8 rows.

# ---------------------------------------------------------------------
# NEXT STEPS (Section 6)
# ---------------------------------------------------------------------
# MUST contain exactly 3 bullets.
#
# One bullet MUST be:
#   - "Run `/map-data-models` to generate data-modeling-guidelines.md for keys flagged in Section 2."
#
# The other two must be concrete engineering actions.

# ---------------------------------------------------------------------
# FINAL JSON FOOTER (REQUIRED)
# ---------------------------------------------------------------------
# Document must end with:
#
# ```json
# {
#   "generated": "<DATE>",
#   "version": "1.0",
#   "status": "ready-for-review"
# }
# ```

# ---------------------------------------------------------------------
# VALIDATION RULES
# ---------------------------------------------------------------------
# If ANY of the following are missing or out of order → do NOT output the file.
# Instead output ONLY a fenced JSON error block describing the missing parts.
#
# Required validation:
#   - All required headings appear in exact order.
#   - Section 4 contains all three required subsections.
#   - All Redis features discovered are labeled correctly (Supported / Supported with caveats / Not supported).
#   - Delegate markers appear when schema is needed.
#
# ---------------------------------------------------------------------
# JIRA TASK CREATION (MANDATORY ON SUCCESS)
# ---------------------------------------------------------------------
# After successfully emitting migration-strategy.md, create a Jira task:
#
# project_key: "SCRUM"
# summary: "Review migration-strategy.md — Redis → Aerospike"
# issue_type: "Task"
# priority: "High"
# description must include:
#   - Reference to Sections 2 & 4
#   - Instruction to run `/map-data-models`
#   - Short checklist for reviewer (completeness, unsupported features handled, etc.)
# labels: ["migration", "redis", "aerospike", "strategy", "review"]
#
# Notification message:
#   "✅ Migration strategy generated. Jira task created for review."
#
# =====================================================================
# END OF PROMPT
# =====================================================================
