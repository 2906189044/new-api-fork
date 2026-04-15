package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
)

func TestGetUserUsableGroupsIncludesAutoWhenDefaultAutoGroupEnabled(t *testing.T) {
	original := setting.DefaultUseAutoGroup
	setting.DefaultUseAutoGroup = true
	t.Cleanup(func() {
		setting.DefaultUseAutoGroup = original
	})

	groups := GetUserUsableGroups("plan_codex_basic")

	if _, ok := groups["auto"]; !ok {
		t.Fatal("expected auto group to be usable when default auto group is enabled")
	}
}

func initGroupServiceTestDB(t *testing.T) {
	t.Helper()
	common.IsMasterNode = true
	common.SQLitePath = filepath.Join(t.TempDir(), fmt.Sprintf("%s.sqlite", t.Name()))
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	if err := model.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(common.SQLitePath)
	})
}

func TestGetUserUsableGroupsForUserIncludesGPTBonusGroups(t *testing.T) {
	initGroupServiceTestDB(t)

	user := &model.User{
		Username: "bonus_group_user",
		Password: "password123",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	plan := &model.SubscriptionPlan{
		Title:            "管理员专用 GPT 日限额计划",
		VisibleToUser:    false,
		Enabled:          true,
		StackableBonus:   true,
		BonusModelScope:  "gpt_series",
		TotalAmount:      10,
		DurationUnit:     model.SubscriptionDurationDay,
		DurationValue:    10,
		QuotaResetPeriod: model.SubscriptionResetDaily,
	}
	if err := model.DB.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if _, err := model.AdminBindSubscription(user.Id, plan.Id, ""); err != nil {
		t.Fatalf("bind bonus subscription: %v", err)
	}

	usableGroups := GetUserUsableGroupsForUser(user.Id, "default")
	if _, ok := usableGroups["gpt_core"]; !ok {
		t.Fatal("expected gpt_core to become usable for bonus user")
	}
	if _, ok := usableGroups["gpt52_unlimited"]; !ok {
		t.Fatal("expected gpt52_unlimited to become usable for bonus user")
	}
}

func TestGetUserUsableGroupsForUserIncludesCodexBasicGroups(t *testing.T) {
	usableGroups := GetUserUsableGroupsForUser(0, "plan_codex_basic")
	if _, ok := usableGroups["gpt_core"]; !ok {
		t.Fatal("expected codex basic plan to include gpt_core")
	}
	if _, ok := usableGroups["claude_core"]; ok {
		t.Fatal("did not expect codex basic plan to include claude_core")
	}
	if _, ok := usableGroups["upstream"]; ok {
		t.Fatal("did not expect codex basic plan to include upstream")
	}
}

func TestGetUserUsableGroupsForUserIncludesCodexAdvancedUnlimitedGroups(t *testing.T) {
	usableGroups := GetUserUsableGroupsForUser(0, "plan_codex_advanced")
	if _, ok := usableGroups["gpt_core"]; !ok {
		t.Fatal("expected codex advanced plan to include gpt_core")
	}
	if _, ok := usableGroups["gpt52_unlimited"]; !ok {
		t.Fatal("expected codex advanced plan to include gpt52_unlimited")
	}
	if _, ok := usableGroups["claude_core"]; ok {
		t.Fatal("did not expect codex advanced plan to include claude_core")
	}
}

func TestGetUserUsableGroupsForUserIncludesCodingPlanAllModelGroups(t *testing.T) {
	usableGroups := GetUserUsableGroupsForUser(0, "plan_coding_all")
	for _, group := range []string{"gpt_core", "gpt52_unlimited", "claude_core", "gemini_core", "upstream"} {
		if _, ok := usableGroups[group]; !ok {
			t.Fatalf("expected coding plan to include %s", group)
		}
	}
}
