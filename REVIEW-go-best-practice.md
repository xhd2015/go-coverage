# Go Best Practice Review — `github.com/xhd2015/go-coverage`

**Date:** 2026-08-06  
**Scope:** codebase structure, CLI design, flag handling, package layout  
**Reference:** [go-best-practice](https://github.com/xhd2015) skill topics — `cli/*`, `flags-parsing/*`, `cmd-exec`, `go-embed-assets`, `kool-create` (as applicable)  
**Mode:** review only (no implementation)

---

## Executive summary

This repo is primarily a **Go library** for semantic coverage-profile merging (AST clean → block map via myers/vscode diff → counter merge), with supporting packages for git snapshots, profiles, and line collection. That domain split is mostly sound.

Against go-best-practice, the weak areas are the **half-built CLIs**, a **custom `flag`+reflection dispatcher** instead of **less-flags**, and a **home-grown shell runner** (`sh`) instead of **`xgo/support/cmd`** with argv-based execution. There is also no first-class product CLI for the library’s main value (`merge`). JS assets ship via a large generated string file rather than `//go:embed`, which works for `go install` but diverges from the embed recipe.

`kool-create` is **not applicable** to this existing library; do not rescaffold.

---

## Codebase map (current)

| Area | Packages / paths | Role |
|------|------------------|------|
| Core merge | `merge/`, `profile/`, `model/`, `cover/`, `code/`, `line/` | Coverage merge pipeline |
| Diff | `diff/myers/`, `diff/vscode/` (+ `goja/`, `gen/`) | Block mapping |
| Git | `git/`, `git/gitops/` | Snapshots, renames, content |
| CLI entrypoints | `cmd/go-ast`, `cmd/go-ast-scan`, `code/go-ast` | AST tooling (incomplete) |
| Flag framework | `cmd/util/prog` | Custom bind + subcommand runner |
| Process helper | `sh/` | `bash -c` wrapper |
| Misc | `ast/`, `util/` | `packages.Load` helpers, path util |

```text
go-coverage/
├── merge/ profile/ model/ cover/ code/ line/   # domain library
├── diff/{myers,vscode}                         # diff backends
├── git/ gitops/                                # git integration
├── sh/ util/ ast/                              # infrastructure
└── cmd/{go-ast,go-ast-scan,util/prog}          # incomplete tools
```

**Build signal at review time:** `go build ./...` fails on `cmd/go-ast` (`ast` declared and not used). Other packages build.

---

## Findings (severity-ordered)

Severity guide:

- **P0** — broken or misleading public surface  
- **P1** — clear violation of go-best-practice with real maintenance/UX cost  
- **P2** — structural / consistency debt  
- **P3** — polish, docs, modernization  

---

### P0 — Product CLIs are incomplete or non-building

**Evidence**

| Path | Issue |
|------|--------|
| `cmd/go-ast/go-ast.go` | Does not compile (`ast` unused). `dump` is documented and implemented but **not** registered in `commands`. Help uses `fmt.Sprintf(...)` with **no print** — usage text is silent. Errors use discarded `fmt.Errorf("requires FILE")`. |
| `cmd/go-ast-scan/go-ast-scan.go` | Empty `main()` stub. |
| `code/go-ast/go-ast.go` | Second `package main` under a **library** tree (`code/`), duplicates AST-dump intent without subcommands or flags. |

**Why it matters (CLI UX / flags-parsing)**

- Users cannot discover or run a working tool surface.  
- Documented `dump` is unreachable after a successful build fix unless registered.  
- Empty-args / help do not match the subcommand checklist in `flags-parsing/subcommand` (“every level needs `--help`”; empty args → that level’s help).

**Recommended direction**

1. Decide product intent:
   - **A)** Library-only: delete or `//go:build ignore` unfinished mains; document library API only.  
   - **B)** Ship CLIs: one binary under `cmd/` (e.g. `go-coverage` or finish `go-ast`), with real subcommands.
2. If shipping a CLI, implement with **less-flags** + subcommand dispatch (see P1), not `cmd/util/prog`.
3. Do not keep `package main` under non-`cmd` library paths (`code/go-ast`).

**Topics:** `cli`, `flags-parsing/subcommand`

---

### P1 — Custom flag/subcommand stack instead of less-flags

**Evidence — `cmd/util/prog`**

- `Bind` uses `reflect` + struct tags `prog:"name default help"` and registers with stdlib `flag` (only `string` / `bool`).  
- `Run` peels the first argv token as a command, rewrites `os.Args`, then `flag.Parse()`.  
- Global `flag` package mutates process-global state; poor for tests and nested commands.  
- No built-in multi-alias (`-h,--help`), no `StopOnFirstArg`, no `Cut`, no `**T` unset detection, no `CollectParsedFlags`.  
- Tag format is space-split and fragile (`list[2]` is rest of help only if exactly three space-separated parts).

