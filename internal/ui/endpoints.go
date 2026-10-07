package ui

// category groups endpoints by API family.
type category struct {
	name      string
	endpoints []endpoint
}

// paramKind distinguishes between query parameters (go in URL query string)
// and path parameters (embedded in the URL path).
type paramKind int

const (
	paramQuery paramKind = iota
	paramPath
)

// param is a single editable query parameter for an endpoint.
type param struct {
	name    string
	value   string
	options []string
	kind    paramKind
}

// endpoint describes a Slurm REST API endpoint and its supported parameters.
type endpoint struct {
	name   string
	method string
	base   string // "slurm", "slurmdb", or "action"
	path   string // may contain {placeholders} for path params
	params []param
}

// getParams returns only query parameters (for URL construction).
func (e endpoint) getParams() []param {
	var out []param
	for _, p := range e.params {
		if p.kind == paramQuery {
			out = append(out, p)
		}
	}
	return out
}

// pathParams returns only path parameters (for URL substitution).
func (e endpoint) pathParams() []param {
	var out []param
	for _, p := range e.params {
		if p.kind == paramPath {
			out = append(out, p)
		}
	}
	return out
}

// categories defines all Slurm REST API endpoints exposed by the composer,
// organized by API family. Each category is selectable in the sidebar.
var categories = []category{
	{
		name: "slurm",
		endpoints: []endpoint{
			{
				name:   "ping",
				method: "GET",
				base:   "slurm",
				path:   "/ping",
			},
			{
				name:   "jobs",
				method: "GET",
				base:   "slurm",
				path:   "/jobs",
				params: []param{
					{name: "users"},
					{name: "account"},
					{name: "partition"},
					{name: "qos"},
					{name: "state"},
					{name: "node"},
					{name: "name"},
				},
			},
			{
				name:   "job detail",
				method: "GET",
				base:   "slurm",
				path:   "/job/{job_id}",
				params: []param{
					{name: "job_id", kind: paramPath},
				},
			},
			{
				name:   "nodes",
				method: "GET",
				base:   "slurm",
				path:   "/nodes",
				params: []param{
					{name: "nodes"},
					{name: "state"},
					{name: "partition"},
					{name: "name"},
				},
			},
			{
				name:   "node detail",
				method: "GET",
				base:   "slurm",
				path:   "/node/{node_name}",
				params: []param{
					{name: "node_name", kind: paramPath},
				},
			},
			{
				name:   "partitions",
				method: "GET",
				base:   "slurm",
				path:   "/partitions",
				params: []param{
					{name: "partition"},
					{name: "nodes"},
					{name: "name"},
				},
			},
			{
				name:   "reservations",
				method: "GET",
				base:   "slurm",
				path:   "/reservations",
				params: []param{
					{name: "users"},
					{name: "name"},
				},
			},
			{
				name:   "shares",
				method: "GET",
				base:   "slurm",
				path:   "/shares",
				params: []param{
					{name: "users"},
					{name: "accounts"},
				},
			},
		},
	},
	{
		name: "slurmdb",
		endpoints: []endpoint{
			{
				name:   "associations",
				method: "GET",
				base:   "slurmdb",
				path:   "/associations",
				params: []param{
					{name: "account"},
					{name: "cluster"},
					{name: "user"},
					{name: "partition"},
					{name: "qos"},
					{name: "tres"},
				},
			},
			{
				name:   "association detail",
				method: "GET",
				base:   "slurmdb",
				path:   "/association",
				params: []param{
					{name: "account"},
					{name: "cluster"},
					{name: "user"},
					{name: "partition"},
				},
			},
			{
				name:   "accounts",
				method: "GET",
				base:   "slurmdb",
				path:   "/accounts",
				params: []param{
					{name: "account"},
					{name: "parent_account"},
					{name: "cluster"},
				},
			},
			{
				name:   "account detail",
				method: "GET",
				base:   "slurmdb",
				path:   "/account/{account_name}",
				params: []param{
					{name: "account_name", kind: paramPath},
				},
			},
			{
				name:   "clusters",
				method: "GET",
				base:   "slurmdb",
				path:   "/clusters",
				params: []param{
					{name: "cluster"},
				},
			},
			{
				name:   "cluster detail",
				method: "GET",
				base:   "slurmdb",
				path:   "/cluster/{cluster_name}",
				params: []param{
					{name: "cluster_name", kind: paramPath},
				},
			},
			{
				name:   "users",
				method: "GET",
				base:   "slurmdb",
				path:   "/users",
				params: []param{
					{name: "name"},
					{name: "account"},
					{name: "cluster"},
				},
			},
			{
				name:   "user detail",
				method: "GET",
				base:   "slurmdb",
				path:   "/user/{name}",
				params: []param{
					{name: "name", kind: paramPath},
				},
			},
			{
				name:   "qos",
				method: "GET",
				base:   "slurmdb",
				path:   "/qos",
				params: []param{
					{name: "qos"},
				},
			},
			{
				name:   "tres",
				method: "GET",
				base:   "slurmdb",
				path:   "/tres",
				params: []param{
					{name: "tres"},
				},
			},
			{
				name:   "wckeys",
				method: "GET",
				base:   "slurmdb",
				path:   "/wckeys",
				params: []param{
					{name: "id_users"},
				},
			},
			{
				name:   "jobs",
				method: "GET",
				base:   "slurmdb",
				path:   "/jobs",
				params: []param{
					{name: "users"},
					{name: "account"},
					{name: "partition"},
					{name: "qos"},
					{name: "state"},
					{name: "cluster"},
					{name: "job_id"},
					{name: "name"},
				},
			},
			{
				name:   "job detail",
				method: "GET",
				base:   "slurmdb",
				path:   "/job/{job_id}",
				params: []param{
					{name: "job_id", kind: paramPath},
				},
			},
			{
				name:   "diag",
				method: "GET",
				base:   "slurmdb",
				path:   "/diag",
			},
			{
				name:   "config",
				method: "GET",
				base:   "slurmdb",
				path:   "/config",
			},
		},
	},
	{
		name: "actions",
		endpoints: []endpoint{
			{
				name:   "submit job",
				method: "POST",
				base:   "slurm",
				path:   "/job/submit",
				params: []param{
					{name: "name"},
					{name: "partition"},
					{name: "qos"},
					{name: "account"},
					{name: "gres"},
					{name: "time_limit", value: "60"},
					{name: "nodes", value: "1"},
					{name: "cpus_per_task", value: "1"},
					{name: "memory_per_node"},
					{name: "standard_output"},
					{name: "standard_error"},
					{name: "current_working_directory", value: "/tmp"},
					{name: "script_path"},
					{name: "script", value: "#!/bin/bash\nhostname"},
				},
			},
			{
				name:   "job action",
				method: "POST",
				base:   "slurm",
				path:   "/job/{job_id}",
				params: []param{
					{name: "job_id", kind: paramPath},
					{name: "action", options: []string{"CANCEL", "REQUEUE", "HOLD", "RELEASE"}},
				},
			},
			{
				name:   "node state",
				method: "POST",
				base:   "slurm",
				path:   "/node/{node_name}",
				params: []param{
					{name: "node_name", kind: paramPath},
					{name: "state", options: []string{"drain", "undrain", "down", "resume", "power_down", "power_up"}},
					{name: "reason"},
				},
			},
			{
				name:   "cancel job",
				method: "DELETE",
				base:   "slurm",
				path:   "/job/{job_id}",
				params: []param{
					{name: "job_id", kind: paramPath},
				},
			},
			{
				name:   "create association",
				method: "POST",
				base:   "slurmdb",
				path:   "/associations",
				params: []param{
					{name: "cluster", value: "linux"},
					{name: "account"},
					{name: "user"},
					{name: "partition"},
				},
			},
			{
				name:   "create account",
				method: "POST",
				base:   "slurmdb",
				path:   "/accounts",
				params: []param{
					{name: "account"},
					{name: "description"},
					{name: "organization"},
				},
			},
		},
	},
}
