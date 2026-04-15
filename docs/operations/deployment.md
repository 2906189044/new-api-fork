# Deployment

## Standard Flow

1. Develop and review changes in the canonical repository.
2. Push the commit to GitHub.
3. On the server, run `/home/admin/src/new-api/scripts/update-staging.sh`.
4. Verify staging health and key business flows.
5. Promote the exact validated tag to production with `/home/admin/src/new-api/scripts/promote-to-production.sh <tag>`.

## One-Command Server Workflow

`update-staging.sh` performs:

1. `git fetch` and `git pull --ff-only origin main`
2. build a commit-tagged image
3. release that image to staging
4. wait for `http://127.0.0.1:13001/api/status`
5. print the image tag to promote if staging looks good

Typical usage:

```bash
/home/admin/src/new-api/scripts/update-staging.sh
```

If staging is approved, promote the exact tag shown by the script:

```bash
/home/admin/src/new-api/scripts/promote-to-production.sh newapi-local:main-<short_sha>
```

## Rules

- Do not deploy from historical `newapi-src-*` directories.
- Do not hand-edit container contents.
- Do not publish unverified tags to production.
- If a release needs schema changes or data repair, stop for human confirmation first.
