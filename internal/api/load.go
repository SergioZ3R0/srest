package api

import (
	"sort"
	"strconv"
	"strings"
)

// ComputePartitionLoads aggregates per-partition resource usage from a list of
// nodes. Partitions with no active nodes are still listed (at zero usage).
func ComputePartitionLoads(nodes []NodeInfo) []PartitionLoad {
	type accum struct {
		name       string
		totalNodes int
		totalCPUs  int
		usedCPUs   int
		totalMem   int64
		usedMem    int64
		totalGPUs  int
		usedGPUs   int
	}

	index := map[string]*accum{}

	for _, n := range nodes {
		for _, pname := range n.Partitions {
			a, ok := index[pname]
			if !ok {
				a = &accum{name: pname}
				index[pname] = a
			}
			a.totalNodes++
			a.totalCPUs += n.CPUs
			a.usedCPUs += n.AllocCPUs
			a.totalMem += n.RealMemory
			a.usedMem += n.AllocMemory
			a.totalGPUs += ParseGRESCount(n.Gres)
			a.usedGPUs += NodeAllocGPUs(n)
		}
	}

	out := make([]PartitionLoad, 0, len(index))
	for _, a := range index {
		out = append(out, PartitionLoad{
			Name:       a.name,
			TotalNodes: a.totalNodes,
			TotalCPUs:  a.totalCPUs,
			UsedCPUs:   a.usedCPUs,
			TotalMemMB: a.totalMem,
			UsedMemMB:  a.usedMem,
			TotalGPUs:  a.totalGPUs,
			UsedGPUs:   a.usedGPUs,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

// ParseGRESCount extracts the GPU count from a GRES string.
// Supported formats:
//
//	""                       → 0
//	"gpu:4"                  → 4
//	"gpu:a100:4"             → 4
//	"gpu:4,gpu:2"            → 6  (summed)
//	"gpu:b200:8(S:0-1)"      → 8  (suffix stripped)
//	"gpu:b200:8(IDX:0-7)"    → 8  (suffix stripped)
func ParseGRESCount(s string) int {
	if s == "" {
		return 0
	}

	total := 0
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "gpu") {
			continue
		}
		// Strip suffix like "(S:0-1)" or "(IDX:0-7)".
		if idx := strings.Index(part, "("); idx >= 0 {
			part = part[:idx]
		}
		fields := strings.Split(part, ":")
		// The last field is always the count.
		if len(fields) < 2 {
			continue
		}
		n, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			continue
		}
		total += n
	}
	return total
}

// ParseTRESGPU extracts the GPU count from a TRES string.
// TRES format: "cpu=64,mem=1216G,gres/gpu=3,gres/gpu:l40s=3,gres/tmpsize=..."
// Returns the gres/gpu count (ignores typed variants like gres/gpu:l40s).
func ParseTRESGPU(s string) int {
	if s == "" {
		return 0
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		// Match "gres/gpu=N" but not "gres/gpu:type=N".
		if strings.HasPrefix(part, "gres/gpu=") {
			val := strings.TrimPrefix(part, "gres/gpu=")
			n, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			return n
		}
	}
	return 0
}

// NodeAllocGPUs returns the allocated GPU count for a node.
// Tries alloc_gres, then gres_used, then tres_used.
func NodeAllocGPUs(n NodeInfo) int {
	if count := ParseGRESCount(n.AllocGres); count > 0 {
		return count
	}
	if count := ParseGRESCount(n.GresUsed); count > 0 {
		return count
	}
	return ParseTRESGPU(n.TresUsed)
}
