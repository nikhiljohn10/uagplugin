package plugintest

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
	"time"

	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
)

type Mode string

const (
	ModeSmoke  Mode = "smoke"
	ModeSource Mode = "source"
	ModeAll    Mode = "all"
)

func ModeFromString(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(ModeSource):
		return ModeSource
	case string(ModeAll):
		return ModeAll
	default:
		return ModeSmoke
	}
}

type RunConfig struct {
	BaseDir       string
	BuildDir      string // legacy fallback
	Name          string // legacy filter (deprecated)
	Files         []string
	SearchDirs    []string
	Timeout       time.Duration
	Mode          Mode
	Auth          models.AuthCredentials
	ContactParams models.ContactQueryParams
	LedgerParams  models.LedgerQueryParams
	JSON          bool
}

type FuncResult struct {
	Name    string        `json:"name"`
	Status  string        `json:"status"` // ok|missing|error|timeout|panic|skipped
	Error   string        `json:"error,omitempty"`
	Elapsed time.Duration `json:"elapsed_ms"`
}

type PluginResult struct {
	Name       string       `json:"name"`
	File       string       `json:"file"`
	Funcs      []FuncResult `json:"funcs"`
	SourceTest *FuncResult  `json:"source_test,omitempty"`
}

type RunResult struct {
	Plugins  []PluginResult `json:"plugins"`
	Failures int            `json:"failures"`
}

func Run(ctx context.Context, cfg RunConfig) RunResult {
	// Resolve files to test
	fileSet := map[string]struct{}{}
	var list []string
	// If explicit files provided
	for _, f := range cfg.Files {
		if f == "" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(f), ".so") {
			continue
		}
		if abs, err := filepath.Abs(f); err == nil {
			if st, err := os.Stat(abs); err == nil && !st.IsDir() {
				fileSet[abs] = struct{}{}
			}
		}
	}
	// If search dirs provided
	if len(fileSet) == 0 && len(cfg.SearchDirs) > 0 {
		for _, d := range cfg.SearchDirs {
			if d == "" {
				continue
			}
			abs := d
			if !filepath.IsAbs(abs) {
				if a, err := filepath.Abs(d); err == nil {
					abs = a
				}
			}
			matches, _ := filepath.Glob(filepath.Join(abs, "*.so"))
			for _, m := range matches {
				fileSet[m] = struct{}{}
			}
		}
	}
	// Legacy fallback to BuildDir + Name filter
	if len(fileSet) == 0 && cfg.BuildDir != "" {
		matches, _ := filepath.Glob(filepath.Join(cfg.BuildDir, "*.so"))
		for _, f := range matches {
			base := strings.TrimSuffix(filepath.Base(f), ".so")
			if cfg.Name == "" || strings.EqualFold(cfg.Name, base) {
				fileSet[f] = struct{}{}
			}
		}
	}
	for f := range fileSet {
		list = append(list, f)
	}
	sort.Strings(list)
	// Deduplicate by plugin basename to avoid loading logically identical plugins
	// from multiple locations (which can trigger "plugin already loaded").
	{
		seen := map[string]bool{}
		filtered := make([]string, 0, len(list))
		for _, f := range list {
			base := strings.TrimSuffix(filepath.Base(f), ".so")
			if seen[base] {
				continue
			}
			seen[base] = true
			filtered = append(filtered, f)
		}
		list = filtered
	}
	if len(list) == 0 {
		logger.Warn("No plugins found to test")
	}

	res := RunResult{}
	for _, f := range list {
		// Respect cancellation between plugins
		if ctx.Err() != nil {
			return res
		}
		pr := testOne(ctx, f, cfg)
		res.Plugins = append(res.Plugins, pr)
		for _, fr := range pr.Funcs {
			if fr.Status != "ok" && fr.Status != "missing" && fr.Status != "skipped" {
				res.Failures++
			}
		}
		if pr.SourceTest != nil {
			// Treat only error-like statuses as failures; "skipped" should not fail the run.
			switch pr.SourceTest.Status {
			case "ok", "skipped":
				// not a failure
			default:
				res.Failures++
			}
		}
	}
	return res
}

