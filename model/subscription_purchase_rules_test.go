package model

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestResolveSubscriptionPurchaseTxRejectsDowngrade(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "downgrade_user", "plan_coding_all")
	lowPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "天卡",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      1,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_basic",
		SortOrder:        10,
	})
	highPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "CODING PLAN尊享套餐",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      44,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_coding_all",
		SortOrder:        30,
	})
	if _, err := AdminBindSubscription(user.Id, highPlan.Id, ""); err != nil {
		t.Fatalf("bind high plan: %v", err)
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		_, err := ResolveSubscriptionPurchaseTx(tx, user.Id, lowPlan)
		return err
	})
	if err == nil {
		t.Fatal("expected downgrade purchase to be rejected")
	}
	if err != ErrSubscriptionPlanDowngrade {
		t.Fatalf("expected ErrSubscriptionPlanDowngrade, got %v", err)
	}
}

func TestResolveSubscriptionPurchaseTxUpgradeUsesPriceDifference(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "upgrade_user", "plan_codex_basic")
	lowPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "天卡",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      1,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_basic",
		SortOrder:        10,
	})
	highPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "CODING PLAN尊享套餐",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      44,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_coding_all",
		SortOrder:        30,
	})
	if _, err := AdminBindSubscription(user.Id, lowPlan.Id, ""); err != nil {
		t.Fatalf("bind low plan: %v", err)
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		ctx, err := ResolveSubscriptionPurchaseTx(tx, user.Id, highPlan)
		if err != nil {
			return err
		}
		if ctx.Mode != SubscriptionPurchaseModeUpgrade {
			t.Fatalf("expected upgrade mode, got %s", ctx.Mode)
		}
		if ctx.ChargeAmount != 43 {
			t.Fatalf("expected price difference 43, got %.2f", ctx.ChargeAmount)
		}
		if ctx.CurrentSubscriptionId == 0 || ctx.CurrentEndTime == 0 {
			t.Fatal("expected current primary subscription context to be captured")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("resolve purchase: %v", err)
	}
}

func TestCompleteSubscriptionOrderRenewsExistingPrimarySubscription(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "renew_user", "plan_codex_basic")
	plan := createTestPlan(t, SubscriptionPlan{
		Title:            "天卡",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      1,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_basic",
		SortOrder:        10,
	})
	if _, err := AdminBindSubscription(user.Id, plan.Id, ""); err != nil {
		t.Fatalf("bind plan: %v", err)
	}

	subs, err := GetAllActiveUserSubscriptions(user.Id)
	if err != nil || len(subs) != 1 {
		t.Fatalf("expected one active subscription, got %d err=%v", len(subs), err)
	}
	originalEnd := subs[0].Subscription.EndTime

	err = DB.Transaction(func(tx *gorm.DB) error {
		_, _, err := CreateSubscriptionOrderTx(tx, user.Id, plan, "epay", "renew-order", time.Now().Unix())
		return err
	})
	if err != nil {
		t.Fatalf("create renewal order: %v", err)
	}
	if err := CompleteSubscriptionOrder("renew-order", ""); err != nil {
		t.Fatalf("complete renewal order: %v", err)
	}

	var activeSubs []UserSubscription
	if err := DB.Where("user_id = ? AND status = ?", user.Id, "active").Find(&activeSubs).Error; err != nil {
		t.Fatalf("load active subscriptions: %v", err)
	}
	if len(activeSubs) != 1 {
		t.Fatalf("expected one active subscription after renewal, got %d", len(activeSubs))
	}
	expectedEnd := time.Unix(originalEnd, 0).AddDate(0, 1, 0).Unix()
	if activeSubs[0].EndTime != expectedEnd {
		t.Fatalf("expected renewed end time %d, got %d", expectedEnd, activeSubs[0].EndTime)
	}
}

func TestCompleteSubscriptionOrderUpgradesPrimarySubscriptionWithoutExtendingExpiry(t *testing.T) {
	initSubscriptionBonusTestDB(t)

	user := createTestUser(t, "upgrade_complete_user", "plan_codex_basic")
	lowPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "天卡",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      1,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_basic",
		SortOrder:        10,
		TotalAmount:      10,
	})
	highPlan := createTestPlan(t, SubscriptionPlan{
		Title:            "CODING PLAN尊享套餐",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      44,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: SubscriptionResetNever,
		UpgradeGroup:     "plan_coding_all",
		SortOrder:        30,
		TotalAmount:      100,
	})
	if _, err := AdminBindSubscription(user.Id, lowPlan.Id, ""); err != nil {
		t.Fatalf("bind low plan: %v", err)
	}

	var before UserSubscription
	if err := DB.Where("user_id = ? AND status = ?", user.Id, "active").First(&before).Error; err != nil {
		t.Fatalf("load original subscription: %v", err)
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		order, ctx, err := CreateSubscriptionOrderTx(tx, user.Id, highPlan, "epay", "upgrade-order", time.Now().Unix())
		if err != nil {
			return err
		}
		if order.Money != 43 {
			t.Fatalf("expected upgrade order amount 43, got %.2f", order.Money)
		}
		if ctx.Mode != SubscriptionPurchaseModeUpgrade {
			t.Fatalf("expected upgrade mode, got %s", ctx.Mode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("create upgrade order: %v", err)
	}
	if err := CompleteSubscriptionOrder("upgrade-order", ""); err != nil {
		t.Fatalf("complete upgrade order: %v", err)
	}

	var activeSubs []UserSubscription
	if err := DB.Where("user_id = ? AND status = ?", user.Id, "active").Order("id asc").Find(&activeSubs).Error; err != nil {
		t.Fatalf("load active subscriptions: %v", err)
	}
	if len(activeSubs) != 1 {
		t.Fatalf("expected one active subscription after upgrade, got %d", len(activeSubs))
	}
	if activeSubs[0].PlanId != highPlan.Id {
		t.Fatalf("expected active plan %d, got %d", highPlan.Id, activeSubs[0].PlanId)
	}
	if activeSubs[0].EndTime != before.EndTime {
		t.Fatalf("expected upgrade to keep end time %d, got %d", before.EndTime, activeSubs[0].EndTime)
	}

	var cancelled UserSubscription
	if err := DB.Where("id = ?", before.Id).First(&cancelled).Error; err != nil {
		t.Fatalf("load original subscription after upgrade: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("expected original subscription to be cancelled, got %s", cancelled.Status)
	}

	updatedUser, err := GetUserById(user.Id, true)
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updatedUser.Group != "plan_coding_all" {
		t.Fatalf("expected user group plan_coding_all after upgrade, got %s", updatedUser.Group)
	}
}
