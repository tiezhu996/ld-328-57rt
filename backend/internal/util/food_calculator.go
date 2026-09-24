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

// ResolveUnitPrice 解析有效采购单价：未填写（nil）时按默认单价 15 元/单位估算。
func (c *FoodCalculator) ResolveUnitPrice(unitPrice *float64) float64 {
	if unitPrice == nil {
		return constants.DefaultUnitPrice
	}
	return *unitPrice
}

// RoundMoney 金额四舍五入保留两位小数。
func (c *FoodCalculator) RoundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

// WasteAmount 浪费金额 = 剩余数量 × 当前有效单价（未填单价按默认 15 元/单位），保留两位小数。
// 统计页、过期清单与导出报表必须统一走本口径。
func (c *FoodCalculator) WasteAmount(quantity float64, unitPrice *float64) float64 {
	return c.RoundMoney(quantity * c.ResolveUnitPrice(unitPrice))
}
