# =====================================================================
# code-iterate-go.md
# =====================================================================
# Purpose:
#   Execute migration tasks sequentially by reading task-sequence.json,
#   checking Jira status via MCP, and implementing the next ready task.
#
#   This command finds the next task where:
#   - All dependencies are "Done" in Jira
#   - Task status is "To Do" in Jira
#   - Then implements the code changes
#
#   User manually validates changes and updates Jira status.
#   Next run picks up the next task automatically.
#
# Inputs:
#   - task-sequence.json (from plan-iterate-go)
#   - Jira status via MCP
#   - migration-strategy.md
#   - data-modeling-guidelines.md
#
# Outputs:
#   - Code changes (files created/modified)
#   - execution-log.md (local execution history)
#
# Workflow:
#   1. Read task-sequence.json
#   2. Check Jira status for each task via MCP
#   3. Find next task ready for execution
#   4. Implement code changes for that task
#   5. Log execution locally
#   6. Wait for user to validate and update Jira
#   7. Next run continues with next task
# =====================================================================

# ------------------------------------------------------------
# Step 1: Read Task Sequence File
# ------------------------------------------------------------
read_task_sequence:
  input_file: "task-sequence.json"
  
  validation:
    - File must exist (run plan-iterate-go first)
    - JSON must be valid
    - Must contain "tasks" array
    - Each task must have: task_id, jira_key, status, dependencies
  
  parse_structure:
    - migration_name: Project name
    - jira_project: Jira project key
    - jira_epic: Epic key linking all tasks
    - current_task_index: Last executed task index
    - tasks: Array of all tasks in sequence
  
  on_error:
    message: |
      ❌ task-sequence.json not found or invalid
      
      Please run /plan-iterate-go first to generate task sequence
      and create Jira tasks.

# ------------------------------------------------------------
# Step 2: Check Jira Status for All Tasks via MCP
# ------------------------------------------------------------
sync_jira_status:
  mcp_connection: "mcp-atlassian"
  
  for_each_task:
    mcp_call: "jira_get_issue"
    parameters:
      issue_key: "{task.jira_key}"
      fields: "status,summary"
    
    extract:
      - current_status: Status name (To Do, In Progress, Done, etc.)
      - summary: Task title
    
    update_local_cache:
      # Update task-sequence.json with current Jira status
      task.status: "{current_status}"
      task.last_synced: "{timestamp}"
  
  summary:
    - Total tasks: {count}
    - Done: {done_count}
    - In Progress: {in_progress_count}
    - To Do: {todo_count}
    - Blocked: {blocked_count}

# ------------------------------------------------------------
# Step 3: Find Next Task Ready for Execution
# ------------------------------------------------------------
find_next_task:
  logic:
    # Iterate through tasks in sequence order
    for_each_task_in_sequence:
      check_conditions:
        1. Skip if status is "Done"
        2. Skip if status is "In Progress" (already being worked on)
        3. Check if all dependencies are "Done":
           for_each_dependency in task.dependencies:
             find_dependency_task_by_id:
               if dependency_task.status != "Done":
                 mark_as: "blocked"
                 reason: "Waiting for {dependency_task.jira_key}"
                 continue_to_next_task: true
        4. If all dependencies done AND status is "To Do":
           select_as: "next_executable_task"
           break: true
  
  output:
    - next_task: Task object to execute
    - blocked_tasks: List of tasks waiting on dependencies
    - reason: Why this task was selected
  
  no_task_found_scenarios:
    - all_done: "🎉 All tasks complete! Migration finished."
    - all_blocked: "⚠️ No tasks ready. Check Jira for blocking issues."
    - in_progress: "⏳ Task {jira_key} is In Progress. Complete and update Jira first."

# ------------------------------------------------------------
# Step 4: Display Task Information
# ------------------------------------------------------------
display_task_info:
  input: next_task
  
  output: |
    ╔════════════════════════════════════════════════════════════════╗
    ║  NEXT TASK TO EXECUTE                                          ║
    ╚════════════════════════════════════════════════════════════════╝
    
    📋 Task ID: {task.task_id}
    🎫 Jira: {task.jira_key} - {jira_url}
    📌 Title: {task.title}
    🏷️  Phase: {task.phase} - {task.phase_name}
    📂 Type: {task.task_type}
    ⏱️  Estimated: {task.estimated_hours} hours
    
    📄 Files to Modify:
    {for_each file in task.files_to_modify:
      - {file}
    }
    
    📝 Files to Create:
    {for_each file in task.files_to_create:
      - {file}
    }
    
    ✅ Dependencies: {dependency_status}
    {for_each dep in task.dependencies:
      ✓ {dep.task_id} ({dep.jira_key}) - Done
    }
    
    📖 Description:
    {task.description}
    
    🎯 Acceptance Criteria:
    {for_each criteria in task.acceptance_criteria:
      - {criteria}
    }
    
    ────────────────────────────────────────────────────────────────

