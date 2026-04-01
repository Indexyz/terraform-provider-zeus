# Zeus Port Resource Design

## Goal

Add a Terraform-managed `zeus_port` resource that supports create, read, delete, and import flows for Zeus port assignments, using the Zeus resource `id` as the Terraform resource identity.

## Context

The current provider exposes `zeus_pool` and `zeus_assign`. Both resources follow the same lifecycle shape:

- create via API
- persist server-returned `id`
- refresh state via read-by-id API
- remove state on 404 during read
- ignore 404 during delete
- import by `id`

The Zeus port API in `docs/zeus.md` fits this model if the provider treats the returned port-assignment `id` as the canonical resource identity.

## Recommended Approach

Model `zeus_port` by Zeus `id`.

### Why this is the right fit

1. It matches existing provider conventions for `pool` and `assign`.
2. It allows simple `terraform import zeus_port.example <id>` behavior.
3. It avoids coupling Terraform state to `host + port`, where `port` is allocated by the server and not known before create.
4. It lets read and delete use the ID-based APIs that do not require `X-Portd-Host`.

## Alternatives Considered

### Alternative A: Model by `host + port`

Rejected. This makes resource identity depend on a server-assigned field and would complicate import, drift recovery, and replacement semantics.

### Alternative B: Support both `id` identity and `host + port` import formats

Deferred. This adds parsing and documentation complexity without helping the main CRUD path. It can be added later if operators have a real import need.

## Resource Shape

Resource type: `zeus_port`

### Input attributes

- `assign_id` (String, required, ForceNew/RequiresReplace)
  - Maps to API field `assignId`
  - Must reference an existing Zeus assign record
- `target_port` (Number, required, ForceNew/RequiresReplace)
  - Maps to API field `targetPort`
- `service` (String, required, ForceNew/RequiresReplace)
  - Maps to API field `service`
- `scope_host` (String, optional, ForceNew/RequiresReplace)
  - Maps to request header `X-Portd-Host`
  - This is a create-time scope selector, not the canonical resource identity
  - When unset or `null`, the provider omits the header and Zeus uses its default scope host
  - Empty string should be normalized the same as unset for request construction, so the provider omits the header rather than sending an empty header value

### Computed attributes

- `id` (String)
  - Zeus port resource ID returned by create API
- `host` (String)
  - Canonical scope host returned by `GET /port/id/:id`
- `port` (Number)
  - Allocated external/public port returned by Zeus
- `created_at` (String)
  - From `GET /port/id/:id`

### Observed attributes copied from read API

- `assign_id` (String)
- `host` (String)
- `target_port` (Number)
- `service` (String)

The resource schema should expose these values even if Zeus normalizes or defaults them during creation.

The critical modeling decision is to separate:

- `scope_host`: optional create-time input that decides whether `X-Portd-Host` is sent, and
- `host`: computed read-time output that reflects the actual Zeus scope host for the port resource.

This avoids Terraform state instability in omitted/default/import scenarios. If the configuration omits `scope_host` and Zeus fills in a concrete default host, the provider should accept the read value in computed `host` as canonical state. That is expected normalization rather than drift.

## Lifecycle Mapping

### Create

- Call `POST /port`
- If `scope_host` is set to a non-empty value, include `X-Portd-Host: <scope_host>`
- Request body:
  - `assignId`
  - `targetPort`
  - `service`
- Save returned `id`
- Immediately refresh via `GET /port/id/:id` to populate canonical state, including the final observed `host`, `port`, and `created_at`

### Read

- Call `GET /port/id/:id`
- On 404: remove resource from Terraform state
- Otherwise, map response fields into state

### Update

- No Zeus update API exists in `docs/zeus.md`
- All configurable attributes should therefore be replacement-only
- Terraform should therefore plan replacement when any configurable attribute changes rather than attempting an in-place update
- If the framework still requires an `Update` method implementation, it should be a minimal defensive pass-through and should never call a Zeus update endpoint

