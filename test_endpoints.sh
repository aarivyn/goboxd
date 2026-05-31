#!/bin/bash
# goboxd - Comprehensive Endpoint Testing Script
# Run this after server is running on port 8080

set -e

BASE_URL="${1:-http://localhost:8080}"
FAILURES=0
PASSED=0

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test helper function
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5
    
    echo -n "Testing ${name}... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        echo "Response: $body" | head -c 100
        echo ""
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC} (Expected HTTP $expected_code, got $http_code)"
        echo "Response: $body" | head -c 100
        echo ""
        ((FAILURES++))
    fi
}

echo "=========================================="
echo "goboxd API Test Suite"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo ""

# ========== HEALTH & INFO ENDPOINTS ==========
echo -e "${YELLOW}[Section 1: Health & Info Endpoints]${NC}"

test_endpoint "GET /healthz" "GET" "/healthz" "" "200"
test_endpoint "GET /readyz" "GET" "/readyz" "" "200"
test_endpoint "GET /info" "GET" "/info" "" "200"

echo ""

# ========== NEW LANGUAGE ENDPOINTS ==========
echo -e "${YELLOW}[Section 2: New Language Endpoints]${NC}"

test_endpoint "GET /languages (list all)" "GET" "/languages" "" "200"
test_endpoint "GET /languages/py3 (valid)" "GET" "/languages/py3" "" "200"
test_endpoint "GET /languages/cpp (valid)" "GET" "/languages/cpp" "" "200"
test_endpoint "GET /languages/unknown (404)" "GET" "/languages/unknown" "" "404"

echo ""

# ========== EXECUTION ENDPOINT ==========
echo -e "${YELLOW}[Section 3: Code Execution Tests]${NC}"

# Python simple test
PYTHON_REQ='{
  "language": "py3",
  "source": "print(\"hello world\")",
  "tests": [{"stdin": "", "expected_stdout": "hello world\n"}]
}'
test_endpoint "POST /run (Python success)" "POST" "/run" "$PYTHON_REQ" "200"

# Python wrong output
PYTHON_WRONG='{
  "language": "py3",
  "source": "print(\"goodbye\")",
  "tests": [{"stdin": "", "expected_stdout": "hello world\n"}]
}'
test_endpoint "POST /run (Python wrong output)" "POST" "/run" "$PYTHON_WRONG" "200"

# Invalid language
INVALID_LANG='{
  "language": "xyz",
  "source": "test",
  "tests": [{"stdin": "", "expected_stdout": ""}]
}'
test_endpoint "POST /run (unknown language)" "POST" "/run" "$INVALID_LANG" "400"

# Missing language
NO_LANG='{
  "source": "test",
  "tests": [{"stdin": "", "expected_stdout": ""}]
}'
test_endpoint "POST /run (missing language)" "POST" "/run" "$NO_LANG" "400"

# Missing tests
NO_TESTS='{
  "language": "py3",
  "source": "test",
  "tests": []
}'
test_endpoint "POST /run (missing tests)" "POST" "/run" "$NO_TESTS" "400"

echo ""

# ========== METHOD NOT ALLOWED TESTS ==========
echo -e "${YELLOW}[Section 4: Method Validation]${NC}"

test_endpoint "POST /languages (method not allowed)" "POST" "/languages" "{}" "405"
test_endpoint "POST /healthz (method not allowed)" "POST" "/healthz" "{}" "405"

echo ""

# ========== RESULTS ==========
echo "=========================================="
echo -e "Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILURES} failed${NC}"
echo "=========================================="

if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
