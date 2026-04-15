# Local Subscription Testing

Use the commands below when you want to verify subscription renew, upgrade, downgrade rejection, and the hidden GPT bonus behavior locally.

## Fast Smoke Check

This runs the targeted Go tests and then executes a local end-to-end smoke scenario against a temporary SQLite database.

```bash
./scripts/local-subscription-smoke.sh
```

Expected result:

- the two Go test commands pass
- the smoke runner prints:
  - `ok: renew and upgrade flow`
  - `ok: downgrade rejected`
  - `ok: hidden GPT bonus is GPT-only`
  - `subscription smoke passed`

## Local Click-Through Sandbox

This creates a fresh local SQLite database, seeds visible plans plus the hidden GPT bonus plan, creates several ready-to-use users, and starts the app locally.

```bash
./scripts/run-local-subscription-sandbox.sh
```

Default local URL:

```text
http://127.0.0.1:3000
```

Default login accounts:

- `root / 123456`
- `sandbox_new / password123`
- `sandbox_basic / password123`
- `sandbox_advanced / password123`
- `sandbox_coding / password123`
- `sandbox_bonus / password123`

What each user is for:

- `sandbox_new`: no active main plan
- `sandbox_basic`: active `天卡`
- `sandbox_advanced`: active `CODECS专属套餐`
- `sandbox_coding`: active `CODING PLAN尊享套餐`
- `sandbox_bonus`: active `CODING PLAN尊享套餐` plus hidden GPT bonus plan

## What To Verify In The Sandbox

- `sandbox_basic` can only renew `天卡` or upgrade upward
- `sandbox_coding` cannot buy lower visible plans
- `sandbox_bonus` keeps the hidden GPT bonus alongside the main plan
- the hidden GPT bonus never replaces the user's primary group

## Notes

- `Stripe` and `Creem` still use fixed product pricing, so list-price-difference upgrades are blocked there by the backend.
- `ePay` is the payment path that can carry the resolved difference amount.
- The smoke runner does not need any external payment configuration.
