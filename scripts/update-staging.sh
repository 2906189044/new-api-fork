#!/usr/bin/env bash
set -euo pipefail

repo_dir="/home/admin/src/new-api"
staging_url="http://127.0.0.1:13001/api/status"
git_ssh_cmd="ssh -i /home/admin/.ssh/new_api_github -o IdentitiesOnly=yes"

cd "${repo_dir}"

echo "Updating source tree in ${repo_dir}"
GIT_SSH_COMMAND="${git_ssh_cmd}" git fetch origin main
git checkout main >/dev/null 2>&1 || true
GIT_SSH_COMMAND="${git_ssh_cmd}" git pull --ff-only origin main

tag="$(./scripts/build-image.sh | awk '/^Built newapi-local:/{print $2}' | tail -n 1)"
if [[ -z "${tag}" ]]; then
  echo "Failed to determine built image tag." >&2
  exit 1
fi

./scripts/release-to-staging.sh "${tag}"

echo "Waiting for staging health"
for _ in $(seq 1 18); do
  if curl -fsS "${staging_url}" >/tmp/newapi-staging-status.json 2>/dev/null; then
    echo "Staging is ready with ${tag}"
    break
  fi
  sleep 5
done

if ! curl -fsS "${staging_url}" >/tmp/newapi-staging-status.json 2>/dev/null; then
  echo "Staging health check failed for ${tag}" >&2
  exit 1
fi

echo "Current commit: $(git rev-parse HEAD)"
echo "Current image: ${tag}"
echo "Staging response:"
sed -n '1p' /tmp/newapi-staging-status.json

cat <<EOF

Next step:
  /home/admin/src/new-api/scripts/promote-to-production.sh ${tag}
EOF