**Evidence — `cmd/go-ast` usage of it**

- Top-level `Prog` has no live flags (only a commented example).  
- Subcommands map is incomplete; default always calls broken help.  
- Example file `cmd/util/prog/example/example_program.go.txt` copies the same broken `fmt.Sprintf` help pattern.

**Recommended direction** (from `flags-parsing` / `flags-parsing/subcommand`)

```text
mytool [global-options] <command> [command-options] [args]
```

- Prefer **no top-level flags** → manual `-h`/`--help` + switch on `args[0]`.  
- Or top-level flags with `StopOnFirstArg()` then per-handler `lessflags.Parse`.  
- Wire `Help("-h,--help", …)` **at every command level**.  
- Empty args → print that level’s help (not panic / not silent).  
- Error to stderr; `os.Exit(1)` only at `main`.  
- Remove or freeze `cmd/util/prog` once migrated (do not grow the custom binder).

**Topics:** `flags-parsing`, `flags-parsing/types`, `flags-parsing/subcommand`, `flags-parsing/cut`, `flags-parsing/collect`

---

### P1 — External commands: shell strings and custom `sh` vs `cmd-exec`

**Evidence**

| Pattern | Where | Concern |
|---------|--------|---------|
| `exec.Command("bash", "-c", expr)` | `sh/sh.go` | Primary runner; always shell |
| `fmt.Sprintf("git -C %s ...", sh.Quote(...))` | `git/git.go`, `git/gitops/*`, `merge.GetModPath` | Shell quoting; harder to audit |
| Direct argv `exec.Command("git", "--git-dir=...", ...)` | `GitSnapshot.runGitCatFile`, committed `ls-files` | **Good** — align rest of package with this |
| `exec.Command("node", file)` | `diff/vscode` | Long-lived / one-shot Node; no shared runner |
| `Verbose: true` hard-coded | `gitops.RunCommand`, working-tree `ls-files` | Always logs command line via `log.Printf` |
| Always buffer stdout/stderr | `RunBashWithOpts` | No inherit-stream mode; opposes `cli/streaming` for interactive tools |

**Bugs / smells in `sh`**

- `bashCommandExpr` builds a `strings.Builder` then **ignores it** and returns `strings.Join(cmd, "\n")` — dead / misleading logic.  
- Errors stuff full command + stdout + stderr into one `error` string (noisy; hard for callers).  
- No `Dir`, `Env`, stdin, or discard controls comparable to `xgo/support/cmd`.

**Recommended direction** (from `cmd-exec`)

```go
// Prefer:
out, err := cmd.Output("git", "-C", dir, "diff", "--name-only", old, new)
// or for user-visible runs:
err := cmd.Debug().Dir(dir).Run("go", "mod", "edit", "-json")
```

Concrete migration map:

1. Replace `sh.RunBash*` call sites that are pure `git`/`go` with **argv** APIs (`cmd.Output` / `cmd.Debug().Dir(...).Run(...)`).  
2. Keep shell only where a pipeline is truly required; then still prefer `cmd` with explicit `bash -c` and documented reason.  
3. Drop hard-coded `Verbose: true` from library paths; use Debug only when the caller opts in.  
4. Retire `sh` (or shrink to `Quote` only if still needed) after migration.  
5. Node spawn in `diff/vscode` can stay specialized (IPC), but init/start errors should return errors instead of `panic` where possible, and lifecycle remains documented (`Destroy` / ping — already partially done).

**Topics:** `cmd-exec`, `cli/streaming` (when adding CLIs that run long jobs)

---

### P1 — No product CLI for the library’s main capability

**Evidence**

- README positions the project as a library (`merge` + `go test -run TestMergeProfile`).  
- No `cmd/go-coverage` (or similar) for: merge two profiles, merge across git commits, print summary.  
- Existing `cmd/*` tools are AST demos, not coverage merge.

**Recommended direction** (CLI UX + less-flags)

If users should run this without writing Go:

```text
go-coverage merge --old cov-old.out --new cov-new.out [--dir .] [--old-commit X] [--new-commit Y] [-o merged.out]
go-coverage --help
go-coverage merge --help
```

- Subcommand layout per `flags-parsing/subcommand`.  
- Stream progress only if work is long; for pure merge, print result or write `-o` (justify buffering for full profile rewrite — see `cli/streaming`).  
- Optional later: `--color` / `--no-color` if human summaries get ANSI (`cli/color`).  
- Dry-run only if the tool gains side effects (write files / git ops); use one pipeline with gates (`cli/dry-run`), not a duplicate dry-run path.

**Topics:** `cli`, `flags-parsing/subcommand`, `cli/streaming`, `cli/dry-run`, `cli/color` (as features appear)

