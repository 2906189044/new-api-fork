package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestSubscriptionPlanCurrencyDefault(t *testing.T) {
	plan := model.SubscriptionPlan{
		Title:    "test",
		Currency: "",
	}

	normalizeSubscriptionPlanInput(&plan)

	if plan.Currency != "CNY" {
		t.Fatalf("expected currency to default to CNY, got %q", plan.Currency)
	}
}

func TestSubscriptionPlanCurrencyCoercesLegacyUSDToCNY(t *testing.T) {
	plan := model.SubscriptionPlan{
		Title:    "test",
		Currency: "USD",
	}

	normalizeSubscriptionPlanInput(&plan)

	if plan.Currency != "CNY" {
		t.Fatalf("expected legacy currency to be coerced to CNY, got %q", plan.Currency)
	}
}
