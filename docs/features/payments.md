# Payment Notes

High-risk files include payment and top-up handlers plus related frontend purchase flows.

## Guardrails

- treat payment callbacks, entitlement creation, and pricing display as one change surface
- verify both successful purchase behavior and failure handling
- do not perform data repair or billing backfills without human confirmation
