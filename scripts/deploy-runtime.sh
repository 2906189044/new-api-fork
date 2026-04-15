#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <runtime-dir> <image-tag>" >&2
  exit 1
fi

runtime_dir="$1"
tag="$2"

runtime_project_name() {
  case "$1" in
    /opt/new-api)
      echo "newapi"
      ;;
    /opt/new-api-staging)
      echo "newapi-staging"
      ;;
    *)
      echo "Unsupported runtime dir: $1" >&2
      return 1
      ;;
  esac
}

project_name="$(runtime_project_name "${runtime_dir}")"

cd "${runtime_dir}"

python3 - "${tag}" <<'PY'
from pathlib import Path
import sys

p = Path('.env')
tag = sys.argv[1]
lines = p.read_text().splitlines()
out = []
replaced = False
for line in lines:
    if line.startswith('APP_IMAGE='):
        out.append(f'APP_IMAGE={tag}')
        replaced = True
    else:
        out.append(line)
if not replaced:
    out.append(f'APP_IMAGE={tag}')
p.write_text('\n'.join(out) + '\n')
PY

# Always pin the compose project name so deployment never creates a second stack
# such as "new-api" beside the real "newapi" production runtime.
sudo docker compose -p "${project_name}" config >/dev/null
sudo docker compose -p "${project_name}" up -d app
sudo docker compose -p "${project_name}" ps app
