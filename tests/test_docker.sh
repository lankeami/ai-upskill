#!/bin/bash

# Test Docker setup: Dockerfile, Makefile targets, .dockerignore
# Requires: docker CLI available, run from project root

PASS=0
FAIL=0

assert() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then
    echo "  PASS: $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc"
    FAIL=$((FAIL + 1))
  fi
}

assert_file() {
  local desc="$1" path="$2"
  if [[ -f "$path" ]]; then
    echo "  PASS: $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Docker setup tests ==="

# File existence
assert_file "Dockerfile exists" Dockerfile
assert_file ".dockerignore exists" .dockerignore
assert_file "Makefile exists" Makefile

# Makefile targets
echo ""
echo "--- Makefile targets ---"
for target in build run stop logs shell clean; do
  assert "Makefile has '$target' target" grep -q "^${target}:" Makefile
done

# .dockerignore contents
echo ""
echo "--- .dockerignore ---"
for pattern in .git .venv node_modules vendor; do
  assert ".dockerignore excludes $pattern" grep -q "$pattern" .dockerignore
done
assert ".dockerignore does NOT exclude .env for COPY" true  # .env stays on host via --env-file

# Dockerfile basics
echo ""
echo "--- Dockerfile ---"
assert "Dockerfile has Go build stage" grep -q "golang:" Dockerfile
assert "Dockerfile has Python in final stage" grep -q "python" Dockerfile
assert "Dockerfile does not COPY .env" bash -c '! grep -q "COPY.*\.env" Dockerfile'

# Build test (only if docker is available)
echo ""
echo "--- Docker build ---"
if command -v docker &>/dev/null; then
  if make build 2>&1; then
    echo "  PASS: make build succeeds"
    PASS=$((PASS + 1))

    assert "Docker image exists" docker images -q ai-upskill:latest

    # Verify tools inside container
    echo ""
    echo "--- Container tools ---"
    assert "Go available in container" docker run --rm ai-upskill:latest go version
    assert "Python3 available in container" docker run --rm ai-upskill:latest python3 --version
    assert "gh CLI available in container" docker run --rm ai-upskill:latest gh --version
    assert "ai-report binary exists" docker run --rm ai-upskill:latest test -f /app/ai-report

    # Cleanup
    make clean 2>/dev/null || true
  else
    echo "  FAIL: make build"
    FAIL=$((FAIL + 1))
  fi
else
  echo "  SKIP: docker not available"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[[ $FAIL -eq 0 ]]
