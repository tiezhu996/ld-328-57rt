package util

import (
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
)

func TestFoodCalculator_ResolveUnitPrice(t *testing.T) {
	calc := NewFoodCalculator()
	price := 8.90
	zero := 0.0
	tests := []struct {
		name  string
		price *float64
		want  float64
	}{
		{name: "nil falls back to default 15", price: nil, want: constants.DefaultUnitPrice},
		{name: "explicit price", price: &price, want: 8.90},
		{name: "zero price is respected", price: &zero, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.ResolveUnitPrice(tt.price); got != tt.want {
				t.Fatalf("ResolveUnitPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFoodCalculator_WasteAmount(t *testing.T) {
	calc := NewFoodCalculator()
	price := 12.75
	tests := []struct {
		name     string
		quantity float64
		price    *float64
		want     float64
	}{
		{name: "nil price uses default 15", quantity: 2, price: nil, want: 30},
		{name: "quantity times current price", quantity: 3, price: &price, want: 38.25},
		{name: "rounds to two decimals", quantity: 0.1, price: &price, want: 1.28},
		{name: "zero quantity", quantity: 0, price: &price, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.WasteAmount(tt.quantity, tt.price); got != tt.want {
				t.Fatalf("WasteAmount() = %v, want %v", got, tt.want)
			}
		})
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
