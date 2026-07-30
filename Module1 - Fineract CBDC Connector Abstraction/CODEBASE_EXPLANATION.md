# Module 1 — Fineract CBDC Connector Abstraction: Codebase Explanation

> A walkthrough of what this module is, how the code is organized, what each file
> does, and the caveats you should know before treating it as buildable code.

---

## 1. What this module is

This is the **connector abstraction layer** for the "E1 Cross-Border CBDC Payment
Platform" — a Go module that sits between Apache Fineract (the core banking
system) and the various national CBDC networks (India's e₹, UAE's Digital Dirham,
GIFT City, etc.).

The problem it solves: every central bank ships a different CBDC API. Some are
blockchain-based, some are centralized ledgers, all have different auth, error
formats, and settlement semantics. Without an abstraction, Fineract's payment
orchestrator would need bespoke code paths per country.

This module defines **one interface** — `CBDCConnector` — plus the shared data
models, error taxonomy, config schema, and observability plumbing that every
network-specific connector must conform to. The orchestrator then talks to
`CBDCConnector` and never knows which country it's settling in.

It is a **library, not a service**. There is no `main()`, no HTTP server, no
database. It ships a contract and the supporting types.

---

## 2. Repository layout (and an important caveat)

```
Module1 - Fineract CBDC Connector Abstraction/
├── Model Structure-Module1 ....docx          # design doc (binary)
├── Module1- Complete Production-Ready Code....zip
└── Module1- Complete Production-Ready Code -Fineract CBDC Connector Abstraction/
    ├── 1.go.mod.go                            # module definition + deps
    ├── 2.connector.go.go                      # THE interface + all request/response types
    ├── 3.models.go.go                         # metadata, capabilities, events
    ├── 4.errors.go.go                         # ConnectorError + error code taxonomy
    ├── 5.types-types.go.go                    # Amount, enums, shared primitives
    ├── 6.config- config.go.go                 # ConnectorConfig and sub-configs
    ├── 7.context-context.go.go                # request/trace ID context propagation
    ├── 8.validation-validation.go.go          # struct-tag + hand-rolled validators
    ├── 9.metrics- metrics.go.go               # Prometheus metrics
    ├── 10.tracing- tracing.go.go              # OpenTelemetry span helpers
    ├── 11.test-mock-mock_connector.go.go      # in-memory MockConnector (639 lines)
    ├── 12.example-sexample_usage.go.go        # runnable-style usage examples
    ├── 13- README.md.md                       # README (part 1)
    ├── 14.Usage/                              # doc snippets, 3 files
    ├── 15. Error Handling.go … 19.Dependencies.txt  # more doc snippets
```

### ⚠️ This is a numbered document dump, not a Go module

The files are **prefixed with sequence numbers and given `.go.go` extensions**.
That means:

- `go build` will not work. Go requires files in directories matching their
  `package` declaration — `package types` must live in a `types/` directory, not
  in a flat folder as `5.types-types.go.go`.
- Files 13–19 are **prose/snippets**, not compilable Go. `15. Error Handling.go`
  is a 7-line README excerpt with a `.go` extension.
- The filenames encode the *intended* layout. Reconstructed, the real module
  should be:

```
connector-abstraction/
├── go.mod
├── connector.go          # package connector
├── models.go             # package connector
├── errors.go             # package connector
├── types/types.go        # package types
├── config/config.go      # package config
├── context/context.go    # package context
├── validation/validation.go
├── metrics/metrics.go
├── tracing/tracing.go
├── test/mock/mock_connector.go
├── examples/example_usage.go
└── README.md
```

Treat the current folder as a **design artifact / code deliverable for review**,
not a checked-out working tree.

---

## 3. Architecture

```
┌─────────────────────────────────────────────────┐
│         Fineract Core / Orchestrator            │
└──────────────────────┬──────────────────────────┘
                       │ depends only on the interface
                       ▼
┌─────────────────────────────────────────────────┐
│           CBDCConnector  (interface)            │
│  Issue · Transfer · Lock · Burn · Redeem        │
│  GetBalance · GetTransaction · Status queries   │
│  HealthCheck · GetNetworkInfo                   │
│  Initialize · Shutdown · Status                 │
└───┬───────────┬────────────┬─────────────┬──────┘
    │           │            │             │
    ▼           ▼            ▼             ▼
 India     UAE Digital   GIFT City     MockConnector
 e₹ conn.    Dirham       conn.        (test double)
    │           │            │             │
    ▼           ▼            ▼             ▼
        each national CBDC network API

  Cross-cutting, shared by all implementations:
  types · config · validation · errors · metrics · tracing · context
```

