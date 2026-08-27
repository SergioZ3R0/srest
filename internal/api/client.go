// Package api proporciona un cliente HTTP puro para la REST API de Slurm.
//
// Este paquete no debe contener ninguna lógica de interfaz: únicamente se
// encarga de construir y enviar peticiones HTTP hacia slurmrestd.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIVersion es el prefijo de versión de la API que consume srest.
const APIVersion = "v0.0.40"

// Client es un cliente HTTP minimalista para interactuar con slurmrestd.
type Client struct {
	baseURL    string
	jwt        string
	username   string
	httpClient *http.Client
}

// New crea un nuevo cliente de la API con la configuración indicada.
func New(baseURL, jwt, username string) *Client {
	return &Client{
		baseURL:  baseURL,
		jwt:      jwt,
		username: username,
		httpClient: &http.Client{
			// Timeout global para evitar que una petición se quede colgada.
			Timeout: 10 * time.Second,
		},
	}
}

// Ping comprueba la conectividad con slurmrestd llamando al endpoint /ping.
//
// Devuelve nil si la API responde con un estado 200 y el cuerpo es un JSON
// de ping válido; en caso contrario devuelve un error descriptivo.
func (c *Client) Ping(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/slurm/%s/ping", c.baseURL, APIVersion)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("construyendo la petición: %w", err)
	}

	// Cabeceras de autenticación requeridas por slurmrestd.
	req.Header.Set("X-SLURM-USER-TOKEN", c.jwt)
	req.Header.Set("X-SLURM-USER-NAME", c.username)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("contactando %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("leyendo la respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("estado inesperado %s: %s", resp.Status, string(body))
	}

	// Respuesta típica de /ping:
	// {"meta": {"plugin": {...}, "Slurm": {"version": {...}}}}
	var payload struct {
		Meta struct {
			Slurm map[string]any `json:"Slurm"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decodificando la respuesta: %w", err)
	}

	return nil
}
