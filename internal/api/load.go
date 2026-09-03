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
			a.totalGPUs += parseGRESCount(n.Gres)
			a.usedGPUs += parseGRESCount(n.AllocGres)
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

// parseGRESCount extracts the GPU count from a GRES string.
// Supported formats:
//
//	""            → 0
//	"gpu:4"       → 4
//	"gpu:a100:4"  → 4
//	"gpu:4,gpu:2" → 6  (summed)
func parseGRESCount(s string) int {
	if s == "" {
		return 0
	}

	total := 0
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "gpu") {
			continue
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