func testOne(ctx context.Context, file string, cfg RunConfig) PluginResult {
	base := strings.TrimSuffix(filepath.Base(file), ".so")
	pr := PluginResult{Name: base, File: file}

	// Open plugin
	p, err := plugin.Open(file)
	if err != nil {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "Open", Status: "error", Error: err.Error()})
		return pr
	}

	// First try the typed interface: Lookup("Plugin") of type typing.Plugin
	if sym, err := p.Lookup("Plugin"); err == nil && sym != nil {
		// Symbol can be the interface value directly, or a pointer to it.
		var impl typing.Plugin
		switch v := sym.(type) {
		case typing.Plugin:
			impl = v
		case *typing.Plugin:
			if v != nil {
				impl = *v
			}
		}
		if impl != nil {
			// Compatibility check using Meta()["contract_version"] if present
			if meta := impl.Meta(); meta != nil {
				if meta.ContractVersion != "" && !typing.IsCompatible(meta.ContractVersion) {
					pr.Funcs = append(pr.Funcs, FuncResult{Name: "Contract", Status: "error", Error: typing.IncompatibilityMessage(meta.ContractVersion)})
					return pr
				}
			}
			// Helper invoker using ctx timeout wrapper
			wrap := func(name string, f func() FuncResult) FuncResult {
				start := time.Now()
				done := make(chan FuncResult, 1)
				go func() { done <- f() }()
				select {
				case r := <-done:
					r.Elapsed = time.Since(start)
					return r
				case <-ctx.Done():
					return FuncResult{Name: name, Status: "timeout", Elapsed: time.Since(start)}
				case <-time.After(cfg.Timeout):
					return FuncResult{Name: name, Status: "timeout", Elapsed: time.Since(start)}
				}
			}

			// Meta
			pr.Funcs = append(pr.Funcs, wrap("Meta", func() FuncResult {
				_ = impl.Meta()
				return FuncResult{Name: "Meta", Status: "ok"}
			}))
			// Health
			pr.Funcs = append(pr.Funcs, wrap("Health", func() FuncResult {
				_ = impl.Health()
				return FuncResult{Name: "Health", Status: "ok"}
			}))
			// Optional RunTests
			if t, ok := any(impl).(typing.Tester); ok {
				pr.Funcs = append(pr.Funcs, wrap("RunTests", func() FuncResult {
					if err := t.RunTests(); err != nil {
						return FuncResult{Name: "RunTests", Status: "error", Error: err.Error()}
					}
					return FuncResult{Name: "RunTests", Status: "ok"}
				}))
			} else {
				pr.Funcs = append(pr.Funcs, FuncResult{Name: "RunTests", Status: "skipped"})
			}
			// Contacts
			pr.Funcs = append(pr.Funcs, wrap("Contacts", func() FuncResult {
				if _, err := impl.Contacts(cfg.Auth, cfg.ContactParams); err != nil {
					return FuncResult{Name: "Contacts", Status: "error", Error: err.Error()}
				}
				return FuncResult{Name: "Contacts", Status: "ok"}
			}))
			// Ledger (now core)
			pr.Funcs = append(pr.Funcs, wrap("Ledger", func() FuncResult {
				if _, err := impl.Ledger(cfg.Auth, cfg.LedgerParams); err != nil {
					return FuncResult{Name: "Ledger", Status: "error", Error: err.Error()}
				}
				return FuncResult{Name: "Ledger", Status: "ok"}
			}))

			// Source tests
			if cfg.Mode == ModeSource || cfg.Mode == ModeAll {
				st := runSourceTests(ctx, cfg.BaseDir, base)
				pr.SourceTest = &st
			}
			return pr
		}
	}

	// Legacy symbol path without reflection: type-assert known signatures
	wrap := func(name string, fn func() FuncResult) FuncResult {
		start := time.Now()
		done := make(chan FuncResult, 1)
		go func() { done <- fn() }()
		select {
		case r := <-done:
			r.Elapsed = time.Since(start)
			return r
		case <-ctx.Done():
			return FuncResult{Name: name, Status: "timeout", Elapsed: time.Since(start)}
		case <-time.After(cfg.Timeout):
			return FuncResult{Name: name, Status: "timeout", Elapsed: time.Since(start)}
		}
	}

	look := func(sym string) (any, bool) { s, err := p.Lookup(sym); return s, err == nil }

	// Meta
	if sym, ok := look("Meta"); ok {
		switch fn := sym.(type) {
		case func() models.MetaData:
			pr.Funcs = append(pr.Funcs, wrap("Meta", func() FuncResult {
				meta := fn()
				if v := meta.ContractVersion; v != "" && !typing.IsCompatible(v) {
					return FuncResult{Name: "Contract", Status: "error", Error: typing.IncompatibilityMessage(v)}
				}
				return FuncResult{Name: "Meta", Status: "ok"}
			}))
		default:
			pr.Funcs = append(pr.Funcs, FuncResult{Name: "Meta", Status: "error", Error: "invalid Meta signature"})
		}
	} else {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "Meta", Status: "missing"})
	}

	// Health
	if sym, ok := look("Health"); ok {
		switch fn := sym.(type) {
		case func() string:
			pr.Funcs = append(pr.Funcs, wrap("Health", func() FuncResult { _ = fn(); return FuncResult{Name: "Health", Status: "ok"} }))
		default:
			pr.Funcs = append(pr.Funcs, FuncResult{Name: "Health", Status: "error", Error: "invalid Health signature"})
		}
	} else {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "Health", Status: "missing"})
	}

	// Optional RunTests
	if sym, ok := look("RunTests"); ok {
		switch fn := sym.(type) {
		case func() error:
			pr.Funcs = append(pr.Funcs, wrap("RunTests", func() FuncResult {
				if err := fn(); err != nil {
					return FuncResult{Name: "RunTests", Status: "error", Error: err.Error()}
				}
				return FuncResult{Name: "RunTests", Status: "ok"}
			}))
		default:
			pr.Funcs = append(pr.Funcs, FuncResult{Name: "RunTests", Status: "skipped"})
		}
	} else {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "RunTests", Status: "skipped"})
	}

	// Contacts
	if sym, ok := look("Contacts"); ok {
		switch fn := sym.(type) {
		case func(models.AuthCredentials, models.ContactQueryParams) (*models.Contacts, error):
			pr.Funcs = append(pr.Funcs, wrap("Contacts", func() FuncResult {
				if _, err := fn(cfg.Auth, cfg.ContactParams); err != nil {
					return FuncResult{Name: "Contacts", Status: "error", Error: err.Error()}
				}
				return FuncResult{Name: "Contacts", Status: "ok"}
			}))
		default:
			pr.Funcs = append(pr.Funcs, FuncResult{Name: "Contacts", Status: "error", Error: "invalid Contacts signature"})
		}
	} else {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "Contacts", Status: "missing"})
	}

	// Ledger (core)
	if sym, ok := look("Ledger"); ok {
		switch fn := sym.(type) {
		case func(models.AuthCredentials, models.LedgerQueryParams) (*models.Ledger, error):
			pr.Funcs = append(pr.Funcs, wrap("Ledger", func() FuncResult {
				if _, err := fn(cfg.Auth, cfg.LedgerParams); err != nil {
					return FuncResult{Name: "Ledger", Status: "error", Error: err.Error()}
				}
				return FuncResult{Name: "Ledger", Status: "ok"}
			}))
		default:
			pr.Funcs = append(pr.Funcs, FuncResult{Name: "Ledger", Status: "error", Error: "invalid Ledger signature"})
		}
	} else {
		pr.Funcs = append(pr.Funcs, FuncResult{Name: "Ledger", Status: "missing"})
	}

	// If mode is "source", run `go test` in the plugin's directory
	if cfg.Mode == ModeSource || cfg.Mode == ModeAll {
		st := runSourceTests(ctx, cfg.BaseDir, base)
		pr.SourceTest = &st
	}

	return pr
}

