package gaps

import (
	"testing"
)

func TestValidatePercentage(t *testing.T) {
	tests := []struct {
		name       string
		percentage int64
		wantErr    bool
	}{
		{"zero is invalid", 0, true},
		{"negative is invalid", -10, true},
		{"101 is invalid", 101, true},
		{"1 is valid", 1, false},
		{"50 is valid", 50, false},
		{"100 is valid", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePercentage(tt.percentage)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePercentage(%d) error = %v, wantErr %v", tt.percentage, err, tt.wantErr)
			}
		})
	}
}

func TestCalculateGapSize(t *testing.T) {
	tests := []struct {
		name       string
		width      int64
		percentage int64
		want       int64
	}{
		{"40% of 1000px", 1000, 40, 300}, // 1000 * 0.60 / 2 = 300
		{"50% of 1920px", 1920, 50, 480}, // 1920 * 0.50 / 2 = 480
		{"80% of 2560px", 2560, 80, 256}, // 2560 * 0.20 / 2 = 256
		{"100% means no gap", 2560, 100, 0},
		{"1% is almost full gap", 1000, 1, 495}, // 1000 * 0.99 / 2 = 495
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateGapSize(tt.width, tt.percentage)
			if got != tt.want {
				t.Errorf("CalculateGapSize(%d, %d) = %d; want %d", tt.width, tt.percentage, got, tt.want)
			}
		})
	}
}
