package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
)

func TestFilterPricingByUsableGroupsLimitsRowsToUsableGroups(t *testing.T) {
	pricing := []model.Pricing{
		{ModelName: "gpt-5.4-thinking", EnableGroup: []string{"gpt_core"}},
		{ModelName: "claude-opus-4.6", EnableGroup: []string{"claude_core"}},
	}
	usableGroup := map[string]string{
		"gpt_core": "GPT 系列",
	}

	filtered := filterPricingByUsableGroups(pricing, usableGroup)

	if len(filtered) != 1 {
		t.Fatalf("expected only GPT pricing row to remain visible, got %d", len(filtered))
	}
	if filtered[0].ModelName != "gpt-5.4-thinking" {
		t.Fatalf("expected GPT row to remain, got %s", filtered[0].ModelName)
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

func TestFilterPricingByUsableGroupsKeepsUpstreamModelsForCodingPlan(t *testing.T) {
	pricing := []model.Pricing{
		{ModelName: "grok-4.20-fast", EnableGroup: []string{"upstream"}},
		{ModelName: "claude-opus-4.6", EnableGroup: []string{"claude_core"}},
	}

	filtered := filterPricingByUsableGroups(pricing, map[string]string{
		"upstream":    "全部模型",
		"claude_core": "Claude 系列",
	})

	if len(filtered) != 2 {
		t.Fatalf("expected coding plan style groups to retain upstream rows, got %d", len(filtered))
	}
}