// reflect-based invocation removed in favor of typed interfaces and type assertions.

func runSourceTests(ctx context.Context, baseDir, name string) FuncResult {
	// pkgs/<name>
	src := filepath.Join(baseDir, "plugins", "pkgs", name)
	if _, err := os.Stat(src); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FuncResult{Name: "go test", Status: "skipped", Error: "source dir not found"}
		}
		return FuncResult{Name: "go test", Status: "error", Error: err.Error()}
	}
	// Run `go test` with a timeout to avoid indefinite hangs.
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = src
	out, err := cmd.CombinedOutput()
	if err != nil {
		return FuncResult{Name: "go test", Status: "error", Error: strings.TrimSpace(string(out))}
	}
	return FuncResult{Name: "go test", Status: "ok"}
}

func PrintHuman(rr RunResult) {
	if len(rr.Plugins) == 0 {
		logger.Info("No plugins tested.")
		return
	}
	for _, p := range rr.Plugins {
		logger.Info("Plugin: %s (%s)", p.Name, p.File)
		for _, f := range p.Funcs {
			msg := f.Status
			if f.Error != "" {
				msg += ": " + f.Error
			}
			logger.Info("  - %s: %s (%s)", f.Name, msg, f.Elapsed.String())
		}
		if p.SourceTest != nil {
			st := p.SourceTest
			msg := st.Status
			if st.Error != "" {
				msg += ": " + st.Error
			}
			logger.Info("  - %s: %s", st.Name, msg)
		}
	}
	if rr.Failures > 0 {
		logger.Warn("Failures: %d", rr.Failures)
	} else {
		logger.Info("All tests passed")
	}
}
