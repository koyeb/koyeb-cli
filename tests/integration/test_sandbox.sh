#!/usr/bin/env bash
#
# Koyeb Sandbox Integration Tests
#
# This script tests the koyeb sandbox command functionality including:
# - Sandbox creation and listing
# - Command execution (run, start, ps, kill, logs)
# - Filesystem operations (read, write, ls, mkdir, rm, upload, download)
# - Port management (expose-port, unexpose-port)
# - Health checks
#
# Prerequisites:
#   - koyeb CLI installed and in PATH (or use KOYEB_CLI env var)
#   - Authenticated with: koyeb login
#   - Access to create apps and services
#
# Usage:
#   ./test_sandbox.sh [OPTIONS]
#
# Options:
#   --help              Show this help message
#   --verbose           Show full command output
#   --keep-resources    Don't delete resources after tests (app auto-deletes after 1 day)
#   --sandbox NAME      Use existing sandbox instead of creating one
#   --app NAME          Use existing app instead of creating one
#   --skip-create       Skip sandbox creation (use with --sandbox)
#
# Environment Variables:
#   KOYEB_CLI           Path to koyeb CLI binary (default: koyeb)
#   SANDBOX_IMAGE       Docker image for sandbox (default: koyeb/sandbox)
#   TEST_TIMEOUT        Timeout for each test in seconds (default: 120)
#

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KOYEB_CLI="${KOYEB_CLI:-koyeb}"
SANDBOX_IMAGE="${SANDBOX_IMAGE:-koyeb/sandbox}"
TEST_TIMEOUT="${TEST_TIMEOUT:-120}"
DEPLOYMENT_TIMEOUT="${DEPLOYMENT_TIMEOUT:-300}"

# Test identifiers (generated)
TEST_ID="sb-$(date +%s)"
TEST_APP="${TEST_APP:-}"
TEST_SANDBOX="${TEST_SANDBOX:-}"

# Options
VERBOSE=false
KEEP_RESOURCES=false
SKIP_CREATE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0
FAILED_TESTS=()
START_TIME=$(date +%s)

# Resources created (for cleanup)
CREATED_APP=""
CREATED_SANDBOX=""
PROCESS_ID=""
TEST_LOCAL_FILE=""

# ============================================================================
# Helper Functions
# ============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $*"
}

show_help() {
    cat << 'EOF'
Koyeb Sandbox Integration Tests

This script tests the koyeb sandbox command functionality including:
- Sandbox creation and listing
- Command execution (run, start, ps, kill, logs)
- Filesystem operations (read, write, ls, mkdir, rm, upload, download)
- Port management (expose-port, unexpose-port)
- Health checks

Prerequisites:
  - koyeb CLI installed and in PATH (or use KOYEB_CLI env var)
  - Authenticated with: koyeb login
  - Access to create apps and services

Usage:
  ./test_sandbox.sh [OPTIONS]

Options:
  --help              Show this help message
  --verbose           Show full command output
  --keep-resources    Don't delete resources after tests (app auto-deletes after 1 day)
  --sandbox NAME      Use existing sandbox instead of creating one
  --app NAME          Use existing app instead of creating one
  --skip-create       Skip sandbox creation (use with --sandbox)

Environment Variables:
  KOYEB_CLI           Path to koyeb CLI binary (default: koyeb)
  SANDBOX_IMAGE       Docker image for sandbox (default: koyeb/sandbox)
  TEST_TIMEOUT        Timeout for each test in seconds (default: 120)
EOF
    exit 0
}

# Run koyeb command with optional verbose output
run_koyeb() {
    if [[ "$VERBOSE" == "true" ]]; then
        "$KOYEB_CLI" "$@"
    else
        "$KOYEB_CLI" "$@" 2>&1
    fi
}

