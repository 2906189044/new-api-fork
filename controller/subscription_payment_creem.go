package controller

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

type SubscriptionCreemPayRequest struct {
	PlanId int `json:"plan_id"`
}

func SubscriptionRequestCreemPay(c *gin.Context) {
	var req SubscriptionCreemPayRequest

	// Keep body for debugging consistency (like RequestCreemPay)
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("read subscription creem pay req body err: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "read query error"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if !plan.VisibleToUser || plan.StackableBonus {
		common.ApiErrorMsg(c, "该套餐不可直接购买")
		return
	}
	if plan.CreemProductId == "" {
		common.ApiErrorMsg(c, "该套餐未配置 CreemProductId")
		return
	}
	if setting.CreemWebhookSecret == "" && !setting.CreemTestMode {
		common.ApiErrorMsg(c, "Creem Webhook 未配置")
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}

	reference := "sub-creem-ref-" + randstr.String(6)
	referenceId := "sub_ref_" + common.Sha1([]byte(reference+time.Now().String()+user.Username))

	var chargeCtx *model.SubscriptionPurchaseContext
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		_, ctx, err := model.CreateSubscriptionOrderTx(tx, userId, plan, PaymentMethodCreem, referenceId, time.Now().Unix())
		if err != nil {
			return err
		}
		chargeCtx = ctx
		return nil
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if chargeCtx != nil && chargeCtx.ChargeAmount != plan.PriceAmount {
		_ = model.ExpireSubscriptionOrder(referenceId)
		common.ApiErrorMsg(c, "当前支付方式暂不支持套餐差价升级")
		return
	}

	// Reuse Creem checkout generator by building a lightweight product reference.
	currency := "USD"
	switch operation_setting.GetGeneralSetting().QuotaDisplayType {
	case operation_setting.QuotaDisplayTypeCNY:
		currency = "CNY"
	case operation_setting.QuotaDisplayTypeUSD:
		currency = "USD"
	default:
		currency = "USD"
	}
	product := &CreemProduct{
		ProductId: plan.CreemProductId,
		Name:      plan.Title,
		Price:     plan.PriceAmount,
		Currency:  currency,
		Quota:     0,
	}

	checkoutUrl, err := genCreemLink(referenceId, product, user.Email, user.Username)
	if err != nil {
		log.Printf("获取Creem支付链接失败: %v", err)
		_ = model.ExpireSubscriptionOrder(referenceId)
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"checkout_url": checkoutUrl,
			"order_id":     referenceId,
		},
	})
}
