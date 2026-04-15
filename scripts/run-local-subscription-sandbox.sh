#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

db_path="${1:-$PWD/.local/subscription-sandbox.sqlite}"
mkdir -p "$(dirname "$db_path")"

go run ./cmd/subscription-sandbox --db "$db_path"

export SQLITE_PATH="$db_path"
export SESSION_SECRET="${SESSION_SECRET:-local-subscription-sandbox}"
export NODE_TYPE=master
export DEBUG="${DEBUG:-true}"
export PORT="${PORT:-3000}"

echo
echo "Starting local new API on http://127.0.0.1:${PORT}"
echo "SQLite DB: ${db_path}"
echo "Use root / 123456 or the sandbox_* users listed above."
echo

go run .
