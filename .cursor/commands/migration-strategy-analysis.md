# =====================================================================
# migration-strategy-analysis.md
# =====================================================================
# Database Migration Strategy Analysis (Redis → Aerospike)
# =====================================================================
# Purpose:
#   Analyze the project codebase to identify Redis usage patterns, classify
#   operations by data structure type, and generate a phased migration
#   strategy with comprehensive testing approach.
#
#   This command is GENERIC - it adapts to whatever Redis usage it finds
#   in the codebase, whether that's user management, caching, sessions,
#   leaderboards, queues, or any other use case.
#
# Output:
#   migration-strategy.md
# =====================================================================

# ------------------------------------------------------------
# Step 1: Discover Redis Usage Patterns
# ------------------------------------------------------------
# Scan the codebase to identify ALL Redis operations and classify
# them by data structure type and operation pattern.

analyze_redis_usage:
  scan_codebase:
    include_extensions: [.py, .js, .ts, .go, .java, .rb, .php]
    exclude_paths:
      - tests/
      - node_modules/
      - build/
      - dist/
      - venv/
      - vendor/
      - target/
      - __pycache__/
  
  detect_patterns:
    # Redis client imports and initialization
    - "import redis"
    - "from redis"
    - "Redis("
    - "redis.NewClient"
    - "createClient"
    
    # Data structure operations
    redis_strings: ["set", "get", "setex", "getset", "incr", "decr"]
    redis_hashes: ["hset", "hget", "hgetall", "hmset", "hincrby", "hdel"]
    redis_lists: ["lpush", "rpush", "lpop", "rpop", "lrange", "llen"]
    redis_sets: ["sadd", "smembers", "srem", "sismember", "scard"]
    redis_sorted_sets: ["zadd", "zrange", "zrevrange", "zrank", "zscore", "zincrby"]
    redis_keys: ["exists", "del", "expire", "ttl", "keys", "scan"]
    redis_transactions: ["pipeline", "multi", "exec", "watch"]
    redis_pubsub: ["publish", "subscribe", "unsubscribe"]
  
  classify_by_structure:
    # For each file found, identify which Redis data structures are used
    # This drives the phasing strategy
    - String operations (simple key-value, counters, flags)
    - Hash operations (structured data, records)
    - List operations (queues, timelines, activity feeds)
    - Set operations (unique collections, tags, relationships)
    - Sorted Set operations (rankings, leaderboards, time-series)
    - TTL/Expiration patterns (sessions, caches, temporary data)
    - Transaction patterns (atomic operations, pipelines)
    - Pub/Sub patterns (event streaming - not migrated)
  
  classify_by_layer:
    # Architectural classification
    - DAO/Repository layer (direct data access)
    - Service/Business layer (Redis + business logic)
    - Cache layer (ephemeral storage)
    - Session management (user sessions)
    - Queue/Job processing (async tasks)
    - Utility/Helper layer (shared Redis functions)
  
  output_analysis:
    - Total Redis files identified
    - Data structures used (breakdown by type)
    - Operation complexity (read-only, write-heavy, mixed)
    - Transaction usage (pipeline, atomic operations)
    - TTL usage patterns
    - Interface abstractions detected (if any)

# ------------------------------------------------------------
# Step 2: Generate Impacted Files Summary
# ------------------------------------------------------------
# Create three categorized tables of files that need attention

generate_impacted_files:
  section: "## 2. Impacted Files Summary"
  
  files_to_update:
    include:
      - Configuration files (environment, settings)
      - Main/entry point files (initialization)
      - Build/deployment files (dependencies, containers)
      - Interface definitions (if abstraction exists)
    exclude:
      - Test files (will create new ones instead)
      - Source service files (will deprecate, not modify)
    columns: ["File Path", "Change Type", "Description"]
  
  files_to_create:
    include:
      - New Aerospike service/DAO implementation
      - New test suites for Aerospike
      - Data migration scripts
      - Validation scripts
      - Deployment/setup documentation
    columns: ["File Path", "Purpose", "Description"]
  
  files_for_review:
    include:
      - Interface contracts (verify compatibility)
      - Data models (verify field compatibility)
      - Indirect dependencies
      - Configuration templates
    columns: ["File Path", "Reason for Review", "Suggested Action"]

# ------------------------------------------------------------
# Step 3: Generate Dynamic Phased Migration Plan
# ------------------------------------------------------------
# Create phases dynamically based on discovered Redis usage.
# Each phase focuses on a specific data structure or operation pattern.

