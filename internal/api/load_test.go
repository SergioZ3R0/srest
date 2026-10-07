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

func TestParseTRESGPU(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"simple", "cpu=64,mem=1216G,gres/gpu=3", 3},
		{"typed", "cpu=64,mem=1216G,gres/gpu=3,gres/gpu:l40s=3", 3},
		{"no gpu", "cpu=64,mem=1216G", 0},
		{"complex", "cpu=128,mem=1547658M,gres/gpu=4,gres/gpu:l40s=4,gres/tmpsize=7696581394432", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTRESGPU(tt.in); got != tt.want {
				t.Errorf("ParseTRESGPU(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestNodeAllocGPUs(t *testing.T) {
	tests := []struct {
		name string
		node NodeInfo
		want int
	}{
		{
			name: "from alloc_gres",
			node: NodeInfo{AllocGres: "gpu:b200:8(IDX:0-7)"},
			want: 8,
		},
		{
			name: "from alloc_tres fallback",
			node: NodeInfo{AllocGres: "", AllocTRES: "cpu=64,mem=1216G,gres/gpu=3,gres/gpu:l40s=3"},
			want: 3,
		},
		{
			name: "alloc_gres preferred",
			node: NodeInfo{AllocGres: "gpu:2", AllocTRES: "gres/gpu=5"},
			want: 2,
		},
		{
			name: "no gpu",
			node: NodeInfo{},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeAllocGPUs(tt.node); got != tt.want {
				t.Errorf("NodeAllocGPUs() = %d, want %d", got, tt.want)
			}
		})
	}
}
