## Testing plugins

The CLI provides a `test` command to run smoke tests against compiled shared object plugins (`*.so`) and (optionally) native `go test` suites in plugin source directories. A lightweight `testkit` package helps plugin authors mock external services and control environment variables.

---

### 1. CLI test command

Command synopsis:

```
uagplugin test [path]
```

Where `[path]` is optional:

- Omitted: searches these directories (first existing) for `*.so` files:
  1. `./.uag/plugins/build`
  2. `$HOME/.uag/plugins/build`
  3. Fallback: build directory resolved by the CLI (same logic as install)
- File path to `plugin.so`: tests only that file
- Directory path: tests every `*.so` inside that directory (non-recursive)

Flags:
| Flag | Description |
|------|-------------|
| `--timeout <sec>` | Per function call timeout (default 5s) |
| `--env-file <file>` | Load additional environment variables from a `.env` file before tests |
| `--auth <json>` | JSON map passed as AuthCredentials argument where applicable |
| `--params <json>` | JSON object mapped to `models.Params` |
| `--mode smoke|source|all` | `smoke`: only symbols in the `.so`; `source`: only `go test` in source dir; `all`: both |
| `--json` | Emit structured JSON report instead of human logs |

Example invocations:

```
# Smoke test all discovered plugins in default locations
uagplugin test

# Test a single compiled plugin file
uagplugin test ./.uag/plugins/build/fileplugin.so

# Test all plugins inside a custom directory
uagplugin test /opt/uag/build-plugins

# Run smoke + native tests, load env vars, supply params & auth, output JSON
uagplugin test --mode all \
    --env-file .env.test \
    --timeout 10 \
    --auth '{"token":"abc123"}' \
    --params '{"search":"alice","sort":true}' \
    --json
```

Exit code:

- `0` when all tested functions either succeed, are missing (non-fatal), or skipped.
- `1` when any function returns error / timeout / panic or when `go test` fails.

JSON report shape (simplified):

```json
{
  "plugins": [
    {
      "name": "fileplugin",
      "file": "/abs/path/fileplugin.so",
      "funcs": [{ "name": "Meta", "status": "ok", "elapsed_ms": 123456 }],
      "source_test": { "name": "go test", "status": "ok" }
    }
  ],
  "failures": 0
}
```

Statuses per function:

- `ok`, `missing`, `skipped`, `error`, `timeout`, `panic`

---

### 2. What gets exercised in smoke mode

For each plugin file:

1. Load `Meta` (and read optional `contract_version` for compatibility)
2. Load `Health`
3. Optionally load and run `RunTests` (if exported by the plugin author)
4. Load & invoke: `Contacts` and `Ledger` (Ledger is required in the typed contract)

Arguments passed to `Contacts` / `Ledger`:

- `AuthCredentials`: parsed from `--auth` (default empty map)
- `models.Params`: parsed from `--params` (default zero-value)

Timeout & panic safety:

- Each function runs inside a goroutine with a deadline (`--timeout`), and respects global cancellation (Ctrl-C).
- Panics are caught and marked with status `panic` (stack logged in debug mode).

---

### 3. Native source tests (mode=source or all)

If the `.so` base name is `foo.so`, the CLI looks for the source directory:

```
~/.uag/plugins/pkgs/foo
```

and runs:

```
go test ./...
```

Failures surface as a single `go test` result entry in the report.

Note: `go test` is run with a bounded timeout and respects global cancellation.

---

### 4. typed contract and testkit package

Typed contract (recommended):

Plugins should export `var Plugin typing.Plugin = MyPlugin{}` and include `"contract_version": typing.ContractVersion` in `Meta()`. The runner validates compatibility before invoking functions.

Import path: `github.com/nikhiljohn10/uagplugin/testkit`

Provided helpers:

| Helper                                            | Description                                                      |
| ------------------------------------------------- | ---------------------------------------------------------------- |
| `StartMockServer(routes map[string]http.Handler)` | Spins up an HTTP test server and returns `(MockServer, baseURL)` |
| `JSONResponse(status int, payload any)`           | Convenience handler generating JSON output                       |
| `MockServer.Close()`                              | Stop the server (usually via `defer`)                            |
| `WithEnv(vars map[string]string, fn func())`      | Temporarily set env vars for the duration of `fn`                |

#### 4.1 Quick reference

