// Package history provides persistent storage for the request history log.
// Entries are kept in a JSON file under ~/.local/share/srest/history.json
// and automatically pruned to the most recent 100 when the limit is reached.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is a single recorded HTTP request/response pair.
type Entry struct {
	Method     string        `json:"method"`
	URL        string        `json:"url"`
	StatusCode int           `json:"status"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
	Warnings   []string      `json:"warnings,omitempty"`
	Timestamp  time.Time     `json:"ts"`
}

// MaxEntries is the maximum number of entries kept in the history file.
const MaxEntries = 100

// Store manages the persistent request history on disk.
type Store struct {
	path string
}

// New creates a Store that reads from and writes to the default path
// (~/.local/share/srest/history.json). The file is created if it does
// not exist.
func New() (*Store, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}

	dir = filepath.Join(dir, ".local", "share", "srest")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating history directory: %w", err)
	}

	s := &Store{path: filepath.Join(dir, "history.json")}

	// Create the file if it doesn't exist.
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		if err := os.WriteFile(s.path, []byte("[]"), 0o644); err != nil {
			return nil, fmt.Errorf("creating history file: %w", err)
		}
	}

	return s, nil
}

// Load reads the history from disk.
func (s *Store) Load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("reading history: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decoding history: %w", err)
	}
	return entries, nil
}

// Save writes the history to disk, keeping only the most recent MaxEntries.
func (s *Store) Save(entries []Entry) error {
	if len(entries) > MaxEntries {
		entries = entries[len(entries)-MaxEntries:]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding history: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("writing history: %w", err)
	}
	return nil
}

// Append adds an entry to the history, prunes to MaxEntries, and persists.
func (s *Store) Append(entry Entry) ([]Entry, error) {
	entries, err := s.Load()
	if err != nil {
		return nil, err
	}

	entries = append(entries, entry)
	if len(entries) > MaxEntries {
		entries = entries[len(entries)-MaxEntries:]
	}

	if err := s.Save(entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// FromQuery converts an api.Query to a history.Entry.
func FromQuery(method, url string, status int, duration time.Duration, err error, warnings []string) Entry {
	e := Entry{
		Method:     method,
		URL:        url,
		StatusCode: status,
		Duration:   duration,
		Timestamp:  time.Now(),
	}
	if err != nil {
		e.Error = err.Error()
	}
	e.Warnings = append(e.Warnings, warnings...)
	return e
}
