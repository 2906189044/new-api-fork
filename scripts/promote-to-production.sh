#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <image-tag>" >&2
  exit 1
fi

/opt/new-api/scripts/deploy.sh "$1"
echo "Production promoted to $1"
