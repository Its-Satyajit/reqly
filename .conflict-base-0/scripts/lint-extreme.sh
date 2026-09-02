#!/usr/bin/env bash
set -uo pipefail

# Ensure ~/go/bin and node_modules/.bin are in PATH
export PATH="$HOME/go/bin:$(pwd)/node_modules/.bin:$PATH"

REPORT_FILE="${1:-LINT_REPORT.md}"
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

START_TIME=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo "=========================================================="
echo " Starting Extreme Quality & Lint Audit (Parallel Execution)"
echo " Report destination: $REPORT_FILE"
echo "=========================================================="

declare -a STAGE_NAMES
declare -a STAGE_COMMANDS

add_stage() {
  STAGE_NAMES+=("$1")
  STAGE_COMMANDS+=("$2")
}

# 1. Frontend TypeScript Typecheck
add_stage "Frontend TypeScript Typecheck" "nub run typecheck"

# 2. Frontend AST Linter (via nub run or node_modules binary)
add_stage "Frontend AST Lint (oxlint)" "nub run lint:frontend"

# 3. Go Compiler Build
add_stage "Go Compiler Build" "go build ./..."

# 4. Go Vet (Suspicious Constructs)
add_stage "Go Vet (Suspicious Constructs)" "go vet ./..."

# 5. Go Unit & Integration Tests
add_stage "Go Unit & Integration Tests" "go test ./..."

# 6. Go Race Detector Tests
add_stage "Go Race Detector Tests" "go test -race ./..."

# 7. Go Static Analysis (staticcheck)
add_stage "Go Deep Static Analysis (staticcheck)" "staticcheck ./..."

# 8. Go Reachability (deadcode)
add_stage "Go Dead Code / Reachability (deadcode)" "deadcode ./..."

# 9. JS/TS Dead Code Detector
add_stage "JS/TS Dead Code (fallow dead-code)" "nubx -y fallow dead-code"

# 10. JS/TS Duplication & Clones
add_stage "JS/TS Duplication & Clones (fallow dupes)" "nubx -y fallow dupes"

# 11. JS/TS Complexity & Health
add_stage "JS/TS Health & Complexity (fallow health)" "nubx -y fallow health"

# 12. JS/TS Advisory Changed-File Review
add_stage "JS/TS Advisory Review (fallow review)" "nubx -y fallow review"

# 13. React Architecture & Hook Doctor
add_stage "React Architecture & Performance (react-doctor)" "CI=1 printf '\n' | nubx -y react-doctor@latest"

echo "Spawning ${#STAGE_NAMES[@]} verification stages in parallel..."

# Run each stage in parallel and capture stdout + stderr, exit status, and timing
for i in "${!STAGE_NAMES[@]}"; do
  NAME="${STAGE_NAMES[$i]}"
  CMD="${STAGE_COMMANDS[$i]}"
  LOG_FILE="$TEMP_DIR/stage_${i}.log"
  META_FILE="$TEMP_DIR/stage_${i}.meta"

  (
    START_SEC=$(date +%s)
    if eval "$CMD" > "$LOG_FILE" 2>&1; then
      EXIT_CODE=0
    else
      EXIT_CODE=$?
    fi
    END_SEC=$(date +%s)
    DURATION=$((END_SEC - START_SEC))
    echo "EXIT_CODE=$EXIT_CODE" > "$META_FILE"
    echo "DURATION=${DURATION}s" >> "$META_FILE"
  ) &
done

# Wait for all background tasks to finish
wait

echo "All stages finished. Aggregating results..."

OVERALL_STATUS="PASSED"
declare -a STAGE_STATUSES
declare -a STAGE_DURATIONS

for i in "${!STAGE_NAMES[@]}"; do
  META_FILE="$TEMP_DIR/stage_${i}.meta"
  EXIT_CODE=$(grep '^EXIT_CODE=' "$META_FILE" | cut -d= -f2)
  DUR=$(grep '^DURATION=' "$META_FILE" | cut -d= -f2)
  
  STAGE_DURATIONS+=("$DUR")

  if [ "$EXIT_CODE" -eq 0 ]; then
    STAGE_STATUSES+=("PASS")
    echo "  [✓ PASS] ${STAGE_NAMES[$i]} ($DUR)"
  else
    STAGE_STATUSES+=("FAIL")
    OVERALL_STATUS="FAILED"
    echo "  [✗ FAIL] ${STAGE_NAMES[$i]} ($DUR)"
  fi
done

END_TIME=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo ""
echo "=========================================================="
echo " Generating Markdown Report: $REPORT_FILE"
echo "=========================================================="

cat << EOF > "$REPORT_FILE"
# Extreme Quality & Lint Audit Report

- **Generated At:** $START_TIME
- **Completed At:** $END_TIME
- **Overall Status:** $([ "$OVERALL_STATUS" = "PASSED" ] && echo "🟢 **PASSED**" || echo "🔴 **FAILED / ISSUES FOUND**")

---

## Executive Summary

| # | Check / Tool | Status | Duration | Command |
|---|--------------|:------:|:--------:|---------|
EOF

for i in "${!STAGE_NAMES[@]}"; do
  NAME="${STAGE_NAMES[$i]}"
  CMD="${STAGE_COMMANDS[$i]}"
  STATUS="${STAGE_STATUSES[$i]}"
  DURATION="${STAGE_DURATIONS[$i]}"

  if [ "$STATUS" = "PASS" ]; then
    ICON="🟢 PASS"
  else
    ICON="🔴 FAIL"
  fi

  echo "| $((i+1)) | **$NAME** | $ICON | $DURATION | \`$CMD\` |" >> "$REPORT_FILE"
done

cat << 'EOF' >> "$REPORT_FILE"

---

## Detailed Tool Output Logs

EOF

for i in "${!STAGE_NAMES[@]}"; do
  NAME="${STAGE_NAMES[$i]}"
  CMD="${STAGE_COMMANDS[$i]}"
  STATUS="${STAGE_STATUSES[$i]}"
  DURATION="${STAGE_DURATIONS[$i]}"
  LOG_FILE="$TEMP_DIR/stage_${i}.log"

  # Strip ANSI color codes for clean markdown rendering
  CLEAN_LOG=$(sed -r 's/\x1B\[[0-9;]*[a-zA-Z]//g' "$LOG_FILE" 2>/dev/null || cat "$LOG_FILE")

  cat << EOF >> "$REPORT_FILE"
### $((i+1)). $NAME

- **Command:** \`$CMD\`
- **Status:** $([ "$STATUS" = "PASS" ] && echo "🟢 PASS" || echo "🔴 FAIL")
- **Duration:** $DURATION

<details>
<summary>Click to expand full output</summary>

\`\`\`text
$CLEAN_LOG
\`\`\`

</details>

---

EOF
done

echo "Report successfully generated: $REPORT_FILE"
exit 0
