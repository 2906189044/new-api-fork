# Deployment

## Standard Flow

1. Develop and review changes in the canonical repository.
2. Push the commit to GitHub.
3. On the server, update `/home/admin/src/new-api` with `git pull`.
4. Build a commit-tagged image with `scripts/build-image.sh`.
5. Release that image to staging with `scripts/release-to-staging.sh <tag>`.
6. Verify staging health and key business flows.
7. Promote the exact validated tag to production with `scripts/promote-to-production.sh <tag>`.

## Rules

- Do not deploy from historical `newapi-src-*` directories.
- Do not hand-edit container contents.
- Do not publish unverified tags to production.
- If a release needs schema changes or data repair, stop for human confirmation first.
