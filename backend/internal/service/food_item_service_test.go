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

func seedFood(t *testing.T, db *gorm.DB, familyID, creatorID uint, quantity float64) *model.FoodItem {
	t.Helper()
	return seedFoodWithPrice(t, db, familyID, creatorID, quantity, nil)
}

func seedFoodWithPrice(t *testing.T, db *gorm.DB, familyID, creatorID uint, quantity float64, unitPrice *float64) *model.FoodItem {
	t.Helper()
	expiry := time.Now().AddDate(0, 0, 5)
	item := &model.FoodItem{
		FamilyID: familyID, Name: "测试食品", Category: constants.FoodCategoryDairy,
		Quantity: quantity, Unit: "盒", UnitPrice: unitPrice, ShelfLifeDays: 5, StorageLocation: constants.StorageFridge,
		Status: constants.FreshnessFresh, CreatorID: creatorID, ExpiryDate: &expiry,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed food: %v", err)
	}
	return item
}

func TestFoodItemService_Consume(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	svc := NewFoodItemService(foodRepo, consumeRepo, familySvc, util.NewFoodCalculator(), testLogger())
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	item := seedFood(t, db, group.ID, 1, 3)

	record, err := svc.Consume(ctx, 1, item.ID, 1.5, time.Now())
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if record.Quantity != 1.5 {
		t.Fatalf("record quantity = %v, want 1.5", record.Quantity)
	}

	got, err := foodRepo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Quantity != 1.5 {
		t.Fatalf("remaining quantity = %v, want 1.5", got.Quantity)
	}

	// 全部消耗后状态流转为 consumed。
	if _, err := svc.Consume(ctx, 1, item.ID, 1.5, time.Now()); err != nil {
		t.Fatalf("second Consume() error = %v", err)
	}
	got, _ = foodRepo.FindByID(item.ID)
	if got.Status != constants.FreshnessConsumed {
		t.Fatalf("status = %s, want %s", got.Status, constants.FreshnessConsumed)
	}

	// 超量消耗应失败。
	if _, err := svc.Consume(ctx, 1, item.ID, 0.1, time.Now()); err == nil {
		t.Fatal("expected error for consuming consumed/empty food")
	}
}

// 消耗时快照当时单价与金额；之后修改食品价格不影响历史消耗记录。
func TestFoodItemService_ConsumePriceSnapshot(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	calculator := util.NewFoodCalculator()
	svc := NewFoodItemService(foodRepo, consumeRepo, familySvc, calculator, testLogger())
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 未填单价：按默认 15 元/单位快照。
	noPrice := seedFoodWithPrice(t, db, group.ID, 1, 5, nil)
	rec, err := svc.Consume(ctx, 1, noPrice.ID, 2, time.Now())
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if rec.UnitPrice != constants.DefaultUnitPrice || rec.Amount != 30 {
		t.Fatalf("default price snapshot = (%v, %v), want (15, 30)", rec.UnitPrice, rec.Amount)
	}

	// 填写单价：按当时单价快照，改价后历史记录不变。
	price := 12.75
	withPrice := seedFoodWithPrice(t, db, group.ID, 1, 5, &price)
	rec, err = svc.Consume(ctx, 1, withPrice.ID, 2, time.Now())
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if rec.UnitPrice != 12.75 || rec.Amount != 25.5 {
		t.Fatalf("price snapshot = (%v, %v), want (12.75, 25.5)", rec.UnitPrice, rec.Amount)
	}

	newPrice := 20.0
	if _, err := svc.Update(ctx, 1, withPrice.ID, CreateFoodInput{FamilyID: group.ID, UnitPrice: &newPrice}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	history, err := consumeRepo.ListByFood(withPrice.ID)
	if err != nil {
		t.Fatalf("ListByFood() error = %v", err)
	}
	if len(history) != 1 || history[0].UnitPrice != 12.75 || history[0].Amount != 25.5 {
		t.Fatalf("history changed after price update: %+v", history)
	}
}

// 采购单价校验：不能为负，最多保留两位小数。
func TestFoodItemService_UnitPriceValidation(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	svc := NewFoodItemService(foodRepo, consumeRepo, familySvc, util.NewFoodCalculator(), testLogger())
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	base := CreateFoodInput{FamilyID: group.ID, Name: "测试食品", Category: constants.FoodCategoryDairy, Quantity: 1}

	negative := -1.0
	if _, err := svc.Create(ctx, 1, withPrice(base, &negative)); err == nil {
		t.Fatal("expected error for negative unit price")
	}
	tooPrecise := 1.234
	if _, err := svc.Create(ctx, 1, withPrice(base, &tooPrecise)); err == nil {
		t.Fatal("expected error for unit price with more than 2 decimals")
	}
	valid := 9.99
	item, err := svc.Create(ctx, 1, withPrice(base, &valid))
	if err != nil {
		t.Fatalf("Create() with valid price error = %v", err)
	}
	if item.UnitPrice == nil || *item.UnitPrice != 9.99 {
		t.Fatalf("unit price = %v, want 9.99", item.UnitPrice)
	}
	// 编辑时清空单价：回到未填写状态（nil）。
	if _, err := svc.Update(ctx, 1, item.ID, CreateFoodInput{FamilyID: group.ID}); err != nil {
		t.Fatalf("Update() clear price error = %v", err)
	}
	got, _ := foodRepo.FindByID(item.ID)
	if got.UnitPrice != nil {
		t.Fatalf("unit price after clear = %v, want nil", *got.UnitPrice)
	}
}

func withPrice(in CreateFoodInput, price *float64) CreateFoodInput {
	in.UnitPrice = price
	return in
}