The design is textbook **ports-and-adapters (hexagonal)**. `CBDCConnector` is the
port; each national connector is an adapter. Nothing in this module imports a
national API — the dependency arrow always points inward.

---

## 4. File-by-file

### `2.connector.go.go` — the contract (442 lines, package `connector`)

The single most important file. Defines:

**`CBDCConnector` interface**, grouped into four concerns:

| Group | Methods | Purpose |
|---|---|---|
| **Value operations** | `Issue`, `Transfer`, `Lock`, `Burn`, `Redeem` | mutate token supply / ownership |
| **Queries** | `GetBalance`, `GetTransaction`, `GetTransactionStatus` | read-only |
| **Health** | `HealthCheck`, `GetNetworkInfo` | for monitoring & readiness probes |
| **Lifecycle** | `Initialize`, `Shutdown`, `Status` | managed startup/teardown |

Every method takes `context.Context` first — so cancellation, deadlines, and
trace IDs propagate to the network call. This is correct and non-negotiable for
a cross-border payment system where a hung RPC must not hold a settlement lock.

**The five value operations map to CBDC lifecycle stages:**

- `Issue` — central bank mints tokens into a wallet (fiat → CBDC).
- `Transfer` — move tokens wallet-to-wallet. The everyday payment.
- `Lock` — *the cross-border primitive.* Reserves tokens for a bounded duration
  so a two-leg settlement (e.g. INR leg + AED leg) can commit atomically. Returns
  a `LockID` and an `ExpiresAt`. This is what enables HTLC/PvP-style settlement.
- `Burn` — permanently destroy tokens.
- `Redeem` — CBDC → fiat, crediting a named `FiatAccount`.

**`ConnectorStatus`** — a lifecycle state machine:
`uninitialized → initializing → ready`, with `degraded`, `unavailable`, `error`
as failure states. `degraded` is separate from `unavailable` deliberately: the
orchestrator can keep routing to a degraded connector at reduced volume rather
than failing over entirely.

**Request/response structs.** Each operation gets a dedicated pair. Consistent
patterns across all of them:

- `ReferenceID` on every request — the external idempotency/reconciliation key.
- `Metadata map[string]interface{}` — escape hatch for network-specific fields
  (purpose codes, regulatory tags) without polluting the shared contract.
- Responses carry `TransactionID`, `BlockHash`, `Timestamp`, `Status`, `Fee`,
  `ConfirmationCount` — the union of what blockchain and centralized ledgers
  return, with blockchain-only fields marked `omitempty`.
- `validate:"required,gt=0"` struct tags drive the validation layer.

### `3.models.go.go` — capability negotiation & events

- **`ConnectorMetadata`** — name, version, supported currencies, network type.
- **`ConnectorCapabilities`** — the feature-discovery mechanism. Booleans like
  `SupportsLock`, `SupportsAtomicTransactions`, plus limits (`MaxBatchSize`,
  `MinConfirmation`, `MaxTransactionAmount`, `TransactionTimeout`).

  This matters because the interface is uniform but the networks aren't. A
  connector for a network without locking still has to *implement* `Lock` — it
  returns `ErrFeatureNotSupported`. Capabilities let the orchestrator check
  *before* attempting, and route a cross-border payment to a corridor that can
  actually settle atomically.

- **`Event` / `EventHandler` / `EventSubscription`** — an async notification
  surface for `transaction_confirmed`, `lock_expired`, `balance_changed`, etc.
  Note: this is **declared but not wired into `CBDCConnector`** — the interface
  has no `Subscribe` method. It's scaffolding for a future revision.

### `4.errors.go.go` — the error taxonomy

`ConnectorError` implements `error` and `Unwrap()`, so `errors.Is`/`errors.As`
work through the chain.

Two tiers of `ErrorCode`:
- **Generic transport/protocol**: `internal_error`, `timeout`, `network_error`,
  `validation_error`, `unauthorized`, `rate_limit_exceeded`, …
