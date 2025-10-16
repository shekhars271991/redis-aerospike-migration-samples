# =====================================================================
# migration-strategy-analysis.md
# =====================================================================
# Step 0: Redis → Aerospike Migration Strategy Analysis
# =====================================================================
# Purpose:
#   Analyze the project codebase to identify files involved in Redis usage,
#   classify them by their architectural role, and determine which files
#   require modification or creation during the Aerospike migration.
#
#   The output must clearly specify:
#     - Files to be updated (existing Redis references)
#     - Files to be created (new Aerospike DAOs, migration scripts, tests)
#     - Optional or manual review files
#     - A short, consistent summary of migration steps and technical notes
#
# Output:
#   migration-strategy.md
# =====================================================================

scan_codebase_structure:
  include_extensions: [.py, .js, .ts, .go, .java]
  exclude_paths:
    - tests/
    - node_modules/
    - build/
    - dist/
    - venv/
    - target/
    - __pycache__/
  detect_patterns:
    # Redis usage indicators
    - "import redis"
    - "from redis"
    - "redis."
    - "redisClient"
    - "StrictRedis("
    - "Redis("
    - "hset"
    - "hget"
    - "set"
    - "setex"
    - "get"
    - "zadd"
    - "lpush"
    - "sadd"
    - "del"
    - "exists"
  classify_files:
    - Identify all files containing Redis client usage or references.
    - Categorize each as one of the following:
        * DAO / Repository layer (data access or persistence code)
        * Service / Business layer (mixed Redis + logic)
        * Utility / Cache helper layer (lightweight caching or ephemeral storage)
    - Exclude files with no Redis usage from the final output.
    - Detect any interface abstractions (DAO interfaces, repository contracts).
  output_to: migration-strategy.md
  output_format:
    - File Path
    - Observed Redis Usage (commands/functions)
    - Layer Type (DAO / Service / Utility / Unknown)
    - Architectural Notes

# ------------------------------------------------------------
# 2. Generate Impacted Files Summary
# ------------------------------------------------------------
# The model must only include relevant files — those that require
# modification, creation, or manual review. Each file should be placed
# under one of three tables with concise descriptions.
#
# Rules for consistency:
#   - Do NOT include files that have no Redis usage.
#   - Do NOT modify Redis test files; create new Aerospike test files instead.
#   - For tests: mark Redis test files as “Retain” and add new Aerospike tests.

generate_impacted_files_summary:
  input_file: migration-strategy.md
  replace_section: "## 2. Impacted Files Summary"
  structure:
    - Heading: "### A. Files to be Updated"
      Columns: ["File Path", "Change Type", "Description of Required Update"]
      Rules:
        * Include only necessary modifications (config, main, build files).
        * Redis service files should be marked **Deprecate**, not Modify.
        * Redis test files should NOT be modified here.
        * Keep descriptions concise and action-oriented.
    - Heading: "### B. Files to be Created"
      Columns: ["File Path", "Purpose", "Description"]
      Rules:
        * Include Aerospike DAO/service implementations.
        * Include **new test files** for Aerospike (`*_aerospike_test.go`).
        * Include migration scripts and setup documentation.
    - Heading: "### C. Optional Files (Manual Review)"
      Columns: ["File Path", "Reason for Review", "Suggested Action"]
      Rules:
        * Include interface definitions, shared models, or indirect dependencies.
        * Provide short notes on what to verify manually.

# ------------------------------------------------------------
# 3. Generate Migration Plan Summary
# ------------------------------------------------------------
# A concise, consistent, stepwise migration plan with no variation in structure.

generate_migration_plan:
  input_file: migration-strategy.md
  replace_section: "## 3. Migration Plan Summary"
  format: |
    ## 3. Migration Plan Summary
    1. Implement Aerospike DAO/service (`aerospike_service.go`).
    2. Create new Aerospike test suite (`aerospike_service_test.go`).
    3. Implement Redis → Aerospike data migration script.
    4. Update configuration to default to Aerospike.
    5. Replace Redis container with Aerospike in Docker setup.
    6. Validate functionality through integration tests.
    7. Remove Redis dependencies after successful validation.

# ------------------------------------------------------------
# 4. Generate Technical Notes
# ------------------------------------------------------------
# Always append consistent data model mapping and operational equivalence notes.

generate_technical_notes:
  append_to: migration-strategy.md
  section_title: "## 4. Technical Notes"
  include_details:
    - Redis Hash → Aerospike Record mapping
    - Redis Sorted Set → Aerospike Ordered List mapping
    - Redis String with TTL → Aerospike Record with TTL
    - Operational equivalents (batch, pipeline, expiration)
    - Known limitations (record size, type conversion, ordered list behavior)

# ------------------------------------------------------------
# 5. Enforce Consistent Output Format
# ------------------------------------------------------------
# Ensure output always follows the same layout and section order.

validate_output:
  ensure_sections_present:
    - "# Redis → Aerospike Migration Strategy Analysis"
    - "## 1. Overview"
    - "## 2. Impacted Files Summary"
    - "## 3. Migration Plan Summary"
    - "## 4. Technical Notes"
  ensure_section_order: true
  ensure_no_extra_sections: true
  ensure_tables_have_data: true
  ensure_overview_is_concise: true
  enforce_uniform_markdown_style: true

# ------------------------------------------------------------
# 6. Output Consistency Profile
# ------------------------------------------------------------
# Ensures that wording, section order, and style remain consistent
# across runs and repositories.

consistency_profile:
  tone: "Professional and concise"
  perspective: "Architectural summary for migration readiness"
  format_rules:
    - Always use numbered steps in Migration Plan Summary
    - Use tables for file changes (Updated, Created, Optional)
    - Exclude completed phase summaries or verbose narratives
    - Use same Markdown heading structure for every report
    - Use consistent terminology:
        * “Deprecate” for Redis files being removed
        * “Create” for new Aerospike equivalents
        * “Modify” for configuration/build files
  output_filename: migration-strategy.md

# =====================================================================
# END OF FILE
# =====================================================================
