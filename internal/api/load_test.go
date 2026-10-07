package api

import "testing"

func TestParseGRESCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"simple", "gpu:4", 4},
		{"typed", "gpu:a100:4", 4},
		{"multi", "gpu:4,gpu:2", 6},
		{"suffix S", "gpu:b200:8(S:0-1)", 8},
		{"suffix IDX", "gpu:b200:8(IDX:0-7)", 8},
		{"mixed gres", "tmpsize:11T,gpu:b200:8(S:0-1)", 8},
		{"used gres", "gpu:b200:8(IDX:0-7),tmpsize:9661528932352", 8},
		{"no gpu", "tmpsize:11T", 0},
		{"complex", "gpu:a100:2,gpu:h100:4(S:0-3),gpu:4", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseGRESCount(tt.in); got != tt.want {
				t.Errorf("ParseGRESCount(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
