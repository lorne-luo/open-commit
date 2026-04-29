# LLM Fallback (Dual-Provider) Design

Date: 2026-04-29
Status: Approved (pending implementation)

## Goal

Allow opencommit to be configured with **two** LLM providers (api1 and api2) and
automatically fall back to the other provider when the active one fails. The
last successful provider is remembered so subsequent invocations start with the
provider that worked last.

## Configuration

New keys in `~/.config/opencommit/config.toml`:

```toml
[api]
key          = "..."
baseurl      = "..."
model        = "..."
last_success = 1          # 1 or 2; default 1

[api2]
key     = "..."
baseurl = "..."
model   = "..."
```

Registered keys (extends `cmd/config/set.go` `ValidConfigKeys`):

| Key                | Type   |
|--------------------|--------|
| `api.last_success` | int    |
| `api2.key`         | string |
| `api2.baseurl`     | string |
| `api2.model`       | string |

`config get` / `config set` help text updated to describe the new keys.

## CLI Flags

`--model`, `--baseurl`, `-m` continue to override **only** the primary provider
(api1). No new flags for api2 — it is configured via `config set` only.

## Data Model

```go
// internal/service/provider.go
type ProviderConfig struct {
    ID      int    // 1 or 2 — used to update last_success
    Key     string
    BaseURL string
    Model   string
}
```

A request carries an ordered slice `[]ProviderConfig` — the first entry is the
preferred one (per `last_success`). The slice has 1 or 2 elements; api2 is only
included when `api2.key` is non-empty.

## Selection Order

Pseudo-code in usecase:

```
primary   = ProviderConfig{ID:1, Key:api.key, BaseURL:api.baseurl, Model:api.model}
secondary = ProviderConfig{ID:2, Key:api2.key, BaseURL:api2.baseurl, Model:api2.model}

candidates = [primary]
if secondary.Key != "" { candidates = append(candidates, secondary) }

if last_success == 2 && len(candidates) == 2 {
    candidates = [secondary, primary]
}
```

CLI flag overrides (`--model`, `--baseurl`) are applied to whichever entry has
`ID == 1` after ordering.

## Fallback Execution

`chatComplete` is replaced by a provider-aware version:

```go
func chatComplete(
    ctx context.Context,
    providers []ProviderConfig,
    systemPrompt, userPrompt string,
    onSuccess func(usedID int),
) (string, error)
```

Behavior:

1. For each provider in order, build an `*openai.Client` and attempt the call
   with the existing 2-attempt internal retry.
2. On success: invoke `onSuccess(provider.ID)`, return text.
3. On failure: record the error and continue to the next provider.
4. If all providers fail, return the last error.

`onSuccess` is responsible for persisting `api.last_success` — it only writes
when the used ID differs from the currently-stored value, to avoid touching
disk on every call. Persistence uses `viper.Set` + `viper.WriteConfig`.

Client construction moves out of `RootUsecase.initializeAIClient` and into the
fallback helper (one client per provider attempt). The single
`*openai.Client` parameter currently threaded through `AIService` methods is
removed; methods receive `[]ProviderConfig` instead.

## Affected Files

- `cmd/config/set.go` — add keys + help text
- `cmd/config/get.go` — add keys to help text
- `internal/service/config_service.go` — no change (defaults still apply to api1)
- `internal/service/provider.go` — **new** — `ProviderConfig`, fallback helper, last_success persistence
- `internal/service/ai_service.go` — `chatComplete` becomes a thin wrapper over the provider helper; method signatures swap `*openai.Client` for `[]ProviderConfig`
- `internal/usecase/root_usecase.go` — build providers from viper, drop `initializeAIClient`
- `internal/usecase/pr_usecase.go` — same
- `internal/delivery/cli/handler/root_handler.go` — surface api2 in the "no key" error path (still requires `api.key`; api2 is optional)
- `internal/delivery/cli/handler/pr_handler.go` — same

## Error Surface

- If `api.key` is empty: existing error is preserved.
- If `api2.key` is empty: fallback is silently disabled — single-provider behavior.
- If all providers fail: error message includes which providers were tried, e.g.
  `all providers failed: api1: <err>; api2: <err>`.

## Testing

Manual verification:

1. `go build ./...` and `go vet ./...` pass.
2. `opencommit config set api2.key xxx` writes to TOML.
3. `opencommit config get` lists new keys.
4. With a valid api1 only → behaves as before.
5. With an invalid api1 + valid api2 → succeeds via api2, `last_success` becomes 2.
6. Next run starts with api2 first.

## Out of Scope

- More than two providers.
- Health-check / circuit-breaker beyond per-call retry.
- CLI flags for api2 (set via config only).
- Concurrent provider racing.
