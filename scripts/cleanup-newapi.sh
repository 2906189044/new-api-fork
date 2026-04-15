#!/usr/bin/env bash
set -euo pipefail

keep_images="${KEEP_IMAGES:-5}"
backup_days="${BACKUP_DAYS:-30}"
apply=false

if [[ "${1:-}" == "--apply" ]]; then
  apply=true
fi

current_images="$(
  {
    grep '^APP_IMAGE=' /opt/new-api/.env 2>/dev/null || true
    grep '^APP_IMAGE=' /opt/new-api-staging/.env 2>/dev/null || true
  } | cut -d= -f2 | sort -u
)"

mapfile -t candidate_images < <(
  sudo docker images --format '{{.Repository}}:{{.Tag}}' |
    grep '^newapi-local:' |
    awk '!seen[$0]++'
)

echo "Protected image tags:"
printf '%s\n' "${current_images}"
echo

count=0
for image in "${candidate_images[@]}"; do
  count=$((count + 1))
  if printf '%s\n' "${current_images}" | grep -qx "${image}"; then
    echo "KEEP in-use image ${image}"
    continue
  fi
  if (( count <= keep_images )); then
    echo "KEEP recent image ${image}"
    continue
  fi
  if [[ "${apply}" == true ]]; then
    echo "DELETE image ${image}"
    sudo docker rmi "${image}"
  else
    echo "DRY-RUN delete image ${image}"
  fi
done

while IFS= read -r backup; do
  [[ -n "${backup}" ]] || continue
  if [[ "${apply}" == true ]]; then
    echo "DELETE backup ${backup}"
    sudo rm -f "${backup}"
  else
    echo "DRY-RUN delete backup ${backup}"
  fi
done < <(find /opt/new-api/backups -type f -mtime +"${backup_days}" 2>/dev/null | sort)

if [[ "${apply}" == false ]]; then
  cat <<EOF

Dry run only. Re-run with:
  /home/admin/src/new-api/scripts/cleanup-newapi.sh --apply
EOF
fi
