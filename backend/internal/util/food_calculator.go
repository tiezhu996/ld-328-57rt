package util

import (
	"math"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
)

// FoodCalculator 食品保质期计算：剩余天数、到期日、新鲜度状态。
type FoodCalculator struct{}

// NewFoodCalculator 构造计算器。
func NewFoodCalculator() *FoodCalculator { return &FoodCalculator{} }

// CalculateExpiryDate 根据生产日期与保质期天数计算到期日（已开封则按开封时间折算）。
func (c *FoodCalculator) CalculateExpiryDate(productionDate *time.Time, shelfLifeDays int, openedAt *time.Time) *time.Time {
	if shelfLifeDays <= 0 {
		return nil
	}
	base := time.Now()
	if productionDate != nil {
		base = *productionDate
	}
	expiry := base.AddDate(0, 0, shelfLifeDays)
	if openedAt != nil {
		opened := openedAt.AddDate(0, 0, 7) // 开封后建议 7 天内食用
		if opened.Before(expiry) {
			expiry = opened
		}
	}
	return &expiry
}

// RemainingDays 计算剩余自然日（今天到期=0，明天=1，昨天=-1）。
func (c *FoodCalculator) RemainingDays(expiryDate *time.Time) int {
	if expiryDate == nil {
		return 365 * 10
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expiry := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())
	return int(expiry.Sub(today).Hours() / 24)
}

// ComputeFreshness 计算新鲜度状态（已消耗优先返回 consumed）。
func (c *FoodCalculator) ComputeFreshness(status string, expiryDate *time.Time) string {
	if status == constants.FreshnessConsumed {
		return constants.FreshnessConsumed
	}
	days := c.RemainingDays(expiryDate)
	switch {
	case days < 0:
		return constants.FreshnessExpired
	case days <= constants.ExpiringThresholdDays:
		return constants.FreshnessExpiring
	default:
		return constants.FreshnessFresh
	}
}

// ValidateUnitPrice 校验采购单价：不能为负，最多保留两位小数。
func (c *FoodCalculator) ValidateUnitPrice(price float64) bool {
	if price < 0 {
		return false
	}
	return math.Abs(price*100-math.Round(price*100)) < 1e-9
}

// EffectiveUnitPrice 有效采购单价：未填写时按默认单价（15 元/单位）计算。
func (c *FoodCalculator) EffectiveUnitPrice(price *float64) float64 {
	if price == nil {
		return constants.DefaultUnitPrice
	}
	return *price
}

// Amount 金额估算：数量 × 有效单价，保留两位小数（消耗快照与浪费统计共用同一口径）。
func (c *FoodCalculator) Amount(quantity float64, price *float64) float64 {
	return math.Round(quantity*c.EffectiveUnitPrice(price)*100) / 100
}