# Run a test with timeout and result tracking
run_test() {
    local test_id="$1"
    local test_name="$2"
    shift 2
    local test_cmd=("$@")
    
    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "
    
    local start_time=$(date +%s)
    local output
    local exit_code=0
    
    # Run the test command with timeout
    if output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $exit_code -eq 0 ]]; then
        echo -e "${GREEN}PASS${NC} (${duration}s)"
        ((TESTS_PASSED++))
        if [[ "$VERBOSE" == "true" && -n "$output" ]]; then
            echo "$output" | sed 's/^/    /'
        fi
        return 0
    elif [[ $exit_code -eq 124 ]]; then
        echo -e "${RED}TIMEOUT${NC} (${TEST_TIMEOUT}s)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_id: $test_name (timeout)")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    else
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi
}

# Run a test that checks output contains expected string
run_test_contains() {
    local test_id="$1"
    local test_name="$2"
    local expected="$3"
    shift 3
    local test_cmd=("$@")
    
    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "
    
    local start_time=$(date +%s)
    local output
    local exit_code=0
    
    if output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi
    
    if echo "$output" | grep -q "$expected"; then
        echo -e "${GREEN}PASS${NC} (${duration}s)"
        ((TESTS_PASSED++))
        if [[ "$VERBOSE" == "true" ]]; then
            echo "$output" | sed 's/^/    /'
        fi
        return 0
    else
        echo -e "${RED}FAIL${NC} (expected output containing: $expected)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "    Output was:"
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi
}

# Skip a test
skip_test() {
    local test_id="$1"
    local test_name="$2"
    local reason="$3"
    
    echo -e "${BLUE}[$test_id]${NC} Testing: $test_name ... ${YELLOW}SKIP${NC} ($reason)"
    ((TESTS_SKIPPED++))
}

# Wait for sandbox to be healthy
wait_for_healthy() {
    local sandbox_name="$1"
    local max_attempts="${2:-60}"
    local attempt=0
    
    log_info "Waiting for sandbox to become healthy..."
    
    while [[ $attempt -lt $max_attempts ]]; do
        if "$KOYEB_CLI" sandbox health "$sandbox_name" 2>&1 | grep -q "Healthy: true"; then
            log_info "Sandbox is healthy!"
            return 0
        fi
        ((attempt++))
        sleep 5
    done
    
    log_error "Sandbox did not become healthy within $((max_attempts * 5)) seconds"
    return 1
}

# Cleanup function
cleanup() {
    local exit_code=$?
    
    echo ""
    log_info "Cleaning up..."
    
    # Remove local test file
    if [[ -n "$TEST_LOCAL_FILE" && -f "$TEST_LOCAL_FILE" ]]; then
        rm -f "$TEST_LOCAL_FILE"
    fi
    
    if [[ "$KEEP_RESOURCES" == "true" ]]; then
        log_info "Keeping resources (--keep-resources specified)"
        if [[ -n "$CREATED_APP" ]]; then
            log_info "App '$CREATED_APP' will auto-delete after 1 day"
        fi
        return $exit_code
    fi
    
    # Delete sandbox service
    if [[ -n "$CREATED_SANDBOX" ]]; then
        log_info "Deleting sandbox service: $CREATED_SANDBOX"
        "$KOYEB_CLI" service delete "$CREATED_SANDBOX" 2>/dev/null || true
    fi
    
    # Delete app (if we created it)
    if [[ -n "$CREATED_APP" ]]; then
        log_info "Deleting app: $CREATED_APP"
        "$KOYEB_CLI" app delete "$CREATED_APP" 2>/dev/null || true
    fi
    
    return $exit_code
}

# ============================================================================
# Parse Arguments
# ============================================================================

while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --keep-resources)
            KEEP_RESOURCES=true
            shift
            ;;
        --sandbox)
            TEST_SANDBOX="$2"
            shift 2
            ;;
        --app)
            TEST_APP="$2"
            shift 2
            ;;
        --skip-create)
            SKIP_CREATE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# ============================================================================
# Pre-flight Checks
# ============================================================================

echo ""
echo "=============================================="
echo "   Koyeb Sandbox Integration Tests"
echo "=============================================="
echo ""

