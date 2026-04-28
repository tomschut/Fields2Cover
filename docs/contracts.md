# Contracts: OpenAPI ↔ proto (intentional disjunction)

**Status:** canonical — do not modify without updating both specs and getting
CODEOWNERS review.

Fields2Cover v2.0 has **two** source-of-truth contract files:

| File | Scope | Consumers |
|---|---|---|
| [`api-go/openapi.yaml`](../api-go/openapi.yaml) | External HTTP surface | wodan, farmmaps, any future v2 client |
| [`proto/f2c.proto`](../proto/f2c.proto) | Internal Go ↔ f2c C++ gRPC | `api-go/` (Go client), `f2c-grpc/` (C++ server) |

**These two files are disjoint on purpose.** There is no cross-generator, no
shared schemas, no auto-derived types across the HTTP↔gRPC boundary. The Go
handler layer (`api-go/internal/convert/`) explicitly converts between them.

## Why two contracts?

1. **Blast radius isolation.** A change to the f2c C++ data model (e.g.,
   adding a field to `PathState`) should be free to land in `proto/f2c.proto`
   without auto-breaking the public HTTP API. The handler layer decides
   whether the new field surfaces externally and on what schedule.

2. **Different serialization tradeoffs per boundary.**
   - HTTP clients want inline GeoJSON (human-debuggable, jq-friendly, browser
     devtools-friendly).
   - The C++ shim wants WKT-as-`bytes` (compact, round-trips exactly through
     GDAL's geometry library, avoids repeated-double message bloat on 10k-vertex
     fields).
   - Sharing a single schema would force one side to tolerate the other's
     format. Disjoint schemas let each side use the right wire format.

3. **Cross-FFI codegen is a known footgun.** Auto-generating OpenAPI schemas
   from proto (or vice versa) couples two rapidly-evolving specs across a
   language boundary. The v2.0 language research called this out explicitly
   as something to avoid. Explicit hand-written conversion is less clever,
   but dramatically more debuggable.

4. **The conversion layer is small.** Each journey step maps to roughly
   20 lines of Go in `api-go/internal/convert/` — one file per step. These
   are pure functions with no dependencies on live gRPC or HTTP state, so
   they're trivially unit-testable.

## The conversion rule

Every Go handler in `api-go/internal/handlers/` MUST follow this shape:

```go
func (h *Handler) GenerateSwaths(w http.ResponseWriter, r *http.Request, id string) {
    // 1. Decode the OpenAPI request body (generated from openapi.yaml).
    var req api.GenerateSwathsRequest
    if err := render.Decode(r, &req); err != nil { /* 400 */ }

    // 2. Convert OpenAPI → proto (hand-written, in convert/).
    protoReq, err := convert.ToGenerateSwathsRequest(&req)
    if err != nil { /* 400 */ }

    // 3. Call the shim over gRPC (client generated from f2c.proto).
    protoResp, err := h.f2c.GenerateSwaths(ctx, protoReq)
    if err != nil { /* map grpc Status → HTTP status */ }

    // 4. Convert proto → OpenAPI (hand-written, in convert/).
    apiResp, err := convert.FromGenerateSwathsResponse(protoResp)
    if err != nil { /* 500 */ }

    // 5. Attach the journey envelope and write the response.
    apiResp.Journey = journey.NextAfter("swaths:generate", id)
    render.JSON(w, r, apiResp)
}
```

No handler may import proto types into an OpenAPI response or OpenAPI types
into a proto request. The `convert/` package is the only place they touch.

## Rules for future changes

- **Adding a field to openapi.yaml** does NOT require a proto change unless
  the Go handler needs new data from the shim to populate it.
- **Adding a field to f2c.proto** does NOT require an openapi.yaml change. If
  the new field should surface externally, update openapi.yaml in a separate
  PR after the proto change lands and the shim can emit it.
- **Renaming a field** on one side does NOT propagate to the other side —
  only `convert/` changes.
- **Any PR that adds a cross-generator** (e.g., a `gen-openapi-from-proto`
  make target) will be rejected. Point the PR author at this document.

## CODEOWNERS

Both files are listed in [`CODEOWNERS`](../CODEOWNERS) so every change gets a
review from someone who has read this document.
