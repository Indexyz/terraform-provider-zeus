# Port Resource Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Terraform CRUD and lookup support for Zeus port assignments using Zeus resource IDs as the canonical Terraform identity.

**Architecture:** Extend the Zeus API client with ID-based port operations plus optional create-time `X-Portd-Host` header support, then add a `zeus_port` resource and `zeus_port` data source that mirror the existing `pool` and `assign` patterns. Model the header input as `scope_host` and the read-back Zeus value as computed `host` so omitted/default/import flows remain stable. All configurable attributes are replacement-only because the documented Zeus API does not expose an update endpoint.

**Tech Stack:** Go 1.24, terraform-plugin-framework, terraform-plugin-testing, tfplugindocs, Zeus HTTP JSON API.

---

### Task 1: Extend Zeus API client for port operations

**Files:**
- Modify: `internal/zeusapi/client.go`
- Test: `internal/zeusapi/client_test.go` (optional if you choose to unit-test header helper behavior separately)

- [ ] **Step 1: Add port request/response types in the client**

Define types for the documented port API:

```go
type CreatePortRequest struct {
	AssignID   string `json:"assignId"`
	TargetPort int64  `json:"targetPort"`
	Service    string `json:"service"`
	Host       string `json:"-"`
}

type CreatePortResponse struct {
	ID   string `json:"id"`
	Port int64  `json:"port"`
}

type PortInfo struct {
	ID         string `json:"id"`
	AssignID   string `json:"assignId"`
	Host       string `json:"host"`
	Port       int64  `json:"port"`
	TargetPort int64  `json:"targetPort"`
	Service    string `json:"service"`
	CreatedAt  string `json:"createdAt"`
}
```

`Host` is request metadata only and must not be serialized into the JSON body.

- [ ] **Step 2: Add a helper path for custom request headers**

Refactor the HTTP layer minimally so create requests can attach `X-Portd-Host` without duplicating the entire request flow.

Before wiring methods, verify the endpoint and header names against `docs/zeus.md` so the code and docs stay aligned.

One acceptable shape:

```go
func (c *Client) doWithHeaders(ctx context.Context, method, path string, payload any, headers map[string]string, out any) error
```

Then keep `do()` as a thin wrapper that passes `nil` headers.

- [ ] **Step 3: Implement port client methods**

Add:

```go
func (c *Client) CreatePort(ctx context.Context, req CreatePortRequest) (CreatePortResponse, error)
func (c *Client) GetPortByID(ctx context.Context, id string) (PortInfo, error)
func (c *Client) DeletePortByID(ctx context.Context, id string) error
```

Behavior:
- `CreatePort` uses `POST /port`
- includes `X-Portd-Host` only when `req.Host != ""`
- read/delete use `/port/id/:id`
- treat empty host exactly like unset by omitting the header instead of sending an empty value

- [ ] **Step 4: Run focused tests after compile-time changes**

Run:

```bash
go test ./internal/provider/... 
```

Expected: existing provider tests still pass or, if port files are not yet added, compile remains green.

- [ ] **Step 5: Commit**

```bash
git add internal/zeusapi/client.go
git commit -m "feat: add zeus port api client support"
```

### Task 2: Implement the `zeus_port` resource

**Files:**
- Create: `internal/provider/port_resource.go`
- Modify: `internal/provider/provider.go`
- Test: `internal/provider/port_resource_test.go`

- [ ] **Step 1: Write the failing acceptance test for resource create/read/import**

Add a new test server flow that covers:
- `POST /port`
- `GET /port/id/port-1`
- `DELETE /port/id/port-1`

Expected checks:

```go
resource.TestCheckResourceAttr("zeus_port.test", "assign_id", "assign-1")
resource.TestCheckResourceAttr("zeus_port.test", "host", "node-1")
resource.TestCheckNoResourceAttr("zeus_port.test", "scope_host") // in the config variant that omits it
resource.TestCheckResourceAttr("zeus_port.test", "port", "32022")
resource.TestCheckResourceAttr("zeus_port.test", "target_port", "22")
resource.TestCheckResourceAttr("zeus_port.test", "service", "ssh")
```

- [ ] **Step 2: Run the new test to verify it fails**

Run:

```bash
go test ./internal/provider -run TestAccPortResourceAndDataSource -v
```

Expected: FAIL with a compile or provider-registration error because `zeus_port` is not registered or implemented yet.

- [ ] **Step 3: Implement the resource schema and lifecycle**

Follow the `pool` / `assign` pattern:

