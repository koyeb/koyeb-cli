#!/usr/bin/env bash
#
# Koyeb Pool Integration Tests
#
# This script tests the `koyeb pool` command group:
# - pool create (HARD assertion: exit 0, output contains a pool ID)
# - pool list (HARD assertion: the created pool appears in the list)
# - pool get by ID (SOFT-SKIP on 501/unimplemented)
# - pool describe by NAME (HARD assertion: exercises name -> UUID
#   resolution via the idmapper; output contains the pool ID)
# - pool claim (SOFT-SKIP on 501/unimplemented)
# - pool claims list (SOFT-SKIP on 501/unimplemented)
# - pool claims get (SOFT-SKIP; needs a claim ID from claims list)
# - pool delete by NAME (HARD assertion; only for pools created by this
#   script — never deletes a pool passed via --pool)
# - pool get after delete (HARD assertion: polls until the pool 404s,
#   proving the async deletion completed; bounded by DELETION_TIMEOUT)
#
# Cleanup: the delete test removes the created pool. If the script never
# gets there (earlier failure), a best-effort 'pool delete' runs at the
# end; on failure a warning is printed so the pool can be deleted manually.
#
# Prerequisites:
#   - koyeb CLI installed and in PATH (or use KOYEB_CLI env var)
#   - Authenticated with: koyeb login
#   - KOYEB_PROJECT set to a project where pools are enabled
#
# Usage:
#   ./test_pool.sh [OPTIONS]
#
# Options:
#   --help            Show this help message
#   --verbose         Show full command output
#   --pool NAME       Use an existing pool name instead of creating one
#                     (still requires the pool to exist; skips creation)
#   --project PROJECT Project to scope the pool to (default: $KOYEB_PROJECT)
#
# Environment Variables:
#   KOYEB_CLI         Path to koyeb CLI binary (default: koyeb)
#   KOYEB_PROJECT     Project (workspace) ID or name pools are created in
#   POOL_IMAGE        Docker image for the pool (default: koyeb/sandbox)
#   TEST_TIMEOUT      Timeout for each command in seconds (default: 120)
#   DELETION_TIMEOUT   Max seconds to wait for the async deletion to
#                     complete in the get-after-delete test (default: 300)
#

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

KOYEB_CLI="${KOYEB_CLI:-koyeb}"
KOYEB_PROJECT="${KOYEB_PROJECT:-}"
POOL_IMAGE="${POOL_IMAGE:-koyeb/sandbox}"
TEST_TIMEOUT="${TEST_TIMEOUT:-120}"
DELETION_TIMEOUT="${DELETION_TIMEOUT:-300}"

TEST_ID="pool-$(date +%s)"
POOL_NAME="${POOL_NAME:-}"
TEST_PROJECT="$KOYEB_PROJECT"

# Options
VERBOSE=false

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

# Resources created (best-effort cleanup at the end via 'pool delete')
CREATED_POOL_ID=""
DELETE_TEST_PASSED=false

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
    awk 'NR>1 && /^#/ {sub(/^# ?/, ""); print; next} NR>1 && !/^#/ {exit}' "${BASH_SOURCE[0]}"
    exit 0
}

# Extract a string field value from a single-line JSON output.
json_string_field() {
    local key="$1"
    sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" | head -1
}

# Detect whether a command output indicates the endpoint is not implemented
# server-side (e.g. gRPC 501 Unimplemented). Such tests are soft-skipped.
is_unimplemented() {
    echo "$1" | grep -qiE '501|unimplemented|not implemented|notimplemented'
}

# ============================================================================
# Test Runners
# ============================================================================

# HARD assertion: the command must succeed and the output must contain a value
# for the given JSON field. On failure the test fails and the script exits.
run_test_hard_json_field() {
    local test_id="$1"
    local test_name="$2"
    local field="$3"
    shift 3
    local test_cmd=("$@")

    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "

    local output
    local exit_code=0
    output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1) || exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    local value
    value=$(echo "$output" | json_string_field "$field")

    if [[ -z "$value" ]]; then
        echo -e "${RED}FAIL${NC} (no '$field' found in output)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    echo -e "${GREEN}PASS${NC} ($field: ${value:0:8}...)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    LAST_JSON_FIELD_VALUE="$value"
    if [[ "$VERBOSE" == "true" ]]; then
        echo "$output" | sed 's/^/    /'
    fi
    return 0
}

# SOFT assertion: the command must succeed, except when the server replies
# 501/unimplemented, in which case the test is skipped with a warning.
run_test_soft() {
    local test_id="$1"
    local test_name="$2"
    shift 2
    local test_cmd=("$@")

    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "

    local output
    local exit_code=0
    output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1) || exit_code=$?

    if [[ $exit_code -eq 124 ]]; then
        echo -e "${RED}TIMEOUT${NC} (${TEST_TIMEOUT}s)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name (timeout)")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    if [[ $exit_code -eq 0 ]]; then
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        if [[ "$VERBOSE" == "true" && -n "$output" ]]; then
            echo "$output" | sed 's/^/    /'
        fi
        LAST_COMMAND_OUTPUT="$output"
        return 0
    fi

    if is_unimplemented "$output"; then
        echo -e "${YELLOW}SKIP${NC} (server replied 501/unimplemented)"
        log_warn "  $test_name is not implemented server-side, skipping"
        # ((TESTS_SKIPPED++)) would evaluate false when TESTS_SKIPPED is 0 and
        # terminate the script under `set -e` on bash >= 4; use an assignment instead.
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        LAST_COMMAND_OUTPUT=""
        return 2
    fi

    echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("$test_id: $test_name")
    echo "$output" | sed 's/^/    /' | head -20
    return 1
}

