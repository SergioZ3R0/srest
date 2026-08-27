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
}

// Load reads the configuration from environment variables and applies default
// values when nothing is defined.
func Load() Config {
	return Config{
		URL:        getEnv("SLURM_URL", defaultURL),
		JWT:        os.Getenv("SLURM_JWT"),
		Username:   getEnv("SLURM_USER_NAME", currentUser()),
		APIVersion: os.Getenv("SLURM_API_VERSION"),
	}
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
