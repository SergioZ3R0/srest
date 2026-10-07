package ui

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SergioZ3R0/srest/internal/api"
)

// exportJobs writes the jobs data to a CSV file and returns the path.
func exportJobs(data []api.JobInfo) (string, error) {
	path := exportPath("jobs")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"job_id", "name", "user", "state", "partition", "run_time", "nodes"}); err != nil {
		return "", err
	}
	for _, j := range data {
		if err := w.Write([]string{
			fmt.Sprintf("%d", j.JobID),
			j.Name,
			j.User,
			strings.Join(j.State, ","),
			j.Partition,
			fmt.Sprintf("%d", j.RunTime),
			j.Nodes,
		}); err != nil {
			return "", err
		}
	}
	return path, nil
}

// exportNodes writes the nodes data to a CSV file and returns the path.
func exportNodes(data []api.NodeInfo) (string, error) {
	path := exportPath("nodes")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"name", "state", "cpus", "alloc_cpus", "memory_mb", "alloc_memory_mb", "partitions", "gres", "alloc_gres"}); err != nil {
		return "", err
	}
	for _, n := range data {
		if err := w.Write([]string{
			n.Name,
			strings.Join(n.State, ","),
			fmt.Sprintf("%d", n.CPUs),
			fmt.Sprintf("%d", n.AllocCPUs),
			fmt.Sprintf("%d", n.RealMemory),
			fmt.Sprintf("%d", n.AllocMemory),
			strings.Join(n.Partitions, ","),
			n.Gres,
			n.AllocGres,
		}); err != nil {
			return "", err
		}
	}
	return path, nil
}

// exportPartitions writes the partitions data to a CSV file and returns the path.
func exportPartitions(data []api.PartitionInfo) (string, error) {
	path := exportPath("partitions")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"name", "configured_nodes", "total_nodes", "max_wall_minutes"}); err != nil {
		return "", err
	}
	for _, p := range data {
		maxWall := ""
		if !p.Maximums.Time.Infinite && p.Maximums.Time.Number > 0 {
			maxWall = fmt.Sprintf("%d", p.Maximums.Time.Number)
		}
		if err := w.Write([]string{
			p.Name,
			p.Nodes.Configured,
			fmt.Sprintf("%d", p.Nodes.Total),
			maxWall,
		}); err != nil {
			return "", err
		}
	}
	return path, nil
}

// exportPath returns a filename like "srest-jobs-20060102-150405.csv".
func exportPath(kind string) string {
	ts := time.Now().Format("20060102-150405")
	return filepath.Join(".", fmt.Sprintf("srest-%s-%s.csv", kind, ts))
}

// exportResponse writes the query response body to a JSON file.
func exportResponse(body string) (string, error) {
	ts := time.Now().Format("20060102-150405")
	path := filepath.Join(".", fmt.Sprintf("srest-response-%s.json", ts))

	// Try to pretty-print as JSON; fall back to raw.
	var pretty []byte
	if err := json.Unmarshal([]byte(body), &pretty); err == nil {
		// Already valid JSON - re-marshal pretty.
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(body), "", "  "); err == nil {
			pretty = buf.Bytes()
		} else {
			pretty = []byte(body)
		}
	} else {
		pretty = []byte(body)
	}

	if err := os.WriteFile(path, pretty, 0644); err != nil {
		return "", err
	}
	return path, nil
}