# HARD assertion: the command must exit 0.
run_test_hard() {
    local test_id="$1"
    local test_name="$2"
    shift 2
    local test_cmd=("$@")

    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "

    local output
    local exit_code=0
    output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1) || exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    LAST_COMMAND_OUTPUT="$output"
    if [[ "$VERBOSE" == "true" && -n "$output" ]]; then
        echo "$output" | sed 's/^/    /'
    fi
    return 0
}

# HARD assertion: the command must exit 0 and its output must contain the
# given substring (fixed-string match).
run_test_hard_contains() {
    local test_id="$1"
    local test_name="$2"
    local needle="$3"
    shift 3
    local test_cmd=("$@")

    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "

    local output
    local exit_code=0
    output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1) || exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    if ! echo "$output" | grep -qF -- "$needle"; then
        echo -e "${RED}FAIL${NC} ('$needle' not found in output)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_id: $test_name")
        echo "$output" | sed 's/^/    /' | head -20
        return 1
    fi

    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    LAST_COMMAND_OUTPUT="$output"
    if [[ "$VERBOSE" == "true" ]]; then
        echo "$output" | sed 's/^/    /'
    fi
    return 0
}

# HARD assertion for async deletion: poll the command until it fails with
# not-found/404, proving the resource is gone server-side. Fails when
# DELETION_TIMEOUT elapses without the pool disappearing.
run_test_hard_eventually_missing() {
    local test_id="$1"
    local test_name="$2"
    shift 2
    local test_cmd=("$@")

    echo -n -e "${BLUE}[$test_id]${NC} Testing: $test_name ... "

    local deadline=$(( $(date +%s) + DELETION_TIMEOUT ))
    local output=""
    local exit_code=0

    while [[ $(date +%s) -lt $deadline ]]; do
        exit_code=0
        output=$(timeout "$TEST_TIMEOUT" "${test_cmd[@]}" 2>&1) || exit_code=$?

        if [[ $exit_code -ne 0 ]] && echo "$output" | grep -qiE '404|not found|notfound'; then
            echo -e "${GREEN}PASS${NC} (deletion completed, pool 404s)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            LAST_COMMAND_OUTPUT="$output"
            return 0
        fi
        sleep 5
    done

    echo -e "${RED}FAIL${NC} (pool still present after ${DELETION_TIMEOUT}s)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("$test_id: $test_name")
    echo "$output" | sed 's/^/    /' | head -20
    return 1
}

# Skip a test
skip_test() {
    local test_id="$1"
    local test_name="$2"
    local reason="$3"

    echo -e "${BLUE}[$test_id]${NC} Testing: $test_name ... ${YELLOW}SKIP${NC} ($reason)"
    # Assignment form: ((TESTS_SKIPPED++)) is false when the counter is 0 and
    # would exit the script under `set -e` on bash >= 4.
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
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
        --pool)
            POOL_NAME="$2"
            shift 2
            ;;
        --project)
            TEST_PROJECT="$2"
            shift 2
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
echo "    Koyeb Pool Integration Tests"
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

# Check project scope (pool create is project-scoped server-side)
if [[ -z "$TEST_PROJECT" ]]; then
    log_error "No project specified"
    log_info "Set the KOYEB_PROJECT environment variable, or use --project"
    exit 1
fi

log_info "Project: $TEST_PROJECT"
log_info "Pool image: $POOL_IMAGE"
log_info "Test timeout: ${TEST_TIMEOUT}s per command"
echo ""

# ============================================================================
# Test: Pool Create (HARD assertion)
# ============================================================================

echo "--- Running Tests ---"
echo ""

LAST_JSON_FIELD_VALUE=""
LAST_COMMAND_OUTPUT=""

if [[ -n "$POOL_NAME" ]]; then
    log_info "Using existing pool: $POOL_NAME (skipping creation)"
    POOL_ID="$POOL_NAME"
else
    POOL_NAME="$TEST_ID"
    run_test_hard_json_field "T01" "pool create" "id" \
        "$KOYEB_CLI" pool create "$POOL_NAME" --size 1 --docker "$POOL_IMAGE" \
        --project "$TEST_PROJECT" -o json || exit 1
    POOL_ID="$LAST_JSON_FIELD_VALUE"
    CREATED_POOL_ID="$POOL_ID"
    log_info "Created pool: $POOL_NAME ($POOL_ID)"
fi

# ============================================================================
# Test: Pool List (HARD assertion: the created pool must appear)
# ============================================================================