- **CBDC-domain**: `insufficient_balance`, `invalid_wallet`, `lock_expired`,
  `lock_conflict`, `compliance_check_failed`, `duplicate_transaction`,
  `blockchain_unreachable`, `invalid_signature`, `feature_not_supported`, …

`HTTPStatusCode()` maps every code to an HTTP status, so a REST gateway above
this layer gets correct status codes for free. Worth noting the deliberate
choices: `insufficient_balance` and `limit_exceeded` → **402 Payment Required**;
`compliance_check_failed` → **403 Forbidden**; lock conflicts and duplicate
transactions → **409 Conflict**.

Helpers: `NewError`, `NewErrorWithDetails`, `IsConnectorError`,
`AsConnectorError`.

### `5.types-types.go.go` — shared primitives (package `types`)

**`Amount` is the critical type here.** Money is represented as
`{Value *big.Int, Decimal int}` — an arbitrary-precision integer of minor units
plus a scale. `1000.00 INR` is stored as `Value=100000, Decimal=2`.

This is the right call. Floating-point money loses cents; `big.Int` scaled
integers don't. Methods: `NewAmount(string, decimal)`, `String()`, `Add`, `Sub`,
`Cmp`, `IsZero`, `IsNegative`, and JSON marshalling that serializes to a
**string** (`"1000.00"`) rather than a JSON number — avoiding precision loss in
JavaScript consumers.

`Add`/`Sub` return an error on decimal-scale mismatch rather than silently
coercing. Good.

Also here: `TransactionStatus` (pending → processing → confirmed / failed /
expired / rolled_back / cancelled), `TransactionType`, `HealthStatus`,
`ComponentHealth`, `EventType`, `FeatureType`, `NetworkType`.

### `6.config- config.go.go` — configuration schema (package `config`)

`ConnectorConfig` composed of eight sub-structs, all with both `json` and `yaml`
tags so config can be file- or API-driven:

- `NetworkConfig` — endpoint, timeout, connection pool sizing, `Extra` map
- `AuthConfig` — supports `api_key`, `oauth2`, `basic`; carries client
  credentials, scopes, refresh token
- `TimeoutConfig` — separate connect/read/write/request/**transaction**/idle
  timeouts. Splitting transaction timeout from request timeout matters: a
  blockchain confirmation can take minutes while the RPC call takes milliseconds.
- `RetryConfig` — exponential backoff with `MaxAttempts`, `InitialDelay`,
  `MaxDelay`, `Multiplier`, and an allowlist of `RetryableErrors`
- `RateLimitConfig` — token bucket (`RequestsPerSecond` + `BurstSize`)
- `TLSConfig` — mTLS-capable (cert/key/CA files, SNI server name)
- `LogConfig` — includes `SensitiveDataMasking`

`Validate()` checks the four mandatory fields (ID, Name, Type, endpoint).
Note that the retry/rate-limit/TLS configs are **declared but not enforced by
this module** — implementing backoff and throttling is left to each connector.

`17.Configuration.yaml` is a worked example of this schema.

### `7.context-context.go.go` — correlation IDs (package `context`)

Typed context keys (`contextKey string`, avoiding collisions) for `request_id`,
`connector_id`, `trace_id`, `span_id`, `user_id`, `client_ip`, with
`With*`/`Get*` accessors.

`NewRequestContext(parent)` auto-generates UUIDv4 request and trace IDs if
absent — so every operation is traceable end-to-end across the orchestrator and
all connectors.

### `8.validation-validation.go.go` — input validation

Two layers:

1. **Declarative** — wraps `go-playground/validator/v10`, driven by the
   `validate:"..."` struct tags on the request types. Registers three custom
   validators: `currency`, `amount`, `transaction_type`.
   `formatValidationError` turns validator errors into readable messages.
2. **Imperative** — `ValidateTransferRequest`, `ValidateLockRequest` for rules
   the tags can't express, e.g. *source wallet must not equal destination
   wallet*, *lock duration must be positive*.

### `9.metrics- metrics.go.go` — Prometheus (package `metrics`)

Five metric families, all labelled by `connector` and `operation`:

