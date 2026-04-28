#!/usr/bin/env bash
# check-pipeline-contract.sh — enforces SHC-03 (single source of truth
# for every pipeline step).
#
# The contract: nothing in api-go/internal/server/ is allowed to call
# f2cclient.F2CClient RPC methods directly outside _test.go files.
# Every production per-step operation MUST go through pipeline.Steps.
#
# Exits 0 when the contract holds, 1 otherwise. Prints the offending
# lines on violation.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SERVER_DIR="${ROOT}/internal/server"

if [ ! -d "${SERVER_DIR}" ]; then
    echo "check-pipeline-contract: ${SERVER_DIR} not found" >&2
    exit 2
fi

# Patterns that indicate a direct F2CClient RPC call. Matches the
# canonical receiver-based form s.f2c.<Method>(...) used in Phase 11
# before the Plan 12-01 refactor.
#
# After the refactor, s.steps.<Method>(...) is the only legal form in
# handlers; s.f2c.<Method>(...) must not exist outside tests.
VIOLATIONS="$(
    grep -rn \
        --include='*.go' \
        --exclude='*_test.go' \
        -E '\bs\.f2c\.[A-Z][A-Za-z]+\(' \
        "${SERVER_DIR}" \
        || true
)"

if [ -n "${VIOLATIONS}" ]; then
    echo "SHC-03 contract violation: direct f2cclient calls found in internal/server/." >&2
    echo "Every per-step operation must go through pipeline.Steps (s.steps.X)." >&2
    echo "" >&2
    echo "${VIOLATIONS}" >&2
    exit 1
fi

# Also verify that pipeline.Steps is actually imported / referenced
# (cheap sanity check — catches a future refactor that accidentally
# removes the pipeline dependency).
if ! grep -rq --include='*.go' --exclude='*_test.go' \
        -E 'pipeline\.Steps|s\.steps\.' "${SERVER_DIR}"; then
    echo "check-pipeline-contract: no references to pipeline.Steps in server/." >&2
    echo "The refactor may have been reverted — verify internal/server/server.go." >&2
    exit 1
fi

echo "check-pipeline-contract: OK (SHC-03 holds)"
