#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <image-tag>" >&2
  exit 1
fi

echo "Releasing $1 to staging"
/opt/new-api-staging/scripts/deploy.sh "$1"
echo "Staging released with $1"
