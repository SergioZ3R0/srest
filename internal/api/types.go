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

// StringList unmarshals a JSON field that may be a string or an array of
// strings (some Slurm fields vary between versions).
type StringList []string

// UnmarshalJSON accepts a string, an array of strings or null.
func (s *StringList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = nil
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		*s = arr
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = []string{str}
		return nil
	}
	return nil
}

// JobInfo is a summarized Slurm job as returned by GET /slurm/vX/jobs.
type JobInfo struct {
	JobID     uint32     `json:"job_id"`
	Name      string     `json:"name"`
	User      string     `json:"user_name"`
	State     StringList `json:"job_state"`
	RunTime   int64      `json:"run_time"`
	Nodes     string     `json:"nodes"`
	Partition string     `json:"partition"`
}

// NoVal models Slurm's optional numeric struct {set, infinite, number}.
type NoVal struct {
	Set      bool  `json:"set"`
	Infinite bool  `json:"infinite"`
	Number   int64 `json:"number"`
}

// ExitCode describes a job's exit status.
type ExitCode struct {
	Status     StringList `json:"status"`
	ReturnCode NoVal      `json:"return_code"`
	Signal     struct {
		Name string `json:"name"`
	} `json:"signal"`
}

// JobDetail is a full job record from GET /slurm/vX/job/{job_id}.
type JobDetail struct {
	JobID          uint32     `json:"job_id"`
	Name           string     `json:"name"`
	User           string     `json:"user_name"`
	Account        string     `json:"account"`
	Partition      string     `json:"partition"`
	State          StringList `json:"job_state"`
	TimeLimit      NoVal      `json:"time_limit"` // minutes
	RunTime        int64      `json:"run_time"`   // seconds
	StartTime      NoVal      `json:"start_time"` // unix
	EndTime        NoVal      `json:"end_time"`
	Nodes          string     `json:"nodes"`
	NodeCount      NoVal      `json:"node_count"`
	CPUs           NoVal      `json:"cpus"`
	StandardOutput string     `json:"standard_output"`
	StandardError  string     `json:"standard_error"`
	ExitCode       ExitCode   `json:"exit_code"`
}

// NodeInfo is a Slurm node as returned by GET /slurm/vX/nodes.
type NodeInfo struct {
	Name        string     `json:"name"`
	State       StringList `json:"state"`
	CPUs        int        `json:"cpus"`
	RealMemory  int64      `json:"real_memory"` // MB
	AllocMemory int64      `json:"alloc_memory"`
	Partitions  StringList `json:"partitions"`
}

// PartitionInfo is a Slurm partition as returned by GET /slurm/vX/partitions.
type PartitionInfo struct {
	Name  string `json:"name"`
	Nodes struct {
		Configured string `json:"configured"`
		Total      int    `json:"total"`
	} `json:"nodes"`
}

// AccountInfo is a Slurm account from GET /slurmdb/vX/accounts.
type AccountInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Organization string `json:"organization"`
}