| Metric | Type | Labels |
|---|---|---|
| `connector_request_duration_seconds` | Histogram | connector, operation, status |
| `connector_request_total` | Counter | connector, operation, status |
| `connector_request_in_flight` | Gauge | connector, operation |
| `connector_status` | Gauge | connector, name |
| `connector_balance` | Gauge | connector, wallet_id, currency |

`RequestMetrics` is a small helper: construct at operation start, call `Record()`
on completion to emit duration, increment the counter, and decrement in-flight.
`SetConnectorStatus` encodes the lifecycle states as 0–5.

⚠️ **`connector_balance` is labelled by `wallet_id`.** In production that's an
unbounded-cardinality label — one time series per wallet will melt a Prometheus
server. It's also arguably PII in a metrics store. This should be removed or
aggregated before real deployment.

### `10.tracing- tracing.go.go` — OpenTelemetry (package `tracing`)

Wraps OTel spans in a `Span` type with fluent, domain-aware helpers:
`WithTransaction(txID, txType, status)`, `WithWallet(walletID)`,
`WithAmount(amount, currency)`. `AddAttribute` does type-switched conversion to
`attribute.KeyValue`. All methods nil-guard the embedded span, so tracing is safe
when uninitialized.

### `11.test-mock-mock_connector.go.go` — the test double (639 lines)

A fully in-memory `MockConnector` implementing all 13 interface methods, with
mutex-guarded state (`balances`, `locks`, `txs`).

The interesting part is the **fault-injection API**, which is what makes this
genuinely useful for testing a payment system:

- `SetSimulateError(code)` / `ClearSimulateError()` — force any error code
- `SetSimulateLatency(d)` — inject delay, to exercise timeout handling
- `SetFailOnOperation("transfer")` — fail one specific operation
- `SetAllowedOperation(op, bool)` — simulate `feature_not_supported`
- `SetBalance(wallet, currency, amount)` — arrange test state
- `GetMockTransactions()` — assert on what happened

`simulateOperation` respects `ctx.Done()` during the injected latency, so
context-cancellation tests actually work.

### `12.example-sexample_usage.go.go` — usage examples

Five example functions: basic usage, error handling with `AsConnectorError` +
`HTTPStatusCode()`, timeout handling via `context.WithTimeout` against injected
latency, health checking, and a batch-operations placeholder.

### `13`–`19` — documentation

README, usage snippets, error-handling and validation recipes, the YAML config
example, and the dependency list. Files 14–19 are prose extracted into separate
files; `13- README.md.md` and `19.Dependencies.txt` are the two halves of the
README.

---

## 5. Dependencies (`1.go.mod.go`)

Go 1.21. Seven direct dependencies, all mainstream:

| Package | Role |
|---|---|
| `google/uuid` | request/trace/transaction ID generation |
| `pkg/errors` | error wrapping *(largely redundant with stdlib `errors` in Go 1.21)* |
| `prometheus/client_golang` | metrics |
| `go.opentelemetry.io/otel` + `/trace` | distributed tracing |
| `stretchr/testify` | test assertions |
| `go-playground/validator/v10` | struct-tag validation |
| `go.uber.org/zap` | structured logging *(declared but unused in the code shown)* |

No blockchain SDKs, no HTTP clients, no database drivers — correct for an
abstraction layer.

---

## 6. Design decisions worth calling out

**Good:**

1. **`big.Int`-based `Amount`.** The single most important correctness decision
   in a payments codebase, and it's right.
2. **`context.Context` on every method.** Cancellation and tracing propagate.
3. **Capability negotiation.** Uniform interface without pretending all networks
   are equivalent.
4. **`Lock` as a first-class primitive.** Without it, cross-border atomic
   settlement is impossible; bolting it on later would break every connector.
5. **Two-tier error taxonomy with HTTP mapping.** Consistent error handling
   across the whole platform, gateway included.
6. **Separate transaction vs. request timeouts.** Reflects how settlement
   actually behaves.
7. **`Metadata` escape hatches** on requests, so network-specific fields don't
   force interface churn.

**Gaps and risks:**

1. **The event system is dead code.** `EventHandler` and `EventSubscription`
   exist but `CBDCConnector` has no `Subscribe`/`Unsubscribe`. Async settlement
   confirmation currently has no path other than polling
   `GetTransactionStatus`.
