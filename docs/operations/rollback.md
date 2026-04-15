# Rollback

## Production Rollback

1. Identify the last known good image tag.
2. Run `/opt/new-api/scripts/rollback.sh <tag>`.
3. Verify `http://127.0.0.1:13000/api/status`.
4. Re-check the affected business flow.

## Notes

- Rollback is image-based. Do not attempt emergency hotfixes inside the running container.
- Database restore is a separate operation and requires human confirmation.
