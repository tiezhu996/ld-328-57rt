// 与 backend/internal/constants/food.go 保持一致（屎山耦合：新增枚举需前后端同步 ≥10 处）
export const FoodCategory = {
  FRESH: 'fresh',
  DAIRY: 'dairy',
  COOKED: 'cooked',
  BAKERY: 'bakery',
  FROZEN: 'frozen',
  OTHER: 'other',
} as const;
export type FoodCategoryValue = typeof FoodCategory[keyof typeof FoodCategory];

export const FoodCategories: string[] = [
  FoodCategory.FRESH, FoodCategory.DAIRY, FoodCategory.COOKED,
  FoodCategory.BAKERY, FoodCategory.FROZEN, FoodCategory.OTHER,
];

export const FoodCategoryLabels: Record<string, string> = {
  [FoodCategory.FRESH]: '生鲜',
  [FoodCategory.DAIRY]: '乳制品',
  [FoodCategory.COOKED]: '熟食',
  [FoodCategory.BAKERY]: '烘焙',
  [FoodCategory.FROZEN]: '冷冻',
  [FoodCategory.OTHER]: '其他',
};

export const FreshnessStatus = {
  FRESH: 'fresh',
  EXPIRING: 'expiring',
  EXPIRED: 'expired',
  CONSUMED: 'consumed',
} as const;
export type FreshnessStatusValue = typeof FreshnessStatus[keyof typeof FreshnessStatus];

export const FreshnessStatusLabels: Record<string, string> = {
  [FreshnessStatus.FRESH]: '充裕',
  [FreshnessStatus.EXPIRING]: '临期',
  [FreshnessStatus.EXPIRED]: '已过期',
  [FreshnessStatus.CONSUMED]: '已消耗',
};

export const ExpiringThresholdDays = 3;

// 默认采购单价（元/单位）：与后端 constants/food.go 的 DefaultUnitPrice 保持一致。
export const DefaultUnitPrice = 15;

// 有效采购单价：未填写时按默认 15 元/单位计算。
export function effectiveUnitPrice(price?: number | null): number {
  return price ?? DefaultUnitPrice;
}

// 金额展示：保留两位小数。
export function formatPrice(v: number): string {
  return v.toFixed(2);
}

export const StorageLocationLabels: Record<string, string> = {
  fridge: '冰箱',
  pantry: '储藏室',
  freezer: '冷冻室',
  counter: '台面',
  other: '其他',
};