generate_phased_plan:
  section: "## 3. Phased Migration Plan with Testing Strategy"
  
  introduction: |
    ### Migration Approach
    This migration follows a phased approach where each phase:
    1. **Pre-Migration Tests**: Establish baseline behavior with Redis
    2. **Implementation**: Build Aerospike equivalent functionality  
    3. **Post-Migration Tests**: Verify Aerospike matches Redis behavior
    4. **Validation Tests**: Compare outputs, performance, and data integrity
    5. **Rollback Plan**: Clear path to revert if issues arise
    
    Each phase is independently testable and can be rolled back without
    affecting other phases.
  
  # Phase generation rules - adapt based on what's found
  phase_generation_rules:
    always_include:
      - phase: "Foundation & Infrastructure"
        triggers: [always]
        objective: "Set up Aerospike infrastructure and basic connectivity"
        includes:
          - Dependency installation
          - Configuration setup
          - Container/deployment setup
          - Basic service initialization
          - Health check implementation
        pre_tests:
          - Run existing test suite (baseline)
          - Document performance metrics
          - Verify test coverage
        post_tests:
          - Connection establishment
          - Health check validation
          - Configuration parsing
        validation:
          - Infrastructure stability
          - Configuration correctness
        success_criteria:
          - Service initializes without errors
          - Health check passes
        rollback: "Remove Aerospike config, continue Redis-only"
    
    conditional_phases:
      - phase: "Simple Key-Value Operations"
        triggers: [has_string_operations]
        data_structures: ["Redis Strings"]
        operations: ["GET", "SET", "SETEX", "INCR", "DECR"]
        objective: "Migrate simple key-value storage"
        aerospike_mapping: "String → Single-bin Record"
        
      - phase: "Structured Data (Hash-based)"
        triggers: [has_hash_operations]
        data_structures: ["Redis Hashes"]
        operations: ["HSET", "HGET", "HGETALL", "HINCRBY"]
        objective: "Migrate hash-based structured data"
        aerospike_mapping: "Hash → Multi-bin Record"
        
      - phase: "Ordered Collections (Lists)"
        triggers: [has_list_operations]
        data_structures: ["Redis Lists"]
        operations: ["LPUSH", "RPUSH", "LRANGE", "LPOP"]
        objective: "Migrate list-based ordered collections"
        aerospike_mapping: "List → Ordered List bin"
        
      - phase: "Unique Collections (Sets)"
        triggers: [has_set_operations]
        data_structures: ["Redis Sets"]
        operations: ["SADD", "SMEMBERS", "SREM"]
        objective: "Migrate set-based unique collections"
        aerospike_mapping: "Set → Ordered List bin (with uniqueness)"
        
      - phase: "Ranked Collections (Sorted Sets)"
        triggers: [has_sortedset_operations]
        data_structures: ["Redis Sorted Sets"]
        operations: ["ZADD", "ZRANGE", "ZREVRANGE", "ZRANK"]
        objective: "Migrate sorted sets with scores/rankings"
        aerospike_mapping: "Sorted Set → Ordered List with score tracking or Secondary Index"
        complexity: "HIGH"
        note: "Most complex migration - requires careful design"
        
      - phase: "TTL & Expiration Patterns"
        triggers: [has_ttl_operations]
        data_structures: ["Any with TTL"]
        operations: ["SETEX", "EXPIRE", "TTL"]
        objective: "Migrate time-based expiration"
        aerospike_mapping: "TTL → Record-level expiration (WritePolicy)"
        
      - phase: "Atomic Operations & Transactions"
        triggers: [has_pipeline_or_multi]
        data_structures: ["Multiple"]
        operations: ["PIPELINE", "MULTI/EXEC", "WATCH"]
        objective: "Ensure atomic multi-operation consistency"
        aerospike_mapping: "Pipeline → Batch operations, Multi → Operate()"
        
      - phase: "Data Migration & Dual-Write"
        triggers: [always]
        objective: "Migrate historical data and enable parallel writes"
        includes:
          - Snapshot existing Redis data
          - Batch migration script
          - Dual-write implementation
          - Data consistency validation
        
      - phase: "Integration Testing & Cutover"
        triggers: [always]
        objective: "Full system testing and production switchover"
        includes:
          - End-to-end testing
          - Load testing
          - Performance validation
          - Primary database switch
        
      - phase: "Cleanup & Decommission"
        triggers: [always]
        objective: "Remove Redis dependencies"
        includes:
          - Remove old service code
          - Remove dependencies
          - Update documentation
  
  # Testing template for each phase
  testing_template_per_phase:
    pre_migration_tests:
      - "Capture baseline behavior from Redis"
      - "Document current performance metrics"
      - "Record expected outputs for validation"
      - "Verify test coverage for operations"
    
    implementation_steps:
      - "List specific methods/functions to implement"
      - "Define data structure mappings"
      - "Specify error handling requirements"
    
    post_migration_tests:
      - "Test Aerospike implementation in isolation"
      - "Verify data structure compatibility"
      - "Test error handling and edge cases"
      - "Validate type conversions if any"
    
    validation_tests:
      - "Side-by-side comparison (Redis vs Aerospike)"
      - "Performance benchmarking"
      - "Data integrity checks"
      - "Load testing if applicable"
    
    success_criteria:
      - "Functional parity with Redis"
      - "Performance within acceptable range"
      - "No data loss or corruption"
      - "All tests passing"
    
    rollback_plan:
      - "Clear steps to revert changes"
      - "No impact on previous phases"

