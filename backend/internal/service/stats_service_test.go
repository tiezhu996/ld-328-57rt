package service

import (
	"context"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

func seedExpiredFood(t *testing.T, db *gorm.DB, familyID, creatorID uint, name string, quantity float64, unitPrice *float64) *model.FoodItem {
	t.Helper()
	expiry := time.Now().AddDate(0, 0, -2)
	item := &model.FoodItem{
		FamilyID: familyID, Name: name, Category: constants.FoodCategoryDairy,
		Quantity: quantity, Unit: "盒", UnitPrice: unitPrice, ShelfLifeDays: 5, StorageLocation: constants.StorageFridge,
		Status: constants.FreshnessFresh, CreatorID: creatorID, ExpiryDate: &expiry,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed expired food: %v", err)
	}
	return item
}

// 统计页浪费金额口径：剩余数量 × 当前单价，未填单价按默认 15 元/单位。
func TestStatsService_WasteAmountUsesCurrentUnitPrice(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	memberSvc := NewFamilyMemberService(memberRepo, testLogger())
	calculator := util.NewFoodCalculator()
	svc := NewStatsService(foodRepo, consumeRepo, notifyRepo, familySvc, memberSvc, calculator, testLogger())
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	price := 12.75
	seedExpiredFood(t, db, group.ID, 1, "过期有单价", 3, &price) // 3 × 12.75 = 38.25
	seedExpiredFood(t, db, group.ID, 1, "过期无单价", 2, nil)     // 2 × 15 = 30
	seedFood(t, db, group.ID, 1, 5)                             // 未过期，不计入

	data, err := svc.Statistics(ctx, 1, group.ID, "")
	if err != nil {
		t.Fatalf("Statistics() error = %v", err)
	}
	if data.WasteAmount != 68.25 {
		t.Fatalf("waste amount = %v, want 68.25", data.WasteAmount)
	}
	if len(data.TopWasted) != 2 {
		t.Fatalf("top wasted len = %d, want 2", len(data.TopWasted))
	}
	amounts := map[string]float64{}
	for _, row := range data.TopWasted {
		amounts[row.Name] = row.Amount
	}
	if amounts["过期有单价"] != 38.25 || amounts["过期无单价"] != 30 {
		t.Fatalf("top wasted amounts = %v, want {过期有单价:38.25 过期无单价:30}", amounts)
	}
}
