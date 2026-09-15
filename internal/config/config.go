// Package config loads srest's configuration from the environment.
//
// Precedence is as follows:
//  1. Environment variables (SLURM_URL, SLURM_JWT, SLURM_USER_NAME,
//     SLURM_API_VERSION).
//  2. Default values (URL only).
package config

import (
	"os"
	"os/user"
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

	// AuthToken is a bearer token sent in the Authorization header. When set,
	// it replaces the default X-SLURM-USER-TOKEN header (for proxy setups).
	AuthToken string

	// CustomHeaders is a comma-separated list of "Key:Value" pairs added to
	// every HTTP request (e.g. "X-Project:myproj,X-User:svc-123").
	CustomHeaders string
}

// Load reads the configuration from environment variables and applies default
// values when nothing is defined.
func Load() Config {
	return Config{
		URL:           getEnv("SLURM_URL", defaultURL),
		JWT:           os.Getenv("SLURM_JWT"),
		Username:      getEnv("SLURM_USER_NAME", currentUser()),
		APIVersion:    os.Getenv("SLURM_API_VERSION"),
		Insecure:      os.Getenv("SLURM_INSECURE") == "true",
		AuthToken:     os.Getenv("SLURM_AUTH_TOKEN"),
		CustomHeaders: os.Getenv("SLURM_CUSTOM_HEADERS"),
	}
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