# ------------------------------------------------------------
# Step 5: Read Reference Documents
# ------------------------------------------------------------
read_reference_docs:
  # Load context for implementation
  documents:
    - migration-strategy.md
    - data-modeling-guidelines.md
  
  extract_relevant_sections:
    # Find relevant sections based on task metadata
    if task.task_type == "implementation":
      - Data model mapping for {data_structure}
      - Phase {phase_number} implementation notes
      - Interface definition from strategy doc
    
    if task.task_type == "pre-migration-test":
      - Baseline testing requirements
      - Test coverage expectations
      - Performance metrics to capture
    
    if task.task_type == "post-migration-test":
      - Aerospike testing guidelines
      - Comparison testing approach
      - Success criteria
    
    if task.task_type == "configuration":
      - Configuration requirements
      - Environment variables
      - Dependency versions

# ------------------------------------------------------------
# Step 6: Implement Task Based on Type
# ------------------------------------------------------------
implement_task:
  input: next_task
  
  task_type_handlers:
    
    # ─────────────────────────────────────────────────────────
    # Handler: Configuration Tasks
    # ─────────────────────────────────────────────────────────
    configuration:
      actions:
        - if "go.mod" in task.files_to_modify:
            update_go_mod:
              add_dependency: "github.com/aerospike/aerospike-client-go/v7"
              run_command: "go mod tidy"
        
        - if "config.go" in task.files_to_modify:
            add_aerospike_config:
              struct_fields:
                - "AerospikeHost string"
                - "AerospikePort int"
                - "AerospikeNamespace string"
              env_mappings:
                - "AEROSPIKE_HOST"
                - "AEROSPIKE_PORT"
                - "AEROSPIKE_NAMESPACE"
        
        - if "docker-compose.yml" in task.files_to_modify:
            add_aerospike_service:
              service_definition: |
                aerospike:
                  image: aerospike/aerospike-server
                  ports:
                    - "3000:3000"
                  volumes:
                    - aerospike_data:/opt/aerospike/data
        
        - if "Makefile" in task.files_to_modify:
            add_aerospike_targets:
              - "aerospike-run"
              - "aerospike-stop"
              - "aerospike-logs"
    
    # ─────────────────────────────────────────────────────────
    # Handler: Implementation Tasks
    # ─────────────────────────────────────────────────────────
    implementation:
      actions:
        - analyze_redis_method:
            # Find the Redis method being migrated
            locate_in_codebase:
              method_name: "{task.method_name}"
              file: "{redis_service_file}"
            extract:
              - method_signature
              - parameters
              - return_type
              - redis_operations_used
        
        - generate_aerospike_equivalent:
            # Create Aerospike implementation
            use_data_model_mapping:
              lookup: "{data_structure_type}"
              from: "data-modeling-guidelines.md"
            
            template: |
              // {method_name} implements {interface_method}
              // Migrated from Redis to Aerospike
              func (a *AerospikeService) {method_signature} {
                  // Aerospike key construction
                  key, err := as.NewKey("{namespace}", "{set}", {key_param})
                  if err != nil {
                      return {error_return}
                  }
                  
                  // {operation_description}
                  {aerospike_operation_code}
                  
                  return {success_return}
              }
        
        - add_to_file:
            target: "{aerospike_service_file}"
            content: "{generated_code}"
            location: "after struct definition"
    
    # ─────────────────────────────────────────────────────────
    # Handler: Pre-Migration Test Tasks
    # ─────────────────────────────────────────────────────────
    pre-migration-test:
      actions:
        - create_baseline_test:
            file: "{test_file_path}"
            content: |
              package service_test
              
              import (
                  "testing"
                  "{module}/internal/service"
              )
              
              // TestRedisBaseline captures baseline behavior
              // Run this before migration to establish expected behavior
              func TestRedisBaseline_{phase_name}(t *testing.T) {
                  // Initialize Redis service
                  redisService := service.NewRedisService(...)
                  
                  // Test operations and capture results
                  {test_operations}
                  
                  // Document results for comparison
                  t.Logf("Baseline captured: %+v", results)
              }
        
        - create_performance_test:
            file: "{benchmark_file_path}"
            content: |
              func BenchmarkRedis_{operation}(b *testing.B) {
                  // Performance baseline for Redis
                  {benchmark_code}
              }
    
    # ─────────────────────────────────────────────────────────
    # Handler: Post-Migration Test Tasks
    # ─────────────────────────────────────────────────────────
    post-migration-test:
      actions:
        - create_aerospike_test:
            file: "{test_file_path}"
            content: |
              func TestAerospike_{method_name}(t *testing.T) {
                  // Test Aerospike implementation
                  asService := service.NewAerospikeService(...)
                  
                  // Run same operations as Redis baseline
                  {test_operations}
                  
                  // Verify results match Redis baseline
                  {assertions}
              }
        
        - create_comparison_test:
            file: "{comparison_test_path}"
            content: |
              func TestComparison_Redis_vs_Aerospike(t *testing.T) {
                  // Side-by-side comparison
                  redisResult := testWithRedis()
                  aerospikeResult := testWithAerospike()
                  
                  // Results must match exactly
                  assert.Equal(t, redisResult, aerospikeResult)
              }
    
    # ─────────────────────────────────────────────────────────
    # Handler: Validation Tasks
    # ─────────────────────────────────────────────────────────
    validation:
      actions:
        - create_validation_script:
            file: "scripts/validate-{phase_name}.go"
            content: |
              // Validation script for {phase_name}
              // Compares Redis and Aerospike outputs
              
              func main() {
                  // Load test data
                  // Run on Redis
                  // Run on Aerospike
                  // Compare results
                  // Generate report
              }

