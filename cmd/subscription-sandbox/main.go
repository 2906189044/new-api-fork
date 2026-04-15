package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func main() {
	dbPath := flag.String("db", filepath.Join(".local", "subscription-sandbox.sqlite"), "sqlite path for sandbox db")
	flag.Parse()

	if err := run(*dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "subscription sandbox seed failed: %v\n", err)
		os.Exit(1)
	}
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
		return err
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
		return err
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
		return err
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
		return err
	}

	if _, err := createUser("sandbox_new", "default"); err != nil {
		return err
	}
	basicUser, err := createUser("sandbox_basic", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(basicUser.Id, basicPlan.Id, ""); err != nil {
		return err
	}
	advancedUser, err := createUser("sandbox_advanced", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(advancedUser.Id, advancedPlan.Id, ""); err != nil {
		return err
	}
	codingUser, err := createUser("sandbox_coding", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(codingUser.Id, codingPlan.Id, ""); err != nil {
		return err
	}
	bonusUser, err := createUser("sandbox_bonus", "default")
	if err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(bonusUser.Id, codingPlan.Id, ""); err != nil {
		return err
	}
	if _, err := model.AdminBindSubscription(bonusUser.Id, bonusPlan.Id, ""); err != nil {
		return err
	}

	fmt.Printf("Seeded local subscription sandbox DB: %s\n", dbPath)
	fmt.Println("Login credentials:")
	fmt.Println("  root / 123456")
	fmt.Println("  sandbox_new / password123")
	fmt.Println("  sandbox_basic / password123")
	fmt.Println("  sandbox_advanced / password123")
	fmt.Println("  sandbox_coding / password123")
	fmt.Println("  sandbox_bonus / password123")
	fmt.Println("Seeded plans:")
	fmt.Printf("  %d 天卡\n", basicPlan.Id)
	fmt.Printf("  %d CODECS专属套餐\n", advancedPlan.Id)
	fmt.Printf("  %d CODING PLAN尊享套餐\n", codingPlan.Id)
	fmt.Printf("  %d 管理员专用 GPT 日限额计划 (hidden stackable bonus)\n", bonusPlan.Id)
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
