package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func initSubscriptionBonusTestDB(t *testing.T) {
	t.Helper()
	common.IsMasterNode = true
	common.SQLitePath = filepath.Join(t.TempDir(), fmt.Sprintf("%s.sqlite", t.Name()))
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	if err := InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(common.SQLitePath)
	})
}

func createTestUser(t *testing.T, username, group string) *User {
	t.Helper()
	user := &User{
		Username: username,
		Password: "password123",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    group,
	}
	if err := DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createTestPlan(t *testing.T, plan SubscriptionPlan) *SubscriptionPlan {
	t.Helper()
	if err := DB.Create(&plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	InvalidateSubscriptionPlanCache(plan.Id)
	return &plan
}

func TestSubscriptionPlanIncludesBonusFields(t *testing.T) {
	plan := SubscriptionPlan{}
	if _, ok := any(plan).(interface{}); !ok {
		t.Fatal("unexpected type assertion failure")
	}
	if plan.StackableBonus {
		t.Fatal("expected StackableBonus to default false")
	}
	if plan.BonusModelScope != "" {
		t.Fatal("expected BonusModelScope to default empty")
	}
}

func TestStackableBonusSubscriptionDoesNotOverrideUserGroup(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "bonus_user", "plan_codex_basic")
	plan := createTestPlan(t, SubscriptionPlan{
		Title:              "管理员专用 GPT 日限额计划",
		VisibleToUser:      false,
		Enabled:            true,
		UpgradeGroup:       "plan_hidden_gpt_daily",
		StackableBonus:     true,
		BonusModelScope:    "gpt_series",
		TotalAmount:        10,
		DurationUnit:       SubscriptionDurationDay,
		DurationValue:      10,
		QuotaResetPeriod:   SubscriptionResetDaily,
		PriceAmount:        0,
		MaxPurchasePerUser: 0,
	})

	if _, err := AdminBindSubscription(user.Id, plan.Id, ""); err != nil {
		t.Fatalf("bind bonus subscription: %v", err)
	}

	updatedUser, err := GetUserById(user.Id, true)
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updatedUser.Group != "plan_codex_basic" {
		t.Fatalf("expected user group to remain plan_codex_basic, got %s", updatedUser.Group)
	}
}

func TestStackableBonusSubscriptionCannotBeClaimedTwice(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "bonus_repeat_user", "default")
	plan := createTestPlan(t, SubscriptionPlan{
		Title:            "管理员专用 GPT 日限额计划",
		VisibleToUser:    false,
		Enabled:          true,
		StackableBonus:   true,
		BonusModelScope:  "gpt_series",
		TotalAmount:      10,
		DurationUnit:     SubscriptionDurationDay,
		DurationValue:    10,
		QuotaResetPeriod: SubscriptionResetDaily,
	})

	if _, err := AdminBindSubscription(user.Id, plan.Id, ""); err != nil {
		t.Fatalf("first bind failed: %v", err)
	}

	if _, err := AdminBindSubscription(user.Id, plan.Id, ""); err == nil {
		t.Fatal("expected second bind to fail for one-time bonus plan")
	}
}

func TestPreConsumeUserSubscriptionPrioritizesGPTStackableBonus(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "bonus_preconsume_user", "plan_coding_all")
	mainPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "Coding Plan",
		VisibleToUser:    true,
		Enabled:          true,
		TotalAmount:      100,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_coding_all",
	})
	bonusPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "管理员专用 GPT 日限额计划",
		VisibleToUser:    false,
		Enabled:          true,
		StackableBonus:   true,
		BonusModelScope:  "gpt_series",
		TotalAmount:      10,
		DurationUnit:     SubscriptionDurationDay,
		DurationValue:    10,
		QuotaResetPeriod: SubscriptionResetDaily,
	})

	if _, err := AdminBindSubscription(user.Id, mainPlan.Id, ""); err != nil {
		t.Fatalf("bind main subscription: %v", err)
	}
	if _, err := AdminBindSubscription(user.Id, bonusPlan.Id, ""); err != nil {
		t.Fatalf("bind bonus subscription: %v", err)
	}

	gptResult, err := PreConsumeUserSubscription("req-gpt", user.Id, "gpt-5.4-thinking", 0, 5)
	if err != nil {
		t.Fatalf("gpt preconsume: %v", err)
	}
	claudeResult, err := PreConsumeUserSubscription("req-claude", user.Id, "claude-sonnet-4.5", 0, 5)
	if err != nil {
		t.Fatalf("claude preconsume: %v", err)
	}

	if gptResult.UserSubscriptionId == claudeResult.UserSubscriptionId {
		t.Fatal("expected GPT request to use bonus subscription and Claude to use main subscription")
	}

	bonusInfo, err := GetSubscriptionPlanInfoByUserSubscriptionId(gptResult.UserSubscriptionId)
	if err != nil {
		t.Fatalf("load gpt subscription plan info: %v", err)
	}
	if !strings.Contains(bonusInfo.PlanTitle, "管理员专用 GPT") {
		t.Fatalf("expected GPT request to consume bonus plan, got %s", bonusInfo.PlanTitle)
	}

	mainInfo, err := GetSubscriptionPlanInfoByUserSubscriptionId(claudeResult.UserSubscriptionId)
	if err != nil {
		t.Fatalf("load claude subscription plan info: %v", err)
	}
	if mainInfo.PlanTitle != "Coding Plan" {
		t.Fatalf("expected Claude request to consume main plan, got %s", mainInfo.PlanTitle)
	}
}
