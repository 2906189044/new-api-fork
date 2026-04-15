#!/usr/bin/env bash
set -euo pipefail

cd /home/admin/src/new-api
sha=$(git rev-parse --short HEAD 2>/dev/null || true)
if [[ -z "${sha}" ]]; then
  echo "Cannot determine git commit for image tag." >&2
  exit 1
fi

tag="newapi-local:main-${sha}"
echo "Building ${tag}"
sudo docker build -t "${tag}" .
echo "Built ${tag}"
