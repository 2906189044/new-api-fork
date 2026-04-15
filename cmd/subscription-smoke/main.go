package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func main() {
	dbPath := flag.String("db", filepath.Join(".local", "subscription-smoke.sqlite"), "sqlite path for smoke test db")
	flag.Parse()

	if err := run(*dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "subscription smoke failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("subscription smoke passed")
}

func run(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	_ = os.Remove(dbPath)

	common.IsMasterNode = true
	common.SQLitePath = dbPath
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	if err := model.InitDB(); err != nil {
		return err
	}
	model.LOG_DB = model.DB

	basicPlan, advancedPlan, codingPlan, bonusPlan, err := createDefaultPlans()
	if err != nil {
		return err
	}

	if err := scenarioRenewAndUpgrade(basicPlan, codingPlan); err != nil {
		return err
	}
	if err := scenarioDowngradeRejected(advancedPlan, codingPlan); err != nil {
		return err
	}
	if err := scenarioHiddenBonusIsGPTOnly(codingPlan, bonusPlan); err != nil {
		return err
	}
	return nil
}

func createUser(username, group string) (*model.User, error) {
	user := &model.User{
		Username:    username,
		Password:    "password123",
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       group,
	}
	if err := user.Insert(0); err != nil {
		return nil, err
	}
	if err := model.DB.Where("username = ?", username).First(user).Error; err != nil {
		return nil, err
	}
	if group != "default" {
		if err := model.DB.Model(user).Update("group", group).Error; err != nil {
			return nil, err
		}
		user.Group = group
	}
	return user, nil
}

func createPlan(plan model.SubscriptionPlan) (*model.SubscriptionPlan, error) {
	if err := model.DB.Create(&plan).Error; err != nil {
		return nil, err
	}
	model.InvalidateSubscriptionPlanCache(plan.Id)
	return &plan, nil
}

func createDefaultPlans() (*model.SubscriptionPlan, *model.SubscriptionPlan, *model.SubscriptionPlan, *model.SubscriptionPlan, error) {
	basicPlan, err := createPlan(model.SubscriptionPlan{
		Title:            "天卡",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      1,
		DurationUnit:     model.SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: model.SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_basic",
		SortOrder:        10,
		TotalAmount:      10,
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	advancedPlan, err := createPlan(model.SubscriptionPlan{
		Title:            "CODECS专属套餐",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      19,
		DurationUnit:     model.SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: model.SubscriptionResetNever,
		UpgradeGroup:     "plan_codex_advanced",
		SortOrder:        20,
		TotalAmount:      30,
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	codingPlan, err := createPlan(model.SubscriptionPlan{
		Title:            "CODING PLAN尊享套餐",
		VisibleToUser:    true,
		Enabled:          true,
		PriceAmount:      44,
		DurationUnit:     model.SubscriptionDurationMonth,
		DurationValue:    1,
		QuotaResetPeriod: model.SubscriptionResetNever,
		UpgradeGroup:     "plan_coding_all",
		SortOrder:        30,
		TotalAmount:      100,
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	bonusPlan, err := createPlan(model.SubscriptionPlan{
		Title:            "管理员专用 GPT 日限额计划",
		VisibleToUser:    false,
		Enabled:          true,
		StackableBonus:   true,
		BonusModelScope:  "gpt_series",
		DurationUnit:     model.SubscriptionDurationDay,
		DurationValue:    10,
		QuotaResetPeriod: model.SubscriptionResetDaily,
		TotalAmount:      10,
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return basicPlan, advancedPlan, codingPlan, bonusPlan, nil
}

func scenarioRenewAndUpgrade(basicPlan, codingPlan *model.SubscriptionPlan) error {
	user, err := createUser("smoke_renew_upgrade", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(user.Id, basicPlan.Id, ""); err != nil {
		return err
	}

	var before model.UserSubscription
	if err := model.DB.Where("user_id = ? AND status = ?", user.Id, "active").First(&before).Error; err != nil {
		return err
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		_, _, err := model.CreateSubscriptionOrderTx(tx, user.Id, basicPlan, "epay", "smoke-renew", time.Now().Unix())
		return err
	}); err != nil {
		return fmt.Errorf("renew order create: %w", err)
	}
	if err := model.CompleteSubscriptionOrder("smoke-renew", ""); err != nil {
		return fmt.Errorf("renew order complete: %w", err)
	}

	var renewed model.UserSubscription
	if err := model.DB.Where("user_id = ? AND status = ?", user.Id, "active").First(&renewed).Error; err != nil {
		return err
	}
	if renewed.Id != before.Id {
		return errors.New("renew should keep the same active main subscription record")
	}
	expectedRenewEnd := time.Unix(before.EndTime, 0).AddDate(0, 1, 0).Unix()
	if renewed.EndTime != expectedRenewEnd {
		return fmt.Errorf("renew end_time mismatch: got %d want %d", renewed.EndTime, expectedRenewEnd)
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		order, ctx, err := model.CreateSubscriptionOrderTx(tx, user.Id, codingPlan, "epay", "smoke-upgrade", time.Now().Unix())
		if err != nil {
			return err
		}
		if ctx.Mode != model.SubscriptionPurchaseModeUpgrade {
			return fmt.Errorf("upgrade mode mismatch: %s", ctx.Mode)
		}
		if order.Money != 43 {
			return fmt.Errorf("upgrade charge mismatch: %.2f", order.Money)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("upgrade order create: %w", err)
	}
	if err := model.CompleteSubscriptionOrder("smoke-upgrade", ""); err != nil {
		return fmt.Errorf("upgrade order complete: %w", err)
	}

	var activeSubs []model.UserSubscription
	if err := model.DB.Where("user_id = ? AND status = ?", user.Id, "active").Order("id asc").Find(&activeSubs).Error; err != nil {
		return err
	}
	if len(activeSubs) != 1 {
		return fmt.Errorf("upgrade should leave one active main subscription, got %d", len(activeSubs))
	}
	if activeSubs[0].PlanId != codingPlan.Id {
		return fmt.Errorf("upgrade plan mismatch: got %d want %d", activeSubs[0].PlanId, codingPlan.Id)
	}
	if activeSubs[0].EndTime != renewed.EndTime {
		return fmt.Errorf("upgrade should keep expiry: got %d want %d", activeSubs[0].EndTime, renewed.EndTime)
	}

	updatedUser, err := model.GetUserById(user.Id, true)
	if err != nil {
		return err
	}
	if updatedUser.Group != "plan_coding_all" {
		return fmt.Errorf("upgrade user group mismatch: %s", updatedUser.Group)
	}
	fmt.Println("ok: renew and upgrade flow")
	return nil
}

func scenarioDowngradeRejected(advancedPlan, codingPlan *model.SubscriptionPlan) error {
	user, err := createUser("smoke_downgrade", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(user.Id, codingPlan.Id, ""); err != nil {
		return err
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		_, err := model.ResolveSubscriptionPurchaseTx(tx, user.Id, advancedPlan)
		return err
	})
	if !errors.Is(err, model.ErrSubscriptionPlanDowngrade) {
		return fmt.Errorf("expected downgrade rejection, got %v", err)
	}
	fmt.Println("ok: downgrade rejected")
	return nil
}

func scenarioHiddenBonusIsGPTOnly(codingPlan, bonusPlan *model.SubscriptionPlan) error {
	user, err := createUser("smoke_bonus", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(user.Id, codingPlan.Id, ""); err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(user.Id, bonusPlan.Id, ""); err != nil {
		return err
	}

	gptResult, err := model.PreConsumeUserSubscription("smoke-gpt", user.Id, "gpt-5.4", 0, 5)
	if err != nil {
		return err
	}
	claudeResult, err := model.PreConsumeUserSubscription("smoke-claude", user.Id, "claude-sonnet-4.5", 0, 5)
	if err != nil {
		return err
	}
	if gptResult.UserSubscriptionId == claudeResult.UserSubscriptionId {
		return errors.New("bonus GPT subscription should not be reused for claude")
	}

	gptPlanInfo, err := model.GetSubscriptionPlanInfoByUserSubscriptionId(gptResult.UserSubscriptionId)
	if err != nil {
		return err
	}
	if gptPlanInfo.PlanTitle != bonusPlan.Title {
		return fmt.Errorf("gpt plan mismatch: %s", gptPlanInfo.PlanTitle)
	}
	claudePlanInfo, err := model.GetSubscriptionPlanInfoByUserSubscriptionId(claudeResult.UserSubscriptionId)
	if err != nil {
		return err
	}
	if claudePlanInfo.PlanTitle != codingPlan.Title {
		return fmt.Errorf("claude plan mismatch: %s", claudePlanInfo.PlanTitle)
	}
	fmt.Println("ok: hidden GPT bonus is GPT-only")
	return nil
}