# Check koyeb CLI is available
if ! command -v "$KOYEB_CLI" &> /dev/null; then
    log_error "koyeb CLI not found: $KOYEB_CLI"
    log_info "Set KOYEB_CLI environment variable or ensure 'koyeb' is in PATH"
    exit 1
fi

log_info "Using koyeb CLI: $(command -v "$KOYEB_CLI")"

# Check authentication
if ! "$KOYEB_CLI" organization list &> /dev/null; then
    log_error "Not authenticated with Koyeb"
    log_info "Run 'koyeb login' first"
    exit 1
fi

log_info "Authentication: OK"
log_info "Sandbox image: $SANDBOX_IMAGE"
log_info "Test timeout: ${TEST_TIMEOUT}s per test"
echo ""

# Set up cleanup trap
trap cleanup EXIT

# ============================================================================
# Setup Phase
# ============================================================================

echo "--- Setup Phase ---"
echo ""

# Create or use existing app
if [[ -n "$TEST_APP" ]]; then
    log_info "Using existing app: $TEST_APP"
    APP_NAME="$TEST_APP"
else
    APP_NAME="$TEST_ID"
    log_info "Creating test app: $APP_NAME (with auto-delete after 1 day)"
    
    if ! "$KOYEB_CLI" app create "$APP_NAME" 2>&1; then
        log_error "Failed to create app"
        exit 1
    fi
    CREATED_APP="$APP_NAME"
    log_success "App created: $APP_NAME"
fi

# Create or use existing sandbox
if [[ -n "$TEST_SANDBOX" ]]; then
    SANDBOX_NAME="$TEST_SANDBOX"
    log_info "Using existing sandbox: $SANDBOX_NAME"
elif [[ "$SKIP_CREATE" == "true" ]]; then
    log_error "--skip-create requires --sandbox to specify existing sandbox"
    exit 1
else
    SANDBOX_NAME="$APP_NAME/sandbox"
    log_info "Creating sandbox: $SANDBOX_NAME"
    log_info "This may take a few minutes..."
    
    if ! "$KOYEB_CLI" sandbox create "$SANDBOX_NAME" \
        --docker "$SANDBOX_IMAGE" \
        --instance-type nano \
        --delete-after-delay 1h \
        --wait \
        --wait-timeout "${DEPLOYMENT_TIMEOUT}s" 2>&1; then
        log_error "Failed to create sandbox"
        exit 1
    fi
    CREATED_SANDBOX="$SANDBOX_NAME"
    log_success "Sandbox created: $SANDBOX_NAME"
    
    # Wait for healthy
    if ! wait_for_healthy "$SANDBOX_NAME" 60; then
        log_error "Sandbox failed to become healthy"
        exit 1
    fi
fi

echo ""
echo "--- Running Tests ---"
echo ""

# ============================================================================
# Test: Sandbox List
# ============================================================================

run_test_contains "T01" "sandbox list" "$APP_NAME" \
    "$KOYEB_CLI" sandbox list || true

run_test_contains "T02" "sandbox list --app filter" "sandbox" \
    "$KOYEB_CLI" sandbox list --app "$APP_NAME" || true

# ============================================================================
# Test: Sandbox Health
# ============================================================================

run_test_contains "T03" "sandbox health" "Healthy: true" \
    "$KOYEB_CLI" sandbox health "$SANDBOX_NAME" || true

# ============================================================================
# Test: Command Execution (run)
# ============================================================================

run_test_contains "T04" "sandbox run - echo command" "hello-world" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" echo "hello-world" || true

run_test_contains "T05" "sandbox run - pwd command" "/" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" pwd || true

run_test_contains "T06" "sandbox run - ls command" "tmp" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" ls / || true

run_test_contains "T07" "sandbox run --cwd" "/tmp" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" --cwd /tmp pwd || true

run_test_contains "T08" "sandbox run --env" "test-value-123" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" --env "TEST_VAR=test-value-123" printenv TEST_VAR || true

run_test_contains "T09" "sandbox run --stream" "hello" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" --stream echo "hello from stream" || true

