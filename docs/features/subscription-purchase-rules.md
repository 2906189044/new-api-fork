# Subscription Purchase Rules

This file documents the server-side purchase rules for user-visible main subscription plans.

## Visible Plan Rank

- Rank is defined by the same order users see in the frontend: `sort_order desc, id desc`.
- `天卡` is the lowest user-visible tier.
- `CODING PLAN尊享套餐` is the highest user-visible tier.

## Main Plan Purchase Rules

- If the user has no active main plan, buying a visible plan creates a new main subscription.
- If the user already has an active main plan:
  - buying the same plan is treated as `renew`
  - buying a higher visible plan is treated as `upgrade`
  - buying a lower visible plan is rejected

## Renew

- Renew extends the existing main subscription instead of creating a second active main subscription.
- The extension uses the target plan's configured duration once per order.

## Upgrade

- Upgrade charges only the list-price difference: `target.price_amount - current.price_amount`.
- Upgrade keeps the original expiry time unchanged.
- Upgrade cancels the previous main subscription record and creates a replacement subscription for the higher plan.

## Hidden Bonus Plans

- Hidden/admin bonus plans must use `stackable_bonus=true`.
- Stackable bonus plans do not participate in main-plan rank comparison.
- The GPT daily bonus plan only applies to GPT-series requests through `bonus_model_scope=gpt_series`.
- Stackable bonus plans must not override the user's primary group.

## Payment Entry Notes

- User payment endpoints must call `model.CreateSubscriptionOrderTx(...)` so all payment methods share the same rule engine.
- `Stripe` and `Creem` currently use fixed product/price bindings, so list-price-difference upgrades are blocked there unless a future implementation adds dynamic checkout pricing.
- `ePay` supports difference charging because the request amount is created from the resolved order amount.