```go
srv, base := testkit.StartMockServer(map[string]http.Handler{
    "/users": testkit.JSONResponse(200, []map[string]any{{"id":1, "name":"Alice", "email":"a@example.com"}}),
})
defer srv.Close()

testkit.WithEnv(map[string]string{"API_BASE_URL": base}, func() {
    out, err := Contacts(nil, models.Params{SearchText: "Alice"})
    // assertions...
})
```

#### 4.2 WithEnv semantics

`WithEnv` saves original values (or absence) of each key, sets new values, runs the closure, then restores originals—even if the closure panics.

#### 4.3 HTTP mocking patterns

Multiple endpoints:

```go
srv, base := testkit.StartMockServer(map[string]http.Handler{
    "/users": testkit.JSONResponse(200, []map[string]any{{"id": 1, "name": "Alice", "email": "a@ex.com"}}),
    "/health": testkit.JSONResponse(200, map[string]string{"status":"ok"}),
})
defer srv.Close()
```

Custom dynamic handler:

```go
srv, base := testkit.StartMockServer(map[string]http.Handler{
    "/users": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Query().Get("q") == "none" {
                    w.WriteHeader(200)
                    w.Write([]byte("[]"))
                    return
            }
            testkit.JSONResponse(200, []map[string]any{{"id":2, "name":"Bob", "email":"b@ex.com"}})(w, r)
    }),
})
```

#### 4.4 Example full test (API plugin)

```go
func TestContacts_FilterAndSort(t *testing.T) {
        srv, base := testkit.StartMockServer(map[string]http.Handler{
                "/users": testkit.JSONResponse(200, []map[string]any{
                        {"id": 2, "name": "Bob Builder", "email": "bob@example.com"},
                        {"id": 1, "name": "Alice Wonderland", "email": "alice@example.com"},
                }),
        })
        defer srv.Close()

        testkit.WithEnv(map[string]string{"API_BASE_URL": base}, func() {
                out, err := Contacts(nil, models.Params{Sort: true})
                if err != nil { t.Fatalf("Contacts error: %v", err) }
                if out.Count != 2 || out.Items[0].Name != "Alice Wonderland" {
                        t.Fatalf("unexpected sort order: %+v", out.Items)
                }
        })
}
```

#### 4.5 Common assertions

```go
if out.NextCursor != nil { t.Logf("next cursor: %s", *out.NextCursor) }
if out.Total < out.Count { t.Fatalf("total mismatch") }
```

---

### 5. Best practices

1. Keep plugin functions pure where possible—accept dependencies as parameters or read them from env so tests can override.
2. Use `WithEnv` to isolate test configuration; avoid mutating global state outside it.
3. Export `RunTests()` if you need custom smoke checks executed by the CLI without running the full native test suite.
4. Fail fast on external API errors so smoke test surfaces problems clearly.
5. Use `--json` in CI and parse the report to produce richer dashboards.

---

### 6. CI/CD integration sketch

```bash
uagplugin install --url github.com/your-org/uag-myplugin --name myplugin
uagplugin test --mode all --json > report.json
jq '.failures' report.json | grep '^0$' || exit 1
```

---

### 7. Troubleshooting

| Symptom           | Cause                   | Fix                                                          |
| ----------------- | ----------------------- | ------------------------------------------------------------ |
| `missing` status  | Symbol not exported     | Ensure function name is capitalized & in main package        |
| `timeout`         | Long-running call       | Increase `--timeout` or optimize code                        |
| `panic`           | Unhandled runtime error | Wrap risky code or validate inputs early                     |
| `go test` skipped | Source dir not found    | Verify plugin source path under `~/.uag/plugins/pkgs/<name>` |

---

### 8. Minimal example test (file plugin)

```go
func TestContacts_Search(t *testing.T) {
        out, err := Contacts(nil, models.Params{SearchText: "doe"})
        if err != nil { t.Fatalf("err: %v", err) }
        if out.Count == 0 { t.Fatalf("expected at least one match") }
}
```

---

### 9. Migration note (if you used --name before)

The `--name` flag was removed. Provide an explicit path or let discovery search default build directories.

---

### 10. Summary

Use `uagplugin test` for quick integration smoke checks of compiled plugins, and native `go test` + `testkit` for deeper unit & contract verification. Prefer the typed contract to simplify loading and compatibility checks.