run_test_contains "T10" "sandbox run --timeout" "" \
    "$KOYEB_CLI" sandbox run "$SANDBOX_NAME" --timeout 10 echo "quick command" || true

# ============================================================================
# Test: Background Processes
# ============================================================================

# Start a background process (use a command that produces output for log testing)
echo -n -e "${BLUE}[T11]${NC} Testing: sandbox start - background process ... "
START_EXIT_CODE=0
START_OUTPUT=$("$KOYEB_CLI" sandbox start "$SANDBOX_NAME" 'echo sandbox-log-test; sleep 300' 2>&1) || START_EXIT_CODE=$?
if [[ $START_EXIT_CODE -eq 0 ]]; then
    PROCESS_ID=$(echo "$START_OUTPUT" | grep -oE '[a-f0-9-]{36}' | head -1)
    if [[ -n "$PROCESS_ID" ]]; then
        echo -e "${GREEN}PASS${NC} (PID: ${PROCESS_ID:0:8}...)"
        ((TESTS_PASSED++))
    else
        # Try alternative format
        PROCESS_ID=$(echo "$START_OUTPUT" | grep -oE 'Process ID: [^ ]+' | cut -d' ' -f3)
        if [[ -n "$PROCESS_ID" ]]; then
            echo -e "${GREEN}PASS${NC} (PID: ${PROCESS_ID:0:8}...)"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}FAIL${NC} (could not extract process ID)"
            echo "    Output: $START_OUTPUT"
            ((TESTS_FAILED++))
            FAILED_TESTS+=("T11: sandbox start")
        fi
    fi
else
    echo -e "${RED}FAIL${NC}"
    echo "    Output: $START_OUTPUT"
    ((TESTS_FAILED++))
    FAILED_TESTS+=("T11: sandbox start")
fi

# List processes
if [[ -n "$PROCESS_ID" ]]; then
    run_test_contains "T12" "sandbox ps - list processes" "sleep" \
        "$KOYEB_CLI" sandbox ps "$SANDBOX_NAME" || true
else
    skip_test "T12" "sandbox ps - list processes" "no process ID from T11"
fi

# Give the process a moment to produce output
sleep 2

# TODO: logs command currently blocks because the server always streams (no non-streaming endpoint).
#       Need to either add a /process_logs JSON endpoint to the sandbox server, or implement
#       client-side idle-timeout logic to collect buffered logs and disconnect.
# Get process logs
#if [[ -n "$PROCESS_ID" ]]; then
#    run_test_contains "T13" "sandbox logs - process logs" "sandbox-log-test" \
#        "$KOYEB_CLI" sandbox logs "$SANDBOX_NAME" "$PROCESS_ID" || true
#else
#    skip_test "T13" "sandbox logs - process logs" "no process ID from T11"
#fi

# Kill process
if [[ -n "$PROCESS_ID" ]]; then
    run_test "T14" "sandbox kill - kill process" \
        "$KOYEB_CLI" sandbox kill "$SANDBOX_NAME" "$PROCESS_ID" || true
else
    skip_test "T14" "sandbox kill - kill process" "no process ID from T11"
fi

# ============================================================================
# Test: Filesystem Operations
# ============================================================================

# Write a file
run_test "T15" "sandbox fs write - write file" \
    "$KOYEB_CLI" sandbox fs write "$SANDBOX_NAME" /tmp/test-file.txt "Hello from integration test!" || true

# Read the file back
run_test_contains "T16" "sandbox fs read - read file" "Hello from integration test" \
    "$KOYEB_CLI" sandbox fs read "$SANDBOX_NAME" /tmp/test-file.txt || true

# List directory
run_test_contains "T17" "sandbox fs ls - list directory" "test-file.txt" \
    "$KOYEB_CLI" sandbox fs ls "$SANDBOX_NAME" /tmp || true

# List with long format
run_test_contains "T18" "sandbox fs ls -l - long format" "test-file.txt" \
    "$KOYEB_CLI" sandbox fs ls "$SANDBOX_NAME" /tmp -l || true

