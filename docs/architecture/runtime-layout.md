# Runtime Layout

`new API` is split into one source tree and two runtime directories.

- `/home/admin/src/new-api`: canonical source repository used for all code changes and image builds
- `/opt/new-api`: production runtime directory with compose, env, data, logs, backups, postgres data, redis data
- `/opt/new-api-staging`: staging runtime directory with isolated compose, env, data, logs, backups, postgres data, redis data

## Release Model

- Build images from `/home/admin/src/new-api`
- Tag images as `newapi-local:main-<short_sha>`
- Write the approved tag into the target runtime `.env` as `APP_IMAGE`
- Restart only the target `app` service with Docker Compose

## Safety Boundary

Only the directories above belong to this maintenance workflow. Shared server services and unrelated projects are out of scope unless a human explicitly approves broader work.
