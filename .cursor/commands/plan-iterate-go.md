# =====================================================================
# plan-iterate-go.md
# =====================================================================
# Purpose:
#   Generate a complete, sequenced task plan for Redis → Aerospike migration
#   by analyzing the codebase and creating all implementation and testing tasks.
#   
#   This command creates ALL tasks upfront and syncs them to Jira, then saves
#   a local task sequence file that code-iterate-go will use to execute tasks
#   one by one based on Jira status.
#
# Inputs:
#   - migration-strategy.md (from migration-strategy-analysis)
#   - data-modeling-guidelines.md (from map-data-models)
#   - Codebase analysis (files, Redis usage, interfaces)
#
# Outputs:
#   - task-sequence.json (local task order and metadata)
#   - Jira issues (all tasks created via MCP)
#   - task-plan.md (human-readable task breakdown)
#
# Workflow:
#   1. Read strategy and data model documents
#   2. Analyze codebase to identify specific work items
#   3. Generate granular tasks organized by phase
#   4. Create pre/post testing tasks for each phase
#   5. Sync all tasks to Jira via MCP
#   6. Save task sequence locally for execution tracking
# =====================================================================

# ------------------------------------------------------------
# Step 1: Read Input Documents
# ------------------------------------------------------------
read_strategy_and_models:
  required_inputs:
    - file: "migration-strategy.md"
      sections_to_extract:
        - "## 1. Overview"
        - "## 2. Impacted Files Summary"
        - "## 3. Phased Migration Plan with Testing Strategy"
        - "## 4. Technical Notes"
      extract:
        - phases: List of migration phases identified
        - files_to_update: Files that need modification
        - files_to_create: New files to implement
        - data_structures_used: Redis data structures in use
    
    - file: "data-modeling-guidelines.md"
      sections_to_extract:
        - "## 1. Redis Data Access and Key Usages"
        - "## 2. Aerospike Mapping Template"
      extract:
        - redis_operations: All Redis operations found
        - key_patterns: Redis key patterns used
        - aerospike_mappings: How each pattern maps to Aerospike
  
  validation:
    - Ensure both files exist and are readable
    - Verify phases are defined in migration-strategy.md
    - Verify Redis operations are documented in data-modeling-guidelines.md

# ------------------------------------------------------------
# Step 2: Analyze Codebase for Detailed Task Generation
# ------------------------------------------------------------
analyze_codebase:
  scan_for:
    - Redis service files (identify all methods to migrate)
    - Configuration files (identify all settings to update)
    - Test files (identify test coverage and gaps)
    - Interface definitions (verify abstraction layer)
    - Docker/deployment files (identify container configs)
    - Build files (go.mod, Makefile, package.json, etc.)
  
  extract_details:
    - method_signatures: List all Redis service methods
    - redis_operations_by_method: Map methods to Redis commands
    - test_coverage: Which methods have tests
    - dependencies: External libraries and versions
    - configuration_keys: Environment variables and config keys
  
  classify_complexity:
    - Simple (CRUD operations): Low complexity
    - Medium (with transactions): Medium complexity
    - Complex (sorted sets, complex queries): High complexity