# Create directory
run_test "T19" "sandbox fs mkdir - create directory" \
    "$KOYEB_CLI" sandbox fs mkdir "$SANDBOX_NAME" /tmp/test-dir || true

# Verify directory was created
run_test_contains "T20" "sandbox fs ls - verify mkdir" "test-dir" \
    "$KOYEB_CLI" sandbox fs ls "$SANDBOX_NAME" /tmp || true

# Write file from local file
TEST_LOCAL_FILE=$(mktemp)
echo "Content from local file" > "$TEST_LOCAL_FILE"
run_test "T21" "sandbox fs write -f - write from file" \
    "$KOYEB_CLI" sandbox fs write "$SANDBOX_NAME" /tmp/from-local.txt -f "$TEST_LOCAL_FILE" || true

# Verify file was written
run_test_contains "T22" "sandbox fs read - verify write from file" "Content from local file" \
    "$KOYEB_CLI" sandbox fs read "$SANDBOX_NAME" /tmp/from-local.txt || true

# Upload a file
echo "Upload test content" > "$TEST_LOCAL_FILE"
run_test "T23" "sandbox fs upload - upload file" \
    "$KOYEB_CLI" sandbox fs upload "$SANDBOX_NAME" "$TEST_LOCAL_FILE" /tmp/uploaded.txt || true

# Verify upload
run_test_contains "T24" "sandbox fs read - verify upload" "Upload test content" \
    "$KOYEB_CLI" sandbox fs read "$SANDBOX_NAME" /tmp/uploaded.txt || true

# Download a file
DOWNLOAD_FILE=$(mktemp)
rm -f "$DOWNLOAD_FILE"  # Remove so download can create it
run_test "T25" "sandbox fs download - download file" \
    "$KOYEB_CLI" sandbox fs download "$SANDBOX_NAME" /tmp/uploaded.txt "$DOWNLOAD_FILE" || true

# Verify download
if [[ -f "$DOWNLOAD_FILE" ]] && grep -q "Upload test content" "$DOWNLOAD_FILE"; then
    echo -e "${BLUE}[T26]${NC} Testing: verify download content ... ${GREEN}PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${BLUE}[T26]${NC} Testing: verify download content ... ${RED}FAIL${NC}"
    ((TESTS_FAILED++))
    FAILED_TESTS+=("T26: verify download content")
fi
rm -f "$DOWNLOAD_FILE"

# Remove file
run_test "T27" "sandbox fs rm - remove file" \
    "$KOYEB_CLI" sandbox fs rm "$SANDBOX_NAME" /tmp/test-file.txt || true

# Remove directory
run_test "T28" "sandbox fs rm -r - remove directory" \
    "$KOYEB_CLI" sandbox fs rm "$SANDBOX_NAME" /tmp/test-dir --recursive || true

# ============================================================================
# Test: Port Management
# ============================================================================

run_test "T29" "sandbox expose-port - expose port 8080" \
    "$KOYEB_CLI" sandbox expose-port "$SANDBOX_NAME" 8080 || true

# Give it a moment
sleep 2

run_test "T30" "sandbox unexpose-port - unexpose port" \
    "$KOYEB_CLI" sandbox unexpose-port "$SANDBOX_NAME" || true

# ============================================================================
# Test Summary
# ============================================================================

echo ""
echo "=============================================="
echo "   Test Summary"
echo "=============================================="
echo ""

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
MINUTES=$((DURATION / 60))
SECONDS=$((DURATION % 60))

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))

echo -e "Passed:  ${GREEN}$TESTS_PASSED${NC} / $TOTAL_TESTS"
echo -e "Failed:  ${RED}$TESTS_FAILED${NC} / $TOTAL_TESTS"
echo -e "Skipped: ${YELLOW}$TESTS_SKIPPED${NC} / $TOTAL_TESTS"
echo ""
echo "Duration: ${MINUTES}m ${SECONDS}s"
echo ""

if [[ ${#FAILED_TESTS[@]} -gt 0 ]]; then
    echo -e "${RED}Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo ""
fi

if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}TESTS FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
    exit 0
fi