2. **No `Unlock` method.** `types.TypeUnlock` is defined as a transaction type
   and `EventLockReleased` exists, but the interface can only *create* locks —
   there's no way to release one early. Locks can only expire.
3. **Retry / rate limit / TLS config is declarative only.** No middleware in this
   module enforces it, so each connector re-implements backoff — exactly the
   duplication an abstraction layer should prevent. A `RetryingConnector`
   decorator wrapping `CBDCConnector` would fix this.
4. **No idempotency contract.** `ReferenceID` is present on every request but
   nothing in the interface docs states whether replaying the same `ReferenceID`
   is safe. For a payments system this must be specified, not implied.
5. **No batch operations.** `SupportsBatchOperations` and `MaxBatchSize` are in
   `ConnectorCapabilities`, but the interface has no batch methods.
6. **Metrics cardinality** — the `wallet_id` label, noted above.
7. **Secrets in config structs.** `APIKey`, `Secret`, `ClientSecret`, and
   `RefreshToken` are plain `string` fields with `json`/`yaml` tags, meaning they
   will serialize into any config dump or log line. They should be a redacting
   type with a `String()`/`MarshalJSON()` that emits `***`.

**Compilation issues in the current snapshot** (beyond the file-layout problem):

- `3.models.go.go` uses `context.Context` in the `EventHandler` interface but
  never imports `context`.
- `4.errors.go.go` calls `errors.As` in `AsConnectorError` but imports only
  `fmt` and `net/http`.
- `5.types-types.go.go` uses `strings.TrimSpace`, `strings.Split`,
  `strings.Repeat`, `strings.TrimRight` without importing `strings`.
- `8.validation-validation.go.go` references `TransferRequest`/`LockRequest`
  unqualified from `package validation`; those live in `package connector`, and
  importing `connector` there would create an import cycle if `connector` ever
  imports `validation`.
- `10.tracing- tracing.go.go`: `Span.AddEvent` calls `s.AddEvent(...)` — infinite
  recursion, since the method shadows the embedded `trace.Span` method. Also
  `trace.StatusError` isn't a valid identifier in OTel v1.21 (should be
  `codes.Error` from `go.opentelemetry.io/otel/codes`).
- `12.example-sexample_usage.go.go` shadows the imported package name
  `connector` with a local variable `connector := mock.NewMockConnector(cfg)`,
  then uses `connector.BalanceRequest` — which resolves to the variable, not the
  package.

None of these are design flaws; they're artifacts of code written as a document
rather than compiled. But they mean **"production-ready" in the folder name is
aspirational** — the module needs a restructuring pass and a `go build` before it
compiles.

---

## 7. How you'd actually use this

**Implementing a new connector** (per `14.Usage/1`):

```go
type MyCBDCConnector struct { /* network client, creds, etc. */ }

func (c *MyCBDCConnector) Issue(ctx context.Context, req *connector.IssueRequest) (*connector.IssueResponse, error) {
    if err := validation.Validate(req); err != nil {
        return nil, connector.NewError(connector.ErrValidation, err.Error(), err)
    }
    // ... call the national CBDC API, translate its errors to ErrorCodes ...
    return &connector.IssueResponse{TransactionID: id, Status: types.StatusConfirmed}, nil
}
// ... plus the other 12 methods
```

**Consuming it** — the orchestrator holds a `connector.CBDCConnector` and never
imports a concrete connector package. Wiring happens once at startup from
`ConnectorConfig`.

**Testing** — swap in `mock.NewMockConnector(cfg)`, arrange balances with
`SetBalance`, inject failures with `SetSimulateError` / `SetSimulateLatency`, and
assert against `GetMockTransactions()`.

---

## 8. Where this fits

This is **Module 1** of a larger platform (`CBDC_MG`). It is the foundation
layer: it defines the contract that Modules 2+ (the national connector
implementations, the settlement orchestrator, the compliance engine) build
against. Getting the interface right here is disproportionately important —
every downstream module inherits its shape, and interface changes ripple across
all connectors.

The abstraction is well-conceived. The main work remaining is mechanical
(restructure into real Go packages, fix imports, make it compile) plus three
substantive additions: **event subscription**, **`Unlock`**, and **shared retry/
rate-limit middleware**.
