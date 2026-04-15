# Backup And Restore

## What To Back Up

- `/opt/new-api/.env`
- `/opt/new-api/docker-compose.yml`
- database dumps under `/opt/new-api/backups`
- any pre-migration snapshots taken before schema or data changes

## Required Human Confirmation

Stop and ask before:
- schema migrations
- bulk data changes
- restore operations
- deleting old backups

## Retention

- keep recent operational dumps in `/opt/new-api/backups`
- keep only a small rolling window of images and backups to control disk usage
- never delete the currently deployed production or staging image tags