Changing `scope_host` is therefore a replacement because it changes create-time allocation scope. Changes to computed `host` alone are drift observations from Zeus, not user configuration changes.

### Delete

- Call `DELETE /port/id/:id`
- On 404: treat as already deleted
- Do not use `DELETE /port/:port`, because Terraform identity is the resource `id`

### Import

- Support import by Zeus resource `id`
- Use passthrough import state on `id`
- Imported resources should be stable even when configuration omits `scope_host`; the provider should read and retain the computed `host` from Zeus without requiring the user to backfill `scope_host`

## API Client Additions

Add Zeus client types and methods for port operations:

- request/response structs for create and read
- a helper that can attach the optional `X-Portd-Host` header
- `CreatePort(ctx, req)`
- `GetPortByID(ctx, id)`
- `DeletePortByID(ctx, id)`

The current generic `do()` helper does not support per-request custom headers, so the design should either:

1. extend `do()` with a headers argument, or
2. add a dedicated helper for requests that need custom headers.

Recommendation: extend the HTTP helper in a minimal way so future header-scoped APIs can reuse it.

## Provider Registration

The provider should register:

- a new resource constructor: `NewPortResource`
- a new data source constructor: `NewPortDataSource`

This keeps parity with existing `pool` and `assign` support.

## Data Source Shape

Data source type: `zeus_port`

Lookup by `id`, mirroring the resource import identity.

### Required attribute

- `id` (String)

### Computed attributes

- `assign_id`
- `host`
- `port`
- `target_port`
- `service`
- `created_at`

The data source should use the same mapping logic as resource refresh to avoid drift between documentation and behavior.

## Error Handling

- Any create/read/delete API failure should surface as a Terraform diagnostic
- Read 404 should remove resource state
- Delete 404 should be ignored
- Validation beyond Terraform types should be left to Zeus unless a clear provider-side rule is necessary
- If `assign_id` does not reference an existing assign, the provider should rely on the Zeus API error response and surface it directly as a create diagnostic rather than attempting a separate preflight lookup
- External changes to computed fields such as `port` are ordinary drift: the next read should update state to match Zeus, not force replacement by themselves
- The provider should not try to synthesize `scope_host` from the read response during import or refresh; `scope_host` is configuration-only and may remain unset while `host` is populated from Zeus

## Testing Strategy

Add acceptance-style tests using the existing `newTestServer` helper.

### Resource test coverage

- create `zeus_port`
- verify request body fields were sent correctly
- verify `X-Portd-Host` header is sent when `scope_host` is configured
- verify read populates computed fields
- verify data source lookup by `id`
- verify import by `id`

### Additional test coverage

- create without `scope_host` and verify the header is omitted
- read 404 removes resource from state
- delete 404 is tolerated
- data source lookup of a non-existent `id` returns a diagnostic error
- import remains no-diff when the resource was created in default scope and configuration omits `scope_host`

If test count needs to stay small initially, the first acceptance test should cover the happy path plus import, and targeted unit/acceptance tests can cover header omission and not-found behavior.

## Documentation Updates

Update the following documentation surfaces:

- `README.md` supported resources and data sources
- generated resource doc for `zeus_port`
- generated data source doc for `zeus_port`

Example documentation should demonstrate the dependency on `zeus_assign`:

```hcl
resource "zeus_assign" "vm" {
  region = ["us-east-1"]
  host   = "node-1"
  key    = "vm-123"
  type   = "vm"
}

resource "zeus_port" "ssh" {
  assign_id   = zeus_assign.vm.id
  scope_host  = "node-1"
  target_port = 22
  service     = "ssh"
}
```

## Open Questions

None currently blocking the first implementation. The documented API is sufficient for CRUD + import when modeled by `id`.

## Non-Goals

- Supporting in-place update of port assignments
- Supporting deletion by `host + port`
- Supporting alternative import formats in the first version
