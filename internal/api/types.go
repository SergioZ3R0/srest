package api

import "encoding/json"

// Response is the common wrapper of every slurmrestd response. Every response
// includes metadata and may include warnings and query errors.
type Response struct {
	Meta     Meta      `json:"meta"`
	Errors   []Error   `json:"errors"`
	Warnings []Warning `json:"warnings"`
}

// Meta groups the metadata slurmrestd includes in every response.
type Meta struct {
	Plugin Plugin       `json:"plugin"`
	Slurm  SlurmVersion `json:"-"`
}

// UnmarshalJSON tolerates the field-name change across versions: older
// versions use "Slurm" (capitalized) while newer ones (v0.0.44+) use "slurm".
func (m *Meta) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if b, ok := raw["plugin"]; ok {
		if err := json.Unmarshal(b, &m.Plugin); err != nil {
			return err
		}
	}

	var slurm []byte
	if b, ok := raw["slurm"]; ok {
		slurm = b
	} else if b, ok := raw["Slurm"]; ok {
		slurm = b
	}
	if slurm != nil {
		if err := json.Unmarshal(slurm, &m.Slurm); err != nil {
			return err
		}
	}
	return nil
}

// Plugin describes the content plugin that generated the response.
type Plugin struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// SlurmVersion holds the Slurm version serving the API. Only release is mapped
// because the remaining fields (major/minor/micro) change type between
// versions (integers in older ones, strings in newer ones).
type SlurmVersion struct {
	Release string `json:"release"`
}

// Error represents an error returned by slurmrestd inside the response.
type Error struct {
	Description string `json:"description"`
	ErrorNumber int    `json:"error_number"`
	Error       string `json:"error"`
	Source      string `json:"source"`
}

// Warning represents a warning returned by slurmrestd. It is common to find
// them when a request includes fields the API version does not recognize
// (they are ignored and reported here).
type Warning struct {
	Description string `json:"description"`
	Source      string `json:"source"`
}

// PingInfo is the result of a successful /ping call.
type PingInfo struct {
	// API is the data_parser version we talked to.
	API Version

	// Slurm is the Slurm version serving the API.
	Slurm SlurmVersion

	// Warnings are the warnings reported by slurmrestd.
	Warnings []Warning
}
