#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go test ./model -run 'TestResolveSubscriptionPurchaseTx|TestCompleteSubscriptionOrder|TestPreConsumeUserSubscriptionPrioritizesGPTStackableBonus'
go test ./controller -run 'Test.*Subscription'
go run ./cmd/subscription-smoke
