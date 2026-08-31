## Description
Please include a summary of the change and the motivation. If it fixes an open bug, please link to the issue.

Fixes # (issue number)

## Type of change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (e.g., new tab, new API method, new query builder capability)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Please describe the tests that you ran to verify your changes.
- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `golangci-lint run ./...`
- [ ] Integration test against a live cluster (`SLURM_URL=... SLURM_JWT=... go test ./internal/api -run TestPingIntegration -v`) if applicable

## Checklist:
- [ ] My code follows the existing style of this project (gofmt, idiomatic Go)
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in complex API or version-gating interactions
- [ ] I have updated the `README.md` / docs if new commands, endpoints or configurations were added