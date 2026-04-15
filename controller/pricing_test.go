package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestFilterPricingByUsableGroupsReturnsAllModelsForDisplay(t *testing.T) {
	pricing := []model.Pricing{
		{ModelName: "gpt-5.4-thinking", EnableGroup: []string{"gpt_core"}},
		{ModelName: "claude-opus-4.6", EnableGroup: []string{"claude_core"}},
	}
	usableGroup := map[string]string{
		"gpt_core": "GPT 系列",
	}

	filtered := filterPricingByUsableGroups(pricing, usableGroup)

	if len(filtered) != len(pricing) {
		t.Fatalf("expected all pricing rows to remain visible, got %d want %d", len(filtered), len(pricing))
	}
}

func TestFilterPricingByUsableGroupsHidesUpstreamOnlyModels(t *testing.T) {
	pricing := []model.Pricing{
		{ModelName: "gpt-5.4", EnableGroup: []string{"upstream"}},
		{ModelName: "gpt-5.4-mini", EnableGroup: []string{"upstream", "gpt_core"}},
		{ModelName: "gpt-5.4-thinking", EnableGroup: []string{"gpt_core"}},
	}

	filtered := filterPricingByUsableGroups(pricing, map[string]string{"gpt_core": "GPT 系列"})

	if len(filtered) != 2 {
		t.Fatalf("expected upstream-only model to be hidden, got %d rows", len(filtered))
	}
	for _, row := range filtered {
		if row.ModelName == "gpt-5.4" {
			t.Fatal("expected upstream-only model gpt-5.4 to be excluded from pricing output")
		}
	}
}