---

### P2 — Package layout and module boundaries

**What works**

- Domain packages are largely MECE: `merge`, `profile`, `cover`, `code`, `line`, `model`, `diff/*`, `git`.  
- Diff backends are nested under `diff/` (good).  
- Goja opt-in is a separate import path (`diff/vscode/goja`) so Node is not forced on pure-Go users (good).

**Issues**

| Issue | Detail | Recommendation |
|-------|--------|----------------|
| `package ast` | Collides conceptually with `go/ast`; confuses imports | Rename to e.g. `pkgload` / `gopkgs` / `load` |
| `main` under library trees | `code/go-ast`, `diff/vscode/gen` | Mains only under `cmd/`; generators under `cmd/` or `internal/gen` with `//go:generate` |
| No `internal/` | Implementation helpers (`sh`, `prog`, git plumbing) are public API surface | Move non-contract packages to `internal/` |
| Overlap `git` vs `git/gitops` | Parallel command helpers | Consolidate behind one package; prefer argv git API |
| Catch-alls `util/`, `sh/` | Easy dumping grounds | Prefer named packages (`filepathx`, `gitcmd`) or `internal` |
| Dual AST tools | `cmd/go-ast` vs `code/go-ast` | Single entry under `cmd/` |
| Module `go 1.16` | Matches historical README constraint; deps (goja) need 1.16+ | Fine if intentional; document support window |

**kool-create:** N/A for this repo. Do not run `kool create` on an existing multi-package library.

**Topics:** package layout (general Go), `kool-create` (out of scope)

---

### P2 — JS assets: generate+string embed vs `//go:embed`

**Current approach**

1. `//go:generate bash gen.sh` → npm build → `go run ./gen/main.go`  
2. Writes `diff_js_gen.go` with ~228KB of `var diffJSCode = "..."`, `diffV2JSCode`, `DiffGojaCode`  
3. Runtime: write JS to temp dir, `exec.Command("node", …)` **or** run via goja string

**Assessment vs `go-embed-assets`**

| Concern | Current | go-embed-assets recipe |
|---------|---------|-------------------------|
| Bare `go install` / module zip | Works (strings are in source) | Works with committed placeholders or committed fat assets |
| Source readability / PR noise | Large generated Go string | Embed files; keep gen artifacts as `.js` or release hydrate |
| Empty embed root | N/A (no `//go:embed`) | Must keep `placeholder.txt` if embed roots gitignored |
| Runtime hydrate | N/A | Needed only if thin install + download fat UI |

**Recommended direction**

- **Short term (keep Node IPC):** switch generated payload to `//go:embed` of built JS under e.g. `diff/vscode/embedded/**` with either:
  - committed built JS (simple library), or  
  - placeholder + generate-into-embed for fat local builds (full recipe).  
- Prefer **goja path as default** for library consumers (no Node process leak / ping complexity); keep Node as optional backend.  
- Generator should live as `go run` tool under `cmd/` or `internal/`, not as an accidental importable main.  
- Ensure `go generate ./...` is documented as **dev-only**; consumers should not need Node for the default path.

**Topics:** `go-embed-assets`, `cmd-exec` (for generate scripts)

---

### P2 — CLI / library error and output conventions

**Evidence**

- Library code often `panic`s on programmer errors (`profile`, `merge` file-not-found callback, vscode output state). Acceptable for invariants; avoid for user/IO errors.  
- CLI paths mix `fmt.Printf` for errors (should be stderr), silent `fmt.Sprintf` for help, `os.Exit` deep in helpers.  
- `Verbose: true` in library git listing paths logs to process log even when used as a library.

**Recommended direction**

- Library: return `error`; never log at info by default.  
- CLI `main`: `if err := run(...); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }`.  
- Help and user messages: `fmt.Fprint(os.Stdout/os.Stderr, …)` explicitly.  
- When adding colored summaries: `cli/color` three-mode policy (`--color` / `--no-color` / `NO_COLOR` + TTY).

**Topics:** `cli`, `cli/color`, `cli/streaming`

---

### P3 — Docs, hygiene, and small consistency items

| Item | Notes |
|------|--------|
| README | Typos (“incrimental”, “muiltpe”, “LISCENSE” in cover README). Broken relative link `diff/myers/map.go` vs `mapping.go`. |
| README CLI | No install line (`go install …/cmd/…@…`); no real tool usage. |
| `ioutil` | Deprecated; fine on go1.16, migrate to `os`/`io` when bumping toolchain. |
| File mode `0777` on written temp JS | Prefer `0600` / `0644`. |
| `cover` package | Fork of std cover; license note present — keep attribution visible. |
| Diff README typo | `gitub.com` in import example. |
| Tests use `exec.Command("bash"…)` | Acceptable in tests; prefer argv `go test` without shell when easy. |

