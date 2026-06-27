package domain_test

import (
	"testing"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/domain"
)

func TestMoney_MulFloat_RoundsHalfUp(t *testing.T) {
	tests := []struct {
		name   string
		amount domain.Money
		factor float64
		want   domain.Money
	}{
		{"half rounds up", 1, 5000.5, 5001},
		{"below half rounds down", 1, 5000.4, 5000},
		{"distance fee 695km at 4 cents", 4, 695, 2780},
		{"weight fee 2.5kg at 1000 cents", 1000, 2.5, 2500},
		{"express surcharge on 10000 at 0.5", 10000, 0.5, 5000},
		{"priority surcharge on 10000 at 0.15", 10000, 0.15, 1500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.amount.MulFloat(tt.factor); got != tt.want {
				t.Errorf("MulFloat(%v) = %d, want %d", tt.factor, got.Cents(), tt.want.Cents())
			}
		})
	}
}

func TestMoney_AddSub(t *testing.T) {
	if got := domain.Money(5000).Add(2780).Add(2500); got != 10280 {
		t.Errorf("Add chain = %d, want 10280", got.Cents())
	}
	if got := domain.Money(13500).Sub(1000); got != 12500 {
		t.Errorf("Sub = %d, want 12500", got.Cents())
	}
}

func TestCentsFromFloat_RoundsHalfUp(t *testing.T) {
	if got := domain.CentsFromFloat(1000.5); got != 1001 {
		t.Errorf("CentsFromFloat(1000.5) = %d, want 1001", got.Cents())
	}
	if got := domain.CentsFromFloat(1000.4); got != 1000 {
		t.Errorf("CentsFromFloat(1000.4) = %d, want 1000", got.Cents())
	}
}
