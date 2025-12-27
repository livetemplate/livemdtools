#!/bin/bash
# validate.sh - Validate all skill examples are valid Livemdtools apps
#
# This script checks that each example file:
# 1. Has valid frontmatter
# 2. Contains at least one lvt code block
# 3. Has valid HTML structure in lvt blocks
#
# Usage: ./validate.sh [--verbose]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLES_DIR="$SCRIPT_DIR/../examples"
VERBOSE=${1:-""}

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass=0
fail=0
warnings=0

log_pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    pass=$((pass + 1))
}

log_fail() {
    echo -e "${RED}FAIL${NC}: $1"
    fail=$((fail + 1))
}

log_warn() {
    echo -e "${YELLOW}WARN${NC}: $1"
    warnings=$((warnings + 1))
}

log_info() {
    if [ "$VERBOSE" = "--verbose" ]; then
        echo "INFO: $1"
    fi
}

validate_example() {
    local file="$1"
    local basename=$(basename "$file")

    log_info "Validating $basename..."

    # Check file exists
    if [ ! -f "$file" ]; then
        log_fail "$basename - File not found"
        return 1
    fi

    # Check for frontmatter
    if ! head -n 1 "$file" | grep -q "^---$"; then
        log_fail "$basename - Missing frontmatter (should start with ---)"
        return 1
    fi

    # Check frontmatter has title
    if ! grep -q "^title:" "$file"; then
        log_warn "$basename - Missing title in frontmatter"
    fi

    # Check for lvt code block
    if ! grep -q '```lvt' "$file"; then
        log_fail "$basename - No lvt code block found"
        return 1
    fi

    # Check for at least one lvt-* attribute
    if ! grep -q 'lvt-' "$file"; then
        log_fail "$basename - No lvt-* attributes found"
        return 1
    fi

    # Check for common patterns (at least one should exist)
    local has_persist=false
    local has_source=false

    grep -q 'lvt-persist' "$file" && has_persist=true || true
    grep -q 'lvt-source' "$file" && has_source=true || true

    if [ "$has_persist" = false ] && [ "$has_source" = false ]; then
        log_warn "$basename - No data binding (lvt-persist or lvt-source)"
    fi

    # Check HTML structure - look for unclosed tags in lvt blocks
    # Extract lvt block content
    local lvt_content
    lvt_content=$(sed -n '/```lvt/,/```/p' "$file" | grep -v '```') || true

    # Simple check: count opening and closing div tags
    local open_divs close_divs
    open_divs=$(echo "$lvt_content" | grep -o '<div' | wc -l | tr -d ' ') || open_divs=0
    close_divs=$(echo "$lvt_content" | grep -o '</div>' | wc -l | tr -d ' ') || close_divs=0

    if [ "$open_divs" -ne "$close_divs" ]; then
        log_warn "$basename - Mismatched div tags (open: $open_divs, close: $close_divs)"
    fi

    log_pass "$basename"
    return 0
}

echo "========================================"
echo "Livemdtools Skill Examples Validator"
echo "========================================"
echo ""

# Find all example files
for file in "$EXAMPLES_DIR"/*.md; do
    if [ -f "$file" ]; then
        validate_example "$file"
    fi
done

echo ""
echo "========================================"
echo "Summary"
echo "========================================"
echo -e "Passed:   ${GREEN}$pass${NC}"
echo -e "Failed:   ${RED}$fail${NC}"
echo -e "Warnings: ${YELLOW}$warnings${NC}"
echo ""

if [ $fail -gt 0 ]; then
    echo -e "${RED}Validation failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All examples valid!${NC}"
    exit 0
fi
