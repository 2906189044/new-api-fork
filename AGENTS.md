# AGENTS.md

## Scope

This repository is the only canonical source tree for the server project `new API`.
Server paths tied to this project:
- Source: `/home/admin/src/new-api`
- Production deploy: `/opt/new-api`
- Staging deploy: `/opt/new-api-staging`

AI must not touch other projects on the server.

## Mandatory Workflow

1. All code changes happen in this repository.
2. Production code is never edited inside containers.
3. Every release goes to staging first.
4. Production is updated only by switching the image tag and restarting the `app` service.
5. If a task crosses the execution boundary below, stop and ask the human before continuing.

## Execution Boundary

Human confirmation is required before any of the following:
- touching paths outside `/home/admin/src/new-api`, `/opt/new-api`, `/opt/new-api-staging`
- modifying shared reverse proxy, shared docker network, firewall, systemd, cron, or system SSH config
- schema migrations, bulk data fixes, destructive SQL, or manual data backfills
- deleting images, backups, or old source archives
- production downtime outside the agreed short restart window
- acting on a state that conflicts with documentation or cannot be safely identified

## Deployment Rules

- Production runtime state lives under `/opt/new-api`.
- Staging runtime state lives under `/opt/new-api-staging`.
- Compose files stay stable; image changes go through `APP_IMAGE` in the respective `.env` file.
- Image tags must be traceable to Git commits, format: `newapi-local:main-<short_sha>`.
- Keep only a small rollback window of recent images.

## Project-Specific Safety Notes

- This is an operations-sensitive site. Prefer small, reversible changes.
- Subscription, model sync, payment, and pricing logic are high-risk areas. Read the matching docs in `docs/features/` before modifying them.
- Preserve cross-database compatibility where model code supports SQLite, MySQL, and PostgreSQL.
- Do not remove or rewrite upstream project identity or attribution without explicit human instruction.

## Verification Requirements

Before claiming a change is complete:
- run the most relevant targeted tests for touched backend/frontend files
- validate `docker compose config` for changed deploy files
- when releasing, verify `/api/status` on staging first, then production
- record any manual steps or follow-up risk in docs if the workflow changed
