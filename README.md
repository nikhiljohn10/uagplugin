# UAG Plugin Tool

A CLI to manage UAG plugins: install from repos or local dirs, build, and test compiled plugins. This major release introduces a typed plugin contract, version compatibility checks, cancellation/timeouts, and better release metadata.

## Features

- Install plugins from a GitHub repo or a local directory
- Build plugins as Go shared objects (.so)
- Test compiled plugins (smoke) and optionally run native `go test` in source
- Typed plugin contract via `typing.Plugin`
- Contract version handshake and compatibility checks
- Signal-aware cancellation and bounded timeouts for HTTP and `go` subprocesses

> Note: Go `-buildmode=plugin` is not supported on Windows. Installing/testing plugins requires Linux or macOS.

## Installation

- From a release: download the archive for your OS/arch from GitHub Releases, extract the `uagplugin` binary and place it on your PATH.
- From source:

```
# build a debug binary
go build -o bin/uagplugin .

# or run directly
go run . --help
```

## Commands

- `uagplugin version` — prints app version/commit/date and contract version info
- `uagplugin install --dir <path>` — build a local plugin directory into `~/.uag/plugins/build`
- `uagplugin install --url github.com/org/repo --name <name>` — clone and build a repo (private supported with `--token`)
- `uagplugin test [path]` — run smoke tests on discovered `.so` files; with `--mode source|all` also run `go test`

See `docs/testing.md` for all flags and output details.

## Typed plugin contract

Plugins should export a single typed symbol that implements the contract:

```go
// in package main of the plugin
import "github.com/nikhiljohn10/uagplugin/typing"

var Plugin typing.Plugin = myPlugin{}

type myPlugin struct{}

func (myPlugin) Meta() map[string]any {
    return map[string]any{
        "platform_id": "myplugin",
        // Required for compatibility checks — set to typing.ContractVersion
        "contract_version": typing.ContractVersion,
    }
}

func (myPlugin) Health() string { return "ok" }
func (myPlugin) Contacts(auth map[string]string, p models.Params) (*models.Contacts, error) { /* ... */ }
func (myPlugin) Ledger(auth map[string]string, p models.Params) (models.Ledger, error) { /* ... */ }
```

- `Ledger` is now part of the core interface.
- Legacy top-level exported functions (Meta/Health/Contacts/Ledger) are still recognized via type assertions for backward compatibility, but new plugins should implement the typed contract.

## Contract versioning policy

- Host declares its contract version in `typing.ContractVersion` and minimum supported in `typing.MinSupportedContractVersion`.
- The runner reads `Meta()["contract_version"]` from the plugin:
  - Major must match host major.
  - Plugin version must be >= min supported.
  - Mismatches are reported as a `Contract` error and execution is aborted for that plugin.

When making changes:

- Backwards compatible additions: bump MINOR (e.g., 1.0.0 -> 1.1.0).
- Breaking changes: bump MAJOR (e.g., 1.x -> 2.0.0) and update `MinSupportedContractVersion` accordingly.

## Timeouts and cancellation

- Ctrl-C cancels in-flight operations (network, build, tests).
- HTTP calls (e.g., GitHub API, critical webhook) have short timeouts.
- `go build` and `go test` are run with bounded timeouts.

## Known limitations

- Go plugins aren’t supported on Windows.
- Plugins and host must share the same types (import paths) to interact.

## Migration guide (from pre-typed plugins)

- Prefer exporting `var Plugin typing.Plugin = MyPlugin{}`.
- Ensure `Meta()` adds `"contract_version": typing.ContractVersion`.
- Implement `Ledger` (now required by the typed interface). Legacy top-level functions will continue to work for now but may be deprecated in future versions.

## License

MIT
