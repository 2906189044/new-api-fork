package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func TestSubscriptionPlanIncludesDisplayFields(t *testing.T) {
	planType := reflect.TypeOf(model.SubscriptionPlan{})

	if _, ok := planType.FieldByName("DisplayConfig"); !ok {
		t.Fatal("expected SubscriptionPlan to include DisplayConfig field")
	}

	if _, ok := planType.FieldByName("DisplayNotes"); !ok {
		t.Fatal("expected SubscriptionPlan to include DisplayNotes field")
	}
}

func TestSubscriptionPlanIncludesVisibleToUserField(t *testing.T) {
	planType := reflect.TypeOf(model.SubscriptionPlan{})
	if _, ok := planType.FieldByName("VisibleToUser"); !ok {
		t.Fatal("expected SubscriptionPlan to include VisibleToUser field")
	}
}

func initSubscriptionTestDB(t *testing.T) {
	t.Helper()
	common.IsMasterNode = true
	common.SQLitePath = filepath.Join(t.TempDir(), fmt.Sprintf("%s.sqlite", t.Name()))
	common.UsingSQLite = true
	if err := model.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(common.SQLitePath)
	})
}

func TestGetSubscriptionPlansFiltersHiddenPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initSubscriptionTestDB(t)

	visiblePlan := model.SubscriptionPlan{
		Title:         "Visible plan",
		Enabled:       true,
		VisibleToUser: true,
	}
	hiddenPlan := model.SubscriptionPlan{
		Title:         "Hidden plan",
		Enabled:       true,
		VisibleToUser: false,
	}
	if err := model.DB.Create(&visiblePlan).Error; err != nil {
		t.Fatalf("create visible plan: %v", err)
	}
	if err := model.DB.Model(&model.SubscriptionPlan{}).Create(map[string]any{
		"title":            hiddenPlan.Title,
		"price_amount":     0,
		"currency":         "CNY",
		"duration_unit":    "month",
		"duration_value":   1,
		"enabled":          true,
		"visible_to_user":  false,
		"quota_reset_period": "never",
	}).Error; err != nil {
		t.Fatalf("create hidden plan: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	GetSubscriptionPlans(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Success bool                  `json:"success"`
		Data    []SubscriptionPlanDTO `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected exactly one visible plan, got %d", len(resp.Data))
	}
	if resp.Data[0].Plan.Title != visiblePlan.Title {
		t.Fatalf("expected visible plan %q, got %q", visiblePlan.Title, resp.Data[0].Plan.Title)
	}
}

func TestAdminListSubscriptionPlansIncludesHiddenPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initSubscriptionTestDB(t)

	visiblePlan := model.SubscriptionPlan{
		Title:         "Visible admin plan",
		Enabled:       true,
		VisibleToUser: true,
	}
	hiddenPlan := model.SubscriptionPlan{
		Title:         "Hidden admin plan",
		Enabled:       true,
		VisibleToUser: false,
	}
	if err := model.DB.Create(&visiblePlan).Error; err != nil {
		t.Fatalf("create visible plan: %v", err)
	}
	if err := model.DB.Model(&model.SubscriptionPlan{}).Create(map[string]any{
		"title":            hiddenPlan.Title,
		"price_amount":     0,
		"currency":         "CNY",
		"duration_unit":    "month",
		"duration_value":   1,
		"enabled":          true,
		"visible_to_user":  false,
		"quota_reset_period": "never",
	}).Error; err != nil {
		t.Fatalf("create hidden plan: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	AdminListSubscriptionPlans(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Success bool                  `json:"success"`
		Data    []SubscriptionPlanDTO `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected both visible and hidden plans, got %d", len(resp.Data))
	}
}

func TestAdminCreateSubscriptionPlanPersistsHiddenVisibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initSubscriptionTestDB(t)

	reqBody := []byte(`{"plan":{"title":"Hidden from user","price_amount":9.9,"enabled":true,"visible_to_user":false}}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/subscription/admin/plans", bytes.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	AdminCreateSubscriptionPlan(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var created model.SubscriptionPlan
	if err := model.DB.Where("title = ?", "Hidden from user").First(&created).Error; err != nil {
		t.Fatalf("load created plan: %v", err)
	}
	if created.VisibleToUser {
		t.Fatal("expected created hidden plan to persist visible_to_user=false")
	}
}
