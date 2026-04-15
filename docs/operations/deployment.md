# Deployment (Single Host, Dual Environment, Best Practice)

## Scope

- **Upstream fixed project**: `CLI API` (deployed at `/opt/cliproxyapi`)
- **Relay project**: `New API` (this repository)
- **Host**: `yifei-server3`
- **Topology**: one server, two fully isolated New API runtimes

## Environment Topology

### Production (Commercial)
- Runtime dir: `/opt/new-api`
- Compose project: `newapi`
- Internal bind: `127.0.0.1:13000 -> app:3000`
- Public domain: `https://7ontop.com`
- Nginx vhost: `/etc/nginx/sites-available/7ontop.com.conf`

### Staging (Test)
- Runtime dir: `/opt/new-api-staging`
- Compose project: `newapi-staging`
- Internal bind: `127.0.0.1:13001 -> app:3000`
- Public domain: `http://test.2662777.xyz` (HTTPS can be enabled after cert issuance)
- Nginx vhost: `/etc/nginx/sites-available/test.2662777.xyz.conf`

### Upstream Fixed Service
- Runtime dir: `/opt/cliproxyapi`
- Role: protocol transformation upstream for New API
- Shared docker network: `shared_ai_net`

## Isolation Rules (Must Keep)

1. **Different compose project names**
   - prod: `newapi`
   - staging: `newapi-staging`
2. **Different runtime dirs**
   - `/opt/new-api` vs `/opt/new-api-staging`
3. **Different domains and ports**
   - prod `7ontop.com -> 13000`
   - staging `test.2662777.xyz -> 13001`
4. **Different secrets**
   - `SESSION_SECRET`, `CRYPTO_SECRET` must differ between prod/staging
5. **No direct edits inside running containers**
6. **Always deploy via repository scripts**

## Standard Release Flow

1. Develop and review changes in GitHub.
2. On server, update source and release to staging:

```bash
/home/admin/src/new-api/scripts/update-staging.sh
```

3. Validate staging:
   - `curl -fsS http://127.0.0.1:13001/api/status`
   - key business flows and model/relay paths
4. Promote the **exact same tested image tag** to production:

```bash
/home/admin/src/new-api/scripts/promote-to-production.sh newapi-local:main-<short_sha>
```

5. Validate production:
   - `curl -fsS http://127.0.0.1:13000/api/status`
   - public domain spot checks on `https://7ontop.com`

## Domain / Nginx Operations

### Reload Nginx safely

```bash
sudo nginx -t && sudo systemctl reload nginx
```

### Staging HTTPS (optional but recommended)

After DNS `A test.2662777.xyz -> yifei-server3` is ready:

```bash
sudo certbot --nginx -d test.2662777.xyz
```

Then verify:

```bash
curl -I https://test.2662777.xyz
```

## Quick Verification Checklist

```bash
# runtime health
curl -fsS http://127.0.0.1:13000/api/status
curl -fsS http://127.0.0.1:13001/api/status

# compose stacks
cd /opt/new-api && sudo docker compose -p newapi ps
cd /opt/new-api-staging && sudo docker compose -p newapi-staging ps

# env guardrails (check staging URL)
grep ^FRONTEND_BASE_URL= /opt/new-api/.env /opt/new-api-staging/.env
```

## Guardrails / Anti-Patterns

- Do **not** deploy from historical `newapi-src-*` snapshots.
- Do **not** use ad-hoc compose project names.
- Do **not** promote an image tag that was not validated in staging.
- Do **not** share prod domain in staging env (`FRONTEND_BASE_URL`).

## Rollback

If production validation fails, redeploy previous known-good image tag:

```bash
/home/admin/src/new-api/scripts/promote-to-production.sh <previous-good-tag>
```

Keep previous-good tag recorded for every release.
