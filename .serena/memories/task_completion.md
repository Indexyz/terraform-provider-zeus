# Task Completion Checklist
- Run formatting: `make fmt` (gofmt) before committing.
- Run lint: `make lint` to satisfy configured golangci linters.
- Run tests: `make test`; include `make testacc` when changes affect real resources and TF_ACC context is available.
- Regenerate docs/assets if schemas/resources change: `make generate`.
- Build/install as needed: `make build` or `make install` to ensure provider compiles.
- Verify git status and diffs; add brief summaries in PR/commits.