# ------------------------------------------------------------
# Step 3: Generate Task Breakdown by Phase
# ------------------------------------------------------------
generate_task_breakdown:
  task_generation_rules:
    # Each phase generates multiple tasks:
    # - Pre-migration test tasks
    # - Implementation tasks (broken down by method/feature)
    # - Post-migration test tasks
    # - Validation tasks
  
  for_each_phase:
    phase_structure:
      # Example for Phase 2: User Management
      pre_migration_tests:
        - task_type: "Pre-Migration Test"
          title: "Establish Redis baseline for {phase_name}"
          description: |
            Run existing tests and capture baseline metrics
            - Execute existing test suite
            - Document performance (latency, throughput)
            - Capture sample outputs
            - Verify test coverage
          estimated_effort: "1-2 hours"
          dependencies: []
          test_type: "baseline"
      
      implementation_tasks:
        # Break down by method/feature
        - task_type: "Implementation"
          title: "Implement {method_name} in AerospikeService"
          description: |
            Implement the Aerospike equivalent for {redis_method}
            
            **Redis Operation:** {redis_operations}
            **Aerospike Mapping:** {aerospike_mapping}
            **Data Structure:** {data_structure}
            
            **Implementation Steps:**
            1. Add method to AerospikeService
            2. Implement data structure mapping
            3. Handle error cases
            4. Add inline documentation
            
            **Files to Modify:**
            - {service_file}
            
            **Reference:**
            - See data-modeling-guidelines.md section {section}
          estimated_effort: "{complexity_hours}"
          dependencies: ["Previous implementation task"]
          complexity: "{simple|medium|high}"
      
      post_migration_tests:
        - task_type: "Post-Migration Test"
          title: "Test {method_name} with Aerospike"
          description: |
            Verify Aerospike implementation matches Redis behavior
            - Test basic functionality
            - Test edge cases and error handling
            - Verify data structure compatibility
            - Test type conversions
          estimated_effort: "1-2 hours"
          dependencies: ["Implementation of {method_name}"]
          test_type: "functional"
      
      validation_tasks:
        - task_type: "Validation"
          title: "Validate {phase_name} - Compare Redis vs Aerospike"
          description: |
            Side-by-side comparison and validation
            - Run same operations on both systems
            - Compare outputs (must match exactly)
            - Performance benchmarking
            - Data integrity checks
            
            **Success Criteria:**
            - Outputs match 100%
            - Performance within acceptable range
            - No data loss or corruption
          estimated_effort: "2-3 hours"
          dependencies: ["All implementation tasks in phase"]
          test_type: "validation"

# ------------------------------------------------------------
# Step 4: Generate Specific Tasks from Codebase Analysis
# ------------------------------------------------------------
generate_concrete_tasks:
  # For each discovered Redis method, create implementation task
  for_each_method:
    template:
      task_id: "IMPL-{phase_number}-{sequence}"
      title: "Implement {method_name} in AerospikeService"
      description: |
        Migrate {method_name} from RedisService to AerospikeService
        
        **Current Redis Implementation:**
        File: {redis_file}:{line_number}
        Method: {method_signature}
        Operations: {redis_commands_used}
        
        **Target Aerospike Implementation:**
        File: {aerospike_file}
        Mapping: {aerospike_mapping_reference}
        
        **Acceptance Criteria:**
        - Method signature matches DBService interface
        - Handles all error cases
        - Type conversions are correct
        - Inline documentation added
        
        **Testing:**
        - Unit tests pass
        - Integration tests pass
      phase: "{phase_name}"
      dependencies: ["{previous_task_ids}"]
      estimated_hours: "{hours}"
      labels: ["implementation", "phase-{N}", "{data_structure_type}"]
  
  # For each configuration change, create config task
  for_each_config:
    template:
      task_id: "CONFIG-{sequence}"
      title: "Add Aerospike configuration to {config_file}"
      description: |
        Update configuration to support Aerospike
        
        **Changes Required:**
        - Add AEROSPIKE_HOST environment variable
        - Add AEROSPIKE_PORT environment variable
        - Add AEROSPIKE_NAMESPACE configuration
        - Update DB_TYPE to support "aerospike"
        
        **Files to Modify:**
        {config_files}
      labels: ["configuration", "phase-1"]
  
  # For each test file, create test task
  for_each_test_needed:
    template:
      task_id: "TEST-{phase_number}-{sequence}"
      title: "{pre|post}-Migration Test: {test_name}"
      description: |
        {test_description}
        
        **Test Scope:**
        {test_scope}
        
        **Success Criteria:**
        {success_criteria}
      test_type: "{pre|post|validation}"
      labels: ["testing", "phase-{N}"]