---

## Topic applicability matrix

| Topic | Applicable? | Notes |
|-------|-------------|--------|
| `flags-parsing` (+ types/subcommand/cut/collect) | **Yes** | Replace `cmd/util/prog`; any future product CLI |
| `cmd-exec` | **Yes** | Replace / shrink `sh`; unify git/`go mod` invocation |
| `cli` / color / streaming / dry-run | **Partial** | Needed once a real CLI ships; color/dry-run only with matching features |
| `cli/skill-cli` | No | Not a skill host |
| `cli/inline-tui-mouse` | No | No TUI |
| `go-embed-assets` | **Partial** | JS payloads; current string embed works but recipe prefers `//go:embed` + clear fat/thin story |
| `kool-create` | **No** | Existing project; do not rescaffold |

---

## Recommended change roadmap (implementation later)

Ordered for value vs risk; **do not implement in this review pass**.

### Phase 1 — Stop the bleeding (CLI surface)

1. Fix or remove non-building / empty mains so `go build ./...` is green.  
2. Choose library-only vs ship CLI; delete dead stubs accordingly.  
3. Move any remaining tools exclusively under `cmd/<name>`.

### Phase 2 — Flags and UX (if CLI remains)

1. Add `github.com/xhd2015/less-flags`.  
2. Implement subcommand dispatcher per `flags-parsing/subcommand` (help at every level).  
3. Delete or internalize `cmd/util/prog`.  
4. Optionally add `go-coverage merge …` as the real product command.

### Phase 3 — External commands

1. Introduce `xgo/support/cmd` for `git` / `go` invocations.  
2. Convert shell-formatted git helpers to argv (follow existing `cat-file` / `ls-files` pattern).  
3. Remove default verbose logging from library paths.  
4. Deprecate `sh` package.

### Phase 4 — Assets and layout

1. Prefer goja default; document Node as optional.  
2. Optionally migrate JS constants to `//go:embed` tree.  
3. Introduce `internal/` for non-API packages; rename `ast` package.  
4. Doc pass (README algorithm + CLI + generate workflow).

---

## Positive notes

- Clear domain model for incremental coverage merge; algorithm is documented at a high level.  
- Diff strategy is pluggable (`DiffImpl`, myers vs vscode vs goja).  
- Newer git content paths already use **direct argv** and worktree-resilient `--git-dir` resolution — good template for the rest of `git`.  
- Goja backend avoids requiring Node for pure-Go consumers.  
- Module intentionally targets go1.16+ with a documented floor.  
- Doctest-style scenarios under `git/tests/get-content/` show investment in edge cases (linked worktrees, missing files, concurrency).

---

## Appendix A — CLI design sketch (future)

Illustrative only; for a product CLI aligned with go-best-practice:

```go
// cmd/go-coverage/main.go (sketch)
func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run(args []string) error {
    if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
        fmt.Print(topHelp)
        return nil
    }
    switch args[0] {
    case "merge":
        return mergeCmd(args[1:])
    default:
        return fmt.Errorf("unknown command: %s", args[0])
    }
}

func mergeCmd(args []string) error {
    var oldPath, newPath, outPath string
    var dir string
    remain, err := lessflags.
        String("--old", &oldPath).
        String("--new", &newPath).
        String("-o,--output", &outPath).
        String("--dir", &dir).
        Help("-h,--help", mergeHelp).
        Parse(args)
    // validate, call merge.MergeGit / Merge, write profile
    _ = remain
    return err
}
```

## Appendix B — External command migration sketch

```go
// before
cmd := fmt.Sprintf(`git -C %s diff --diff-filter=M --name-only ...`, sh.Quote(dir), ...)
stdout, _, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{NeedStdOut: true})

// after (cmd-exec)
out, err := cmd.Output("git", "-C", dir, "diff", "--diff-filter=M", "--name-only",
    "--ignore-submodules", oldRef, newRef)
```

---

## Appendix C — Files most relevant to this review

| Path | Why |
|------|-----|
| `cmd/go-ast/go-ast.go` | Broken CLI |
| `cmd/go-ast-scan/go-ast-scan.go` | Empty CLI |
| `cmd/util/prog/{bind,run}.go` | Custom flags/subcommands |
| `code/go-ast/go-ast.go` | Misplaced main |
| `sh/sh.go` | Shell runner |
| `git/git.go`, `git/gitops/*` | Mixed argv vs shell |
| `merge/merge.go` | Product API; `GetModPath` shell |
| `diff/vscode/{diff.go,diff_v2.go,diff_js_gen.go,gen/}` | Node + generated assets |
| `README.md` | User-facing story |

---

*End of review. Implementation deferred until explicitly requested.*
