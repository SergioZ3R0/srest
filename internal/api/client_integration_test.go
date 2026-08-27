package api

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestPingIntegration validates version auto-detection and ping against a real
// slurmrestd. It is skipped automatically when SLURM_URL/SLURM_JWT are unset.
//
// Example:
//
//	SLURM_URL=http://localhost:6820 SLURM_JWT=<token> go test ./internal/api -run TestPingIntegration -v
func TestPingIntegration(t *testing.T) {
	url := os.Getenv("SLURM_URL")
	jwt := os.Getenv("SLURM_JWT")
	if url == "" || jwt == "" {
		t.Skip("SLURM_URL and SLURM_JWT not set; skipping integration test")
	}

	c := New(url, jwt, "slurm")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	v, err := c.Detect(ctx)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	t.Logf("detected version: %s", v)

	info, err := c.Ping(ctx)
	if err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Logf("Slurm release: %s", info.Slurm.Release)

	if info.API != v {
		t.Errorf("inconsistent API version: %s != %s", info.API, v)
	}
	if info.Slurm.Release == "" {
		t.Error("empty Slurm release")
	}
}