# ------------------------------------------------------------
# Step 5: Organize Tasks into Sequence
# ------------------------------------------------------------
organize_task_sequence:
  sequencing_rules:
    # Tasks must be ordered such that:
    # 1. Dependencies are satisfied
    # 2. Pre-tests come before implementation
    # 3. Implementation comes before post-tests
    # 4. All phase tasks complete before next phase
  
  phase_order:
    - phase: 1
      name: "Foundation & Infrastructure"
      tasks:
        - PRE-TEST-1-1: "Establish Redis baseline"
        - CONFIG-1: "Add Aerospike dependencies"
        - CONFIG-2: "Configure Aerospike settings"
        - IMPL-1-1: "Create AerospikeService skeleton"
        - IMPL-1-2: "Implement HealthCheck()"
        - POST-TEST-1-1: "Test Aerospike connection"
        - POST-TEST-1-2: "Test HealthCheck()"
        - VALIDATE-1: "Validate infrastructure setup"
    
    - phase: 2
      name: "User Management"
      tasks:
        - PRE-TEST-2-1: "Capture Redis user CRUD baseline"
        - IMPL-2-1: "Implement CreateUser()"
        - POST-TEST-2-1: "Test CreateUser() with Aerospike"
        - IMPL-2-2: "Implement GetUser()"
        - POST-TEST-2-2: "Test GetUser() with Aerospike"
        - IMPL-2-3: "Implement UpdateScore()"
        - POST-TEST-2-3: "Test UpdateScore() with Aerospike"
        - VALIDATE-2: "Compare user operations Redis vs Aerospike"
    
    # ... continue for all phases
  
  dependency_graph:
    # Create dependency graph for visualization
    # Ensures no circular dependencies
    # Allows parallel execution where possible

# ------------------------------------------------------------
# Step 6: Create All Tasks in Jira via MCP
# ------------------------------------------------------------
create_jira_tasks:
  connection: "mcp-atlassian"
  project_key: "SCRUM"
  
  batch_create_strategy:
    # Create tasks in batches to avoid overwhelming Jira
    batch_size: 10
    rate_limit: "1 batch per 2 seconds"
  
  task_creation:
    for_each_task:
      jira_issue:
        project: "SCRUM"
        issue_type: "Task"
        summary: "{task.title}"
        description: |
          {task.description}
          
          **Phase:** {task.phase}
          **Estimated Effort:** {task.estimated_hours} hours
          **Task Type:** {task.task_type}
          
          **Dependencies:**
          {task.dependencies}
          
          **Acceptance Criteria:**
          {task.acceptance_criteria}
          
          **Reference Documents:**
          - migration-strategy.md
          - data-modeling-guidelines.md
        priority: "{High for blocking tasks, Medium otherwise}"
        labels: [
          "migration",
          "redis-aerospike",
          "phase-{phase_number}",
          "{task.task_type}",
          "{task.complexity}"
        ]
        custom_fields:
          task_sequence: "{task_id}"
          phase: "{phase_name}"
          estimated_hours: "{hours}"
  
  link_dependencies:
    # After all tasks created, link them in Jira
    for_each_dependency:
      create_issue_link:
        type: "Blocks"
        inward_issue: "{dependency_task_jira_key}"
        outward_issue: "{current_task_jira_key}"
  
  create_epic:
    # Optional: Create an Epic to group all migration tasks
    epic_task:
      summary: "Redis → Aerospike Migration"
      description: |
        Complete migration from Redis to Aerospike
        
        **Total Tasks:** {total_task_count}
        **Phases:** {phase_count}
        **Estimated Duration:** {total_hours} hours
        
        **Documents:**
        - migration-strategy.md
        - data-modeling-guidelines.md
        - task-plan.md
      issue_type: "Epic"
    
    link_all_tasks_to_epic:
      # Link all generated tasks to this epic

# ------------------------------------------------------------
# Step 7: Save Task Sequence Locally
# ------------------------------------------------------------
save_task_sequence:
  output_file: "task-sequence.json"
  format: |
    {
      "migration_name": "Redis to Aerospike",
      "generated_at": "{timestamp}",
      "total_tasks": {count},
      "total_phases": {phase_count},
      "estimated_hours": {total_hours},
      "jira_project": "SCRUM",
      "jira_epic": "{epic_key}",
      "current_task_index": 0,
      "tasks": [
        {
          "task_id": "PRE-TEST-1-1",
          "sequence": 1,
          "jira_key": "SCRUM-10",
          "title": "Establish Redis baseline",
          "phase": 1,
          "phase_name": "Foundation & Infrastructure",
          "task_type": "pre-migration-test",
          "status": "To Do",
          "dependencies": [],
          "estimated_hours": 2,
          "files_to_modify": [],
          "files_to_create": [],
          "acceptance_criteria": [...]
        },
        {
          "task_id": "CONFIG-1",
          "sequence": 2,
          "jira_key": "SCRUM-11",
          "title": "Add Aerospike dependencies",
          "phase": 1,
          "phase_name": "Foundation & Infrastructure",
          "task_type": "configuration",
          "status": "To Do",
          "dependencies": ["PRE-TEST-1-1"],
          "estimated_hours": 1,
          "files_to_modify": ["go.mod", "Makefile"],
          "files_to_create": [],
          "acceptance_criteria": [...]
        },
        // ... all tasks in sequence
      ]
    }

