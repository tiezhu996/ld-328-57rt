package util

import (
	"testing"
	"time"
)

func TestFoodCalculator_ValidateUnitPrice(t *testing.T) {
	calc := NewFoodCalculator()
	tests := []struct {
		name  string
		price float64
		want  bool
	}{
		{name: "zero ok", price: 0, want: true},
		{name: "integer ok", price: 15, want: true},
		{name: "two decimals ok", price: 12.55, want: true},
		{name: "one decimal ok", price: 8.8, want: true},
		{name: "negative rejected", price: -0.01, want: false},
		{name: "three decimals rejected", price: 9.999, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.ValidateUnitPrice(tt.price); got != tt.want {
				t.Fatalf("ValidateUnitPrice(%v) = %v, want %v", tt.price, got, tt.want)
			}
		})
	}
}

func TestFoodCalculator_EffectiveUnitPrice(t *testing.T) {
	calc := NewFoodCalculator()
	price := 12.5
	if got := calc.EffectiveUnitPrice(nil); got != 15.0 {
		t.Fatalf("EffectiveUnitPrice(nil) = %v, want 15", got)
	}
	if got := calc.EffectiveUnitPrice(&price); got != 12.5 {
		t.Fatalf("EffectiveUnitPrice(12.5) = %v, want 12.5", got)
	}
}

func TestFoodCalculator_Amount(t *testing.T) {
	calc := NewFoodCalculator()
	price := 12.5
	if got := calc.Amount(2, &price); got != 25.0 {
		t.Fatalf("Amount(2, 12.5) = %v, want 25", got)
	}
	// 未填单价按默认 15 元/单位。
	if got := calc.Amount(1.5, nil); got != 22.5 {
		t.Fatalf("Amount(1.5, nil) = %v, want 22.5", got)
	}
	// 金额保留两位小数。
	p := 3.33
	if got := calc.Amount(3, &p); got != 9.99 {
		t.Fatalf("Amount(3, 3.33) = %v, want 9.99", got)
	}
}

func TestFoodCalculator_RemainingDays(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	yesterday := time.Now().AddDate(0, 0, -1)
	calc := NewFoodCalculator()
	tests := []struct {
		name       string
		expiry     *time.Time
		wantRemain int
	}{
		{name: "tomorrow", expiry: &tomorrow, wantRemain: 1},
		{name: "yesterday", expiry: &yesterday, wantRemain: -1},
		{name: "nil", expiry: nil, wantRemain: 3650},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.RemainingDays(tt.expiry); got != tt.wantRemain {
				t.Fatalf("RemainingDays() = %d, want %d", got, tt.wantRemain)
			}
		})
	}
}

func TestFoodCalculator_ComputeFreshness(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	tenDays := time.Now().AddDate(0, 0, 10)
	yesterday := time.Now().AddDate(0, 0, -1)
	calc := NewFoodCalculator()
	tests := []struct {
		name   string
		status string
		expiry *time.Time
		want   string
	}{
		{name: "consumed wins", status: "consumed", expiry: &tenDays, want: "consumed"},
		{name: "expiring within 3 days", status: "fresh", expiry: &tomorrow, want: "expiring"},
		{name: "fresh far away", status: "fresh", expiry: &tenDays, want: "fresh"},
		{name: "expired", status: "fresh", expiry: &yesterday, want: "expired"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.ComputeFreshness(tt.status, tt.expiry); got != tt.want {
				t.Fatalf("ComputeFreshness() = %s, want %s", got, tt.want)
			}
		})
	}
}