# ------------------------------------------------------------
# Step 7: Verify Implementation
# ------------------------------------------------------------
verify_implementation:
  checks:
    - files_created:
        verify_each:
          - file_exists: true
          - file_not_empty: true
          - valid_go_syntax: true
    
    - files_modified:
        verify_each:
          - changes_applied: true
          - imports_added: true
          - no_syntax_errors: true
    
    - if task.task_type == "implementation":
        additional_checks:
          - method_implements_interface: true
          - error_handling_present: true
          - documentation_added: true
  
  on_failure:
    log_error: true
    rollback_changes: true
    output: "❌ Implementation failed. Changes rolled back."

# ------------------------------------------------------------
# Step 8: Update Local Execution Log
# ------------------------------------------------------------
update_execution_log:
  output_file: "execution-log.md"
  
  append_entry: |
    ## Task Executed: {task.task_id}
    
    **Jira:** [{task.jira_key}]({jira_url})
    **Title:** {task.title}
    **Phase:** {task.phase} - {task.phase_name}
    **Type:** {task.task_type}
    **Executed At:** {timestamp}
    
    ### Changes Made:
    {for_each file in modified_files:
      - Modified: `{file}` ({lines_changed} lines)
    }
    {for_each file in created_files:
      - Created: `{file}` ({lines_added} lines)
    }
    
    ### Implementation Notes:
    {implementation_notes}
    
    ### Next Steps:
    1. Review the code changes carefully
    2. Run tests if applicable: `go test ./...`
    3. Verify functionality manually
    4. Update Jira task status to "Done"
    5. Run /code-iterate-go again for next task
    
    ---

# ------------------------------------------------------------
# Step 9: Output Summary and Next Steps
# ------------------------------------------------------------
output_summary:
  display: |
    ╔════════════════════════════════════════════════════════════════╗
    ║  ✅ TASK IMPLEMENTED SUCCESSFULLY                              ║
    ╚════════════════════════════════════════════════════════════════╝
    
    📋 Task: {task.task_id} - {task.title}
    🎫 Jira: {task.jira_key}
    
    📝 Changes Summary:
    {changes_summary}
    
    📄 Files Modified: {modified_count}
    📄 Files Created: {created_count}
    
    ╔════════════════════════════════════════════════════════════════╗
    ║  📋 YOUR ACTION REQUIRED                                       ║
    ╚════════════════════════════════════════════════════════════════╝
    
    Please complete these steps:
    
    1. 🔍 Review Code Changes
       - Check all modified/created files
       - Verify implementation matches requirements
       - Look for any TODOs or customization needed
    
    2. 🧪 Test the Changes (if applicable)
       ```bash
       go test ./...
       go run cmd/server/main.go
       ```
    
    3. ✅ Validate Functionality
       - Manual testing
       - Verify acceptance criteria met
       - Check for edge cases
    
    4. 🎫 Update Jira Status
       - Go to: {jira_url}
       - Change status from "To Do" to "Done"
       - Add comment with validation notes
    
    5. ▶️ Continue to Next Task
       Run: /code-iterate-go
       (Will automatically pick up next task after Jira updated)
    
    ────────────────────────────────────────────────────────────────
    
    💡 Tip: You can also transition task status via MCP:
       - Tell me: "Mark {task.jira_key} as Done"
       - I'll update Jira for you
    
    ╔════════════════════════════════════════════════════════════════╗
    ║  📊 MIGRATION PROGRESS                                         ║
    ╚════════════════════════════════════════════════════════════════╝
    
    Phase {current_phase}: {done_in_phase}/{total_in_phase} tasks complete
    Overall: {total_done}/{total_tasks} tasks complete ({percentage}%)
    
    Next Task: {next_task_preview}
    
    ════════════════════════════════════════════════════════════════

# ------------------------------------------------------------
# Step 10: Save Updated Task Sequence
# ------------------------------------------------------------
save_updated_sequence:
  file: "task-sequence.json"
  
  updates:
    - current_task_index: {executed_task_index}
    - tasks[{index}].status: "In Progress" (local marker)
    - tasks[{index}].executed_at: "{timestamp}"
    - tasks[{index}].local_changes: {file_list}
  
  note: |
    Local status is "In Progress" until user validates and updates Jira.
    Next run will sync from Jira and see if it's "Done" or still "In Progress".

# =====================================================================
# END OF FILE
# =====================================================================
