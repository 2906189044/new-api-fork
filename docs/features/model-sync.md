# Model Sync Notes

High-risk files usually include:
- `controller/model.go`
- `controller/pricing.go`
- `service/group.go`
- `service/channel_select.go`
- `middleware/distributor.go`

## Guardrails

- preserve existing model mappings unless the change explicitly requires remapping
- pricing and model visibility can drift together; validate both API output and admin UI
- confirm any new model behavior in staging before production release
