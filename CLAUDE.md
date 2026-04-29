# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build binary
go build -o vault-env ./cmd/vault-env

# Build with version string
go build -ldflags "-X main.VersionString=1.0.0" -o vault-env ./cmd/vault-env

# Compile check all packages
go build ./...

# Run tests
go test ./...

# Tidy dependencies
go mod tidy
```

## Architecture

The binary is a process wrapper: authenticates to Vault, reads secrets from KV v2, injects them as environment variables, then `exec`s the target command. Signal forwarding (SIGHUP, SIGINT, SIGQUIT, SIGABRT, SIGTERM) ensures the child process receives OS signals correctly.

Running without arguments prints help. All errors are printed through logrus (colored `[ERROR]` prefix); urfave/cli's own error printing is suppressed via `ExitErrHandler`.

### Data flow

```
CLI flags / env vars → config.Config
    → auth.New()          → Authenticator (registry lookup by method name)
    → vault.NewClient()   → authenticates, holds *vault.Client
    → secret.NewKV2Loader() → reads each --path, returns map[string]string
    → os.Setenv() per key → optional $VAR expansion across all env
    → executor.Execute()  → starts child process, forwards signals
```

### Auth method registry (`internal/vault/auth/`)

Auth methods self-register via `init()` — no changes to `main.go` needed when adding a new method.

```go
// in a new file, e.g. auth/jwt.go
func init() {
    Register("jwt", func(cfg *config.Config) (Authenticator, error) {
        // validate and return Authenticator
    }, "--jwt-token") // declare which config fields this method uses
}
```

The registry automatically warns about auth-specific flags that are set but not claimed by the chosen method (e.g. passing `--role-id` when using `token` auth). `authFields` in `registry.go` is the single place to register new auth-specific flags for this check.

`auth.SupportedMethods()` is called at startup to populate the `--auth-method` flag usage string and error messages — no hardcoded method names outside the registry.

### Secret backend (`internal/vault/secret/`)

`secret.Backend` is the extension interface for secret sources:

```go
type Backend interface {
    Load(ctx context.Context, path string) (map[string]string, error)
}
```

`KV2Loader` is the only current implementation. The `--path` flag is the secret name within the mount (no mount prefix); `--mount` is separate.

### Logging (`internal/logger/`)

Custom logrus formatter with ANSI colors: timestamp in gray, bold colored level tag (`[DEBUG]`/`[INFO ]`/`[WARN ]`/`[ERROR]`), fields printed as `key=value` in cyan. Go's time format reference date applies — `"01:00:00"` means HH:MM:SS.

### Key decisions

- `--no-expand` disables `os.Expand` pass over all env vars after injection, which resolves `$VAR` references in secret values against the current environment.
- `Response[T].Auth.ClientToken` — vault-client-go wraps all responses in a generic `Response[T]`; auth responses populate `.Auth`, not `.Data`.
- `KvV2ReadResponse.Data` is `map[string]interface{}` — non-string values are coerced via `fmt.Sprintf`.
- `VersionString` is injected at build time via `-ldflags "-X main.VersionString=..."`.