# ------------------------------------------------------------
# Step 4: Generate Technical Notes & Considerations
# ------------------------------------------------------------
# High-level technical notes about the migration
# (Detailed data model mappings handled by separate map-data-models command)

generate_technical_notes:
  section: "## 4. Technical Notes"
  
  include_overview:
    - Summary of Redis data structures discovered
    - High-level Aerospike equivalents overview
    - Reference to detailed data modeling document
    - Key architectural differences to consider
  
  operational_considerations:
    - Connection management and pooling
    - Transaction and atomicity patterns
    - Performance characteristics comparison
    - Migration complexity assessment
  
  key_differences:
    - Type system (Redis strings vs Aerospike native types)
    - Record size limitations (1MB default, 8MB max)
    - Data structure equivalents (hashes, lists, sorted sets)
    - TTL and expiration handling
    - Batch operations and pipelines
  
  recommendations:
    - Run detailed data modeling analysis (use map-data-models command)
    - Establish performance baselines before migration
    - Plan for data validation and reconciliation
    - Consider phased rollout strategy
  
  next_steps:
    - Execute map-data-models command for detailed mappings
    - Review data model compatibility
    - Validate Aerospike schema design
    - Create data migration scripts

# ------------------------------------------------------------
# Step 5: Enforce Consistent Output Format
# ------------------------------------------------------------

validate_output:
  required_sections:
    - "# Database Migration Strategy: Redis → Aerospike"
    - "## 1. Overview"
    - "## 2. Impacted Files Summary"
    - "## 3. Phased Migration Plan with Testing Strategy"
    - "## 4. Technical Notes"
  
  overview_requirements:
    - Current architecture summary
    - Redis usage statistics (files, operations, data structures)
    - Migration approach summary
    - Complexity assessment
  
  consistency_rules:
    - Professional and concise tone
    - Numbered steps in migration plan
    - Tables for file changes
    - Code examples in fenced blocks
    - Clear success criteria for each phase
    - Explicit rollback plans

# ------------------------------------------------------------
# Step 6: Create Jira Validation Task
# ------------------------------------------------------------

create_jira_task:
  trigger: "After migration-strategy.md is successfully generated"
  project_key: "SCRUM"
  task_details:
    summary: "Review and validate Redis → Aerospike migration strategy"
    issue_type: "Task"
    priority: "High"
    description: |
      A migration strategy document has been generated and requires validation.
      
      **Document:** `migration-strategy.md`
      
      **Strategy Validation Checklist:**
      - [ ] All Redis usage patterns identified
      - [ ] File classification is accurate (DAO/Service/Cache)
      - [ ] Files to update/create lists are complete
      - [ ] Migration phases are logical and well-defined
      - [ ] Each phase has clear pre/post testing strategy
      - [ ] Data model mappings are technically sound
      - [ ] Rollback plans are feasible
      - [ ] Success criteria are measurable
      - [ ] Performance expectations are realistic
      - [ ] No missing dependencies or edge cases
      
      **Key Review Areas:**
      - **Section 1:** Overview - verify completeness
      - **Section 2:** Impacted files - validate all files identified
      - **Section 3:** Phased plan - review phase breakdown and testing
      - **Section 4:** Technical notes - review high-level considerations
      
      **Next Steps:**
      - Once strategy approved, run /map-data-models for detailed mappings
      - Create phase-specific implementation tasks
      - Set up test baseline (Phase 1 pre-tests)
      - Begin infrastructure setup
      
      **Generated:** {timestamp}
    labels: ["migration", "redis", "aerospike", "strategy", "review"]
  
  notification:
    message: |
      ✅ Migration strategy analysis complete!
      📄 Document: migration-strategy.md
      🎫 Jira task created: {issue_key}
      🔗 {issue_url}
      
      📊 Review the phased migration plan with testing strategy
      🧪 Each phase includes pre/post tests and validation

# =====================================================================
# END OF FILE
# =====================================================================
