// Package config loads srest's configuration from the environment.
//
// Precedence is as follows:
//  1. Environment variables (SLURM_URL, SLURM_JWT, SLURM_USER_NAME,
//     SLURM_API_VERSION).
//  2. ~/.srest/config.vault (encrypted, prompts for password).
//  3. Default values (URL only).
package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// defaultURL is the default slurmrestd endpoint.
const defaultURL = "http://localhost:6820"

// Config groups the parameters needed to talk to the Slurm REST API.
type Config struct {
	// URL is the base address of slurmrestd (e.g. http://localhost:6820).
	URL string

	// JWT is the authentication token sent in the X-SLURM-USER-TOKEN header.
	JWT string

	// Username is the user sent in the X-SLURM-USER-NAME header.
	Username string

	// APIVersion is the API version to use (optional, e.g. "v0.0.44"). When
	// empty, srest auto-detects the version supported by the cluster.
	APIVersion string

	// Insecure skips TLS certificate verification when true.
	Insecure bool

	// CACert is the path to a PEM-encoded CA certificate file for custom
	// or self-signed TLS endpoints. When empty, the system trust store is used.
	CACert string

	// AuthToken is a bearer token sent in the Authorization header. When set,
	// it replaces the default X-SLURM-USER-TOKEN header (for proxy setups).
	AuthToken string

	// CustomHeaders is a comma-separated list of "Key:Value" pairs added to
	// every HTTP request (e.g. "X-Project:myproj,X-User:svc-123").
	CustomHeaders string
}

// Load reads the configuration from environment variables, the encrypted vault,
// and applies default values when nothing is defined.
func Load() Config {
	cfg := Config{}

	// --- 1. Environment variables (highest priority) ---
	cfg.URL = getEnv("SLURM_URL", defaultURL)
	cfg.JWT = os.Getenv("SLURM_JWT")
	cfg.Username = getEnv("SLURM_USER_NAME", currentUser())
	cfg.APIVersion = os.Getenv("SLURM_API_VERSION")
	cfg.Insecure = os.Getenv("SLURM_INSECURE") == "true"
	cfg.CACert = os.Getenv("SLURM_CA_CERT")
	cfg.AuthToken = os.Getenv("SLURM_AUTH_TOKEN")
	cfg.CustomHeaders = os.Getenv("SLURM_CUSTOM_HEADERS")

	// --- 2. Vault file (if JWT not provided via env) ---
	if cfg.JWT == "" && cfg.AuthToken == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			vaultPath := filepath.Join(home, ".srest", "config.vault")
			if data, readErr := os.ReadFile(vaultPath); readErr == nil && IsVaultFile(data) {
				pass := VaultPassword()
				if pass == "" {
					pass = promptVaultPassword()
				}
				plain, decErr := Decrypt(data, pass)
				if decErr != nil {
					fmt.Fprintf(os.Stderr, "srest: vault decrypt: %v\n", decErr)
				} else {
					cfg = applyVaultConfig(string(plain), cfg)
				}
			}
		}
	}

	// --- 3. Defaults ---
	if cfg.URL == "" {
		cfg.URL = defaultURL
	}
	if cfg.Username == "" {
		cfg.Username = currentUser()
	}

	return cfg
}

// applyVaultConfig parses KEY=VALUE lines from the vault plaintext and applies
// them to the config (only for fields not already set by env vars).
func applyVaultConfig(data string, cfg Config) Config {
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "SLURM_URL":
			if cfg.URL == defaultURL {
				cfg.URL = value
			}
		case "SLURM_JWT":
			if cfg.JWT == "" {
				cfg.JWT = value
			}
		case "SLURM_USER_NAME":
			if cfg.Username == "" || cfg.Username == currentUser() {
				cfg.Username = value
			}
		}
	}
	return cfg
}

// promptVaultPassword prompts the user for the vault password (hidden input).
func promptVaultPassword() string {
	fmt.Fprint(os.Stderr, "Vault password: ")

	// Hide input on Linux/macOS
	cmd := exec.Command("stty", "-echo")
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err == nil {
		defer func() {
			cmd = exec.Command("stty", "echo")
			cmd.Stdin = os.Stdin
			_ = cmd.Run()
		}()
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

// ParseCustomHeaders parses the SLURM_CUSTOM_HEADERS env var into a map.
// Format: "Key1:Value1,Key2:Value2".
func (c Config) ParseCustomHeaders() map[string]string {
	if c.CustomHeaders == "" {
		return nil
	}
	headers := map[string]string{}
	for _, pair := range strings.Split(c.CustomHeaders, ",") {
		pair = strings.TrimSpace(pair)
		if idx := strings.IndexByte(pair, ':'); idx > 0 {
			key := strings.TrimSpace(pair[:idx])
			val := strings.TrimSpace(pair[idx+1:])
			if key != "" {
				headers[key] = val
			}
		}
	}
	return headers
}

// getEnv returns the value of key or, when empty, the fallback value.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// currentUser returns the username of the current OS user, used as the default
// for X-SLURM-USER-NAME.
func currentUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}
