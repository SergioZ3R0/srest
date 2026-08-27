// Package config carga la configuración de srest desde el entorno.
//
// La prioridad es la siguiente:
//  1. Variables de entorno (SLURM_URL, SLURM_JWT, SLURM_USER_NAME).
//  2. Valores por defecto (solo para la URL).
package config

import (
	"os"
	"os/user"
)

// defaultURL es el endpoint por defecto de slurmrestd.
const defaultURL = "http://localhost:6820"

// Config agrupa los parámetros necesarios para hablar con la REST API de Slurm.
type Config struct {
	// URL es la dirección base de slurmrestd (por ejemplo http://localhost:6820).
	URL string

	// JWT es el token de autenticación que se envía en la cabecera
	// X-SLURM-USER-TOKEN.
	JWT string

	// Username es el usuario que se envía en la cabecera X-SLURM-USER-NAME.
	Username string
}

// Load lee la configuración desde las variables de entorno y aplica los
// valores por defecto cuando no hay nada definido.
func Load() Config {
	return Config{
		URL:      getEnv("SLURM_URL", defaultURL),
		JWT:      os.Getenv("SLURM_JWT"),
		Username: getEnv("SLURM_USER_NAME", currentUser()),
	}
}

// getEnv devuelve el valor de key o, si está vacío, el valor fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// currentUser devuelve el nombre del usuario actual del sistema, usado como
// valor por defecto para X-SLURM-USER-NAME.
func currentUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}
