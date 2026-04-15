#!/usr/bin/env bash
set -euo pipefail

cd /home/admin/src/new-api
sha=$(git rev-parse --short HEAD 2>/dev/null || true)
branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
if [[ -z "${sha}" ]]; then
  echo "Cannot determine git commit for image tag." >&2
  exit 1
fi

if [[ -z "${branch}" || "${branch}" == "HEAD" ]]; then
  branch="main"
fi

tag="newapi-local:${branch}-${sha}"
echo "Source branch: ${branch}"
echo "Source commit: ${sha}"
echo "Building ${tag}"
sudo docker build -t "${tag}" .
echo "Built ${tag}"
