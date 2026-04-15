# Subscription Notes

High-risk files usually include:
- `controller/subscription.go`
- `model/subscription.go`
- `web/src/components/topup/*`
- `web/src/components/table/users/modals/UserSubscriptionsModal.jsx`

## Guardrails

- subscription plans can affect entitlements, pricing display, and user group behavior
- hidden or bonus-style plans are especially sensitive because they can bypass the normal upgrade path
- changes here require targeted backend tests plus UI validation in staging