run_test_hard_contains "T02" "pool list contains the pool" "$POOL_ID" \
    "$KOYEB_CLI" pool list --project "$TEST_PROJECT" -o json

# ============================================================================
# Test: Pool Get by ID (SOFT-SKIP on 501/unimplemented)
# ============================================================================

run_test_soft "T03" "pool get by ID" \
    "$KOYEB_CLI" pool get "$POOL_ID" --project "$TEST_PROJECT" -o json || true

# ============================================================================
# Test: Pool Describe by NAME (HARD assertion: name -> UUID resolution)
# ============================================================================

run_test_hard_contains "T04" "pool describe by name" "$POOL_ID" \
    "$KOYEB_CLI" pool describe "$POOL_NAME" --project "$TEST_PROJECT" -o json

# ============================================================================
# Test: Pool Claim (SOFT-SKIP on 501/unimplemented)
# ============================================================================

CLAIM_SOFT_SKIPPED=false
run_test_soft "T05" "pool claim" \
    "$KOYEB_CLI" pool claim "$POOL_ID" --project "$TEST_PROJECT" -o json || {
    rc=$?
    if [[ $rc -eq 2 ]]; then
        CLAIM_SOFT_SKIPPED=true
    fi
}

# ============================================================================
# Test: Pool Claims List (SOFT-SKIP on 501/unimplemented)
# ============================================================================

CLAIM_ID=""
if run_test_soft "T06" "pool claims list" \
    "$KOYEB_CLI" pool claims list "$POOL_ID" --project "$TEST_PROJECT" -o json; then
    # Extract the first claim ID from the list output for T07.
    CLAIM_ID=$(echo "$LAST_COMMAND_OUTPUT" | json_string_field "id")
    if [[ -z "$CLAIM_ID" ]]; then
        log_warn "No claim found in the claims list output"
    fi
fi

# ============================================================================
# Test: Pool Claims Get (SOFT-SKIP; skipped entirely if claim was soft-skipped)
# ============================================================================

if [[ "$CLAIM_SOFT_SKIPPED" == "true" ]]; then
    skip_test "T07" "pool claims get" "pool claim was soft-skipped (unimplemented)"
elif [[ -z "$CLAIM_ID" ]]; then
    skip_test "T07" "pool claims get" "no claim ID available from claims list"
else
    run_test_soft "T07" "pool claims get" \
        "$KOYEB_CLI" pool claims get "$CLAIM_ID" --project "$TEST_PROJECT" -o json || true
fi

# ============================================================================
# Test: Pool Delete by NAME (HARD assertion)
# ============================================================================

# Never delete a pool this script did not create.
if [[ -z "$CREATED_POOL_ID" ]]; then
    skip_test "T08" "pool delete" "using an existing pool (--pool); not deleting it"
elif run_test_hard "T08" "pool delete by name" \
    "$KOYEB_CLI" pool delete "$POOL_NAME" --project "$TEST_PROJECT"; then
    DELETE_TEST_PASSED=true
fi

# ============================================================================
# Test: Pool Get After Delete (HARD assertion: polls until the pool 404s)
# ============================================================================

if [[ "$DELETE_TEST_PASSED" == "true" ]]; then
    run_test_hard_eventually_missing "T09" "pool get 404s after delete" \
        "$KOYEB_CLI" pool get "$POOL_ID" --project "$TEST_PROJECT" -o json
else
    skip_test "T09" "pool get after delete" "pool delete did not pass"
fi

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
REMAINING_SECONDS=$((DURATION % 60))

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))

echo -e "Passed:  ${GREEN}$TESTS_PASSED${NC} / $TOTAL_TESTS"
echo -e "Failed:  ${RED}$TESTS_FAILED${NC} / $TOTAL_TESTS"
echo -e "Skipped: ${YELLOW}$TESTS_SKIPPED${NC} / $TOTAL_TESTS"
echo ""
echo "Duration: ${MINUTES}m ${REMAINING_SECONDS}s"
echo ""

if [[ ${#FAILED_TESTS[@]} -gt 0 ]]; then
    echo -e "${RED}Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo ""
fi

if [[ -n "$CREATED_POOL_ID" && "$DELETE_TEST_PASSED" != "true" ]]; then
    # Best-effort cleanup: the delete test did not run or did not pass
    # (e.g. an earlier test failed), so the pool may still exist.
    if timeout "$TEST_TIMEOUT" "$KOYEB_CLI" pool delete "$CREATED_POOL_ID" --project "$TEST_PROJECT" >/dev/null 2>&1; then
        log_info "Cleanup: pool '$POOL_NAME' ($CREATED_POOL_ID) deleted."
    else
        log_warn "Cleanup failed: pool '$POOL_NAME' ($CREATED_POOL_ID) PERSISTS in the organization."
        log_warn "Delete it manually: $KOYEB_CLI pool delete $CREATED_POOL_ID --project $TEST_PROJECT"
    fi
    echo ""
fi

if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}TESTS FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}ALL TESTS PASSED (skips allowed)${NC}"
    exit 0
fi