# ------------------------------------------------------------
# Step 8: Generate Human-Readable Task Plan
# ------------------------------------------------------------
generate_task_plan_document:
  output_file: "task-plan.md"
  format: |
    # Migration Task Plan: Redis → Aerospike
    
    **Generated:** {timestamp}
    **Total Tasks:** {total_tasks}
    **Total Phases:** {phase_count}
    **Estimated Duration:** {total_hours} hours
    **Jira Epic:** [{epic_key}]({epic_url})
    
    ---
    
    ## Task Sequence Overview
    
    | Seq | Task ID | Jira | Title | Phase | Type | Effort | Dependencies |
    |-----|---------|------|-------|-------|------|--------|--------------|
    {for_each_task}
    | {seq} | {task_id} | [{jira_key}]({jira_url}) | {title} | {phase} | {type} | {hours}h | {deps} |
    
    ---
    
    ## Phase Breakdown
    
    {for_each_phase}
    ### Phase {number}: {phase_name}
    
    **Objective:** {objective}
    **Total Tasks:** {phase_task_count}
    **Estimated Duration:** {phase_hours} hours
    
    #### Tasks in This Phase:
    {for_each_task_in_phase}
    - [{task_id}] {title} ({hours}h) - [{jira_key}]({jira_url})
    
    ---
    
    ## Execution Instructions
    
    1. Tasks will be executed in sequence using code-iterate-go command
    2. code-iterate-go checks Jira status via MCP before executing each task
    3. A task only executes when:
       - All dependency tasks are "Done" in Jira
       - Previous task in sequence is "Done" in Jira
    4. After completing a task, update Jira status to "Done"
    5. code-iterate-go will automatically pick up the next task
    
    ---
    
    ## Task Dependencies Visualization
    
    ```
    {generate_ascii_dependency_graph}
    ```

# ------------------------------------------------------------
# Step 9: Validation and Summary
# ------------------------------------------------------------
validate_and_summarize:
  validations:
    - Verify all phases have tasks
    - Verify all dependencies are valid (no circular deps)
    - Verify all Jira tasks were created successfully
    - Verify task-sequence.json is valid JSON
    - Verify task-plan.md is readable
  
  summary_output: |
    ✅ Migration Task Plan Generated Successfully!
    
    📊 **Statistics:**
    - Total Tasks: {total_tasks}
    - Total Phases: {phase_count}
    - Estimated Duration: {total_hours} hours ({days} days)
    
    📄 **Generated Files:**
    - task-sequence.json (for code-iterate-go execution)
    - task-plan.md (human-readable breakdown)
    
    🎫 **Jira Integration:**
    - Epic Created: {epic_key}
    - Tasks Created: {task_count}
    - All tasks linked with dependencies
    - Project: {jira_project}
    
    📋 **Task Type Breakdown:**
    - Pre-Migration Tests: {pre_test_count}
    - Implementation Tasks: {impl_count}
    - Post-Migration Tests: {post_test_count}
    - Validation Tasks: {validate_count}
    - Configuration Tasks: {config_count}
    
    🔗 **View in Jira:**
    {jira_epic_url}
    
    ▶️ **Next Steps:**
    1. Review task-plan.md for task breakdown
    2. Review Jira epic and tasks
    3. Run code-iterate-go to start executing tasks in sequence
    4. code-iterate-go will check Jira status and execute tasks one by one

# =====================================================================
# END OF FILE
# =====================================================================