```go
type portModel struct {
	ID         types.String `tfsdk:"id"`
	AssignID   types.String `tfsdk:"assign_id"`
	ScopeHost  types.String `tfsdk:"scope_host"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	TargetPort types.Int64  `tfsdk:"target_port"`
	Service    types.String `tfsdk:"service"`
	CreatedAt  types.String `tfsdk:"created_at"`
}
```

Schema rules:
- `assign_id`, `target_port`, `service` required + `RequiresReplace`
- `scope_host` optional + `RequiresReplace`
- `host`, `id`, `port`, `created_at` computed

Important modeling rule:
- `scope_host` is configuration-only input for the create header
- `host` is read-only observed state from Zeus
- do not try to backfill `scope_host` from `GET /port/id/:id`
- imported resources must remain stable when config omits `scope_host`

Lifecycle rules:
- `Create`: call `CreatePort`, save `id`, refresh state
- `Read`: call `GetPortByID`, remove state on 404
- `Delete`: call `DeletePortByID`, ignore 404
- `ImportState`: passthrough `id`

- [ ] **Step 4: Register the resource in provider metadata**

Update `provider.go` resource list:

```go
return []func() resource.Resource{
	NewPoolResource,
	NewAssignResource,
	NewPortResource,
}
```

- [ ] **Step 5: Run the resource test again**

Run:

```bash
go test ./internal/provider -run TestAccPortResourceAndDataSource -v
```

Expected: PASS for the resource path, or fail only because the data source task is not implemented yet.

- [ ] **Step 6: Commit**

```bash
git add internal/provider/port_resource.go internal/provider/provider.go internal/provider/port_resource_test.go
git commit -m "feat: add zeus port resource"
```

### Task 3: Implement the `zeus_port` data source

**Files:**
- Create: `internal/provider/port_datasource.go`
- Modify: `internal/provider/provider.go`
- Modify: `internal/provider/port_resource_test.go`

- [ ] **Step 1: Add the failing data source lookup in the test config**

Use the same style as existing resources:

```hcl
data "zeus_port" "by_id" {
  id = zeus_port.test.id
}
```

Expected checks:

```go
resource.TestCheckResourceAttr("data.zeus_port.by_id", "port", "32022")
resource.TestCheckResourceAttr("data.zeus_port.by_id", "service", "ssh")
```

- [ ] **Step 2: Run the focused test to verify it fails**

Run:

```bash
go test ./internal/provider -run TestAccPortResourceAndDataSource -v
```

Expected: FAIL because `zeus_port` data source is not registered or implemented yet.

- [ ] **Step 3: Implement the data source**

Match the existing ID-lookup data source pattern:
- required `id`
- computed `assign_id`, `host`, `port`, `target_port`, `service`, `created_at`
- use `GetPortByID`
- keep `scope_host` out of the data source because it is not a read API field

- [ ] **Step 4: Register the data source in `provider.go`**

Update the list:

```go
return []func() datasource.DataSource{
	NewPoolDataSource,
	NewAssignDataSource,
	NewPortDataSource,
}
```

- [ ] **Step 5: Run the focused resource/data source test again**

Run:

```bash
go test ./internal/provider -run TestAccPortResourceAndDataSource -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/provider/port_datasource.go internal/provider/provider.go internal/provider/port_resource_test.go
git commit -m "feat: add zeus port data source"
```

### Task 4: Cover header and not-found edge cases

**Files:**
- Modify: `internal/provider/port_resource_test.go`

- [ ] **Step 1: Add a test for optional scope host header behavior**

Create a test server assertion that verifies `X-Portd-Host` is omitted when Terraform config does not set `scope_host`.

- [ ] **Step 2: Add a test for import stability with omitted `scope_host`**

Create/import a resource whose read response includes `host = "Default"` (or another concrete host value) while Terraform config omits `scope_host`, then verify the next plan is empty.

- [ ] **Step 3: Add a test for read 404 removing state**

Pattern target: same style as existing resource acceptance tests, but force the GET endpoint to return 404 after create/import.

- [ ] **Step 4: Add a test for delete 404 tolerance**

Force `DELETE /port/id/:id` to return 404 and verify destroy still succeeds.

- [ ] **Step 5: Add a test for data source 404 diagnostics**

Look up a non-existent port ID through `data "zeus_port"` and verify Terraform returns a diagnostic instead of silently succeeding.

- [ ] **Step 6: Run the port-focused test suite**

Run:

```bash
go test ./internal/provider -run Port -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/provider/port_resource_test.go
git commit -m "test: cover zeus port edge cases"
```

### Task 5: Update docs and generated artifacts

**Files:**
- Modify: `README.md`
- Create: `examples/resources/zeus_port/resource.tf`
- Create: `examples/resources/zeus_port/import.sh`
- Create: `examples/data-sources/zeus_port/data-source.tf`
- Create: `docs/resources/port.md`
- Create: `docs/data-sources/port.md`
- Modify: any generated docs/index output affected by `make generate`

- [ ] **Step 1: Update README supported resources and data sources**

Add:
- `zeus_port` to supported resources
- `zeus_port` (lookup by `id`) to supported data sources

- [ ] **Step 2: Create tfplugindocs example files**

Create:
- `examples/resources/zeus_port/resource.tf`
- `examples/resources/zeus_port/import.sh`
- `examples/data-sources/zeus_port/data-source.tf`

Ensure examples show:
- `zeus_assign` feeding `assign_id`
- `scope_host` as the create-time optional input
- import by ID

- [ ] **Step 3: Generate docs**

Run:

```bash
make generate
```

Expected: `docs/resources/port.md` and `docs/data-sources/port.md` are generated or updated.

- [ ] **Step 4: Verify generated docs mention import-by-id**

Confirm resource docs include an import example like:

```shell
terraform import zeus_port.example "port-id"
```

- [ ] **Step 5: Commit**

```bash
git add README.md docs/resources/port.md docs/data-sources/port.md docs/
git commit -m "docs: add zeus port resource documentation"
```

### Task 6: Final verification

**Files:**
- Modify: none expected unless verification uncovers issues

- [ ] **Step 1: Run diagnostics/build-quality checks**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run documentation generation one final time**

Run:

```bash
make generate
git diff --exit-code
```

Expected: generation is stable and leaves no unexpected diffs.

- [ ] **Step 3: Inspect changed files before handoff**

Review:
- `internal/zeusapi/client.go`
- `internal/provider/port_resource.go`
- `internal/provider/port_datasource.go`
- `internal/provider/port_resource_test.go`
- `internal/provider/provider.go`
- `examples/resources/zeus_port/resource.tf`
- `examples/resources/zeus_port/import.sh`
- `examples/data-sources/zeus_port/data-source.tf`
- generated docs and `README.md`

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "chore: finish zeus port resource support"
```
