# Cleanup

Use the cleanup script to manage old `new API` images and backups without touching other projects.

## Dry Run

```bash
/home/admin/src/new-api/scripts/cleanup-newapi.sh
```

This prints:
- protected images currently used by production or staging
- recent images kept for rollback
- images that would be removed
- backups older than the retention window that would be removed

## Apply

```bash
/home/admin/src/new-api/scripts/cleanup-newapi.sh --apply
```

## Retention Defaults

- keep the 5 most recent `newapi-local:*` images in addition to the production and staging tags
- list backups older than 30 days for deletion

You can override defaults per run:

```bash
KEEP_IMAGES=7 BACKUP_DAYS=45 /home/admin/src/new-api/scripts/cleanup-newapi.sh
```
