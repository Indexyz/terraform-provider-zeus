# Style and Conventions
- Language: Go; enforce gofmt on all Go sources.
- Linting via golangci-lint with enabled linters: copyloopvar, durationcheck, errcheck, forcetypeassert, godot, ineffassign, makezero, misspell, nilerr, predeclared, staticcheck, unconvert, unparam, unused, usetesting. Excludes generated/comments/common false positives; skips third_party, builtin, examples.
- Provider scaffolding patterns: provider type implements Metadata/Schema/Configure; resources implement CRUD with types from terraform-plugin-framework; use tflog for logging; models use `types.*` values and plan modifiers as needed.
- Documentation generated via tfplugindocs; terraform fmt applied to examples during generation.
- Licensing headers via hashicorp/copywrite; keep MPL-2.0 headers intact.
- Naming: current provider type `scaffolding`, resource type `${provider}_example`; update TypeName/Address/provider-name when customizing.