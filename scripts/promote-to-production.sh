#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <image-tag>" >&2
  exit 1
fi

echo "Promoting $1 to production"
/opt/new-api/scripts/deploy.sh "$1"
echo "Verifying production health"
for _ in $(seq 1 12); do
  if curl -fsS http://127.0.0.1:13000/api/status >/tmp/newapi-production-status.json 2>/dev/null; then
    echo "Production promoted to $1"
    sed -n '1p' /tmp/newapi-production-status.json
    exit 0
  fi
  sleep 5
done

echo "Production health check failed for $1" >&2
exit 1
