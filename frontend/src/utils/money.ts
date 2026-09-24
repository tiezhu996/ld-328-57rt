import { DefaultUnitPrice } from '../constants/food';

// resolveUnitPrice 解析有效采购单价：未填写（null/undefined）时按默认 15 元/单位估算。
export function resolveUnitPrice(unitPrice?: number | null): number {
  return unitPrice ?? DefaultUnitPrice;
}

// wasteAmount 浪费金额 = 剩余数量 × 当前有效单价（与后端统计口径一致）。
export function wasteAmount(quantity: number, unitPrice?: number | null): number {
  return Math.round(quantity * resolveUnitPrice(unitPrice) * 100) / 100;
}

// formatMoney 金额保留两位小数展示。
export function formatMoney(v: number): string {
  return v.toFixed(2);
}
