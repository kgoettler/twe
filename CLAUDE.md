# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build both binaries to ./bin/
make build

# Run all tests (sets TIMEWARRIORDB and TZ=America/New_York)
make test

# Run a specific test
TIMEWARRIORDB=$(PWD)/pkg/timewarrior/testdata/db TZ=America/New_York go test -v -run TestCLISuite/TestExport ./pkg/timewarrior/...

# Install twe to ~/.local/bin and the echo extension to TIMEWARRIORDB/extensions/
make install

# Lint
golangci-lint run
```

## Architecture

`twe` is a CLI tool built on top of [Timewarrior](https://timewarrior.net/). It ships two binaries:

- **`cmd/twe/`** — the main CLI (Cobra), with subcommands `edit`, `timecard`, `import`, and `last`
- **`cmd/echo/`** — a Timewarrior extension that echoes its stdin to stdout; installed into `TIMEWARRIORDB/extensions/` and invoked by `timew echo` so `twe timecard` can read interval data through the Timewarrior [Extension API](https://timewarrior.net/docs/api/)

### Package layout

| Path | Role |
|---|---|
| `pkg/timewarrior/` | Public library. `CLI` wraps `timew` shell commands; `Report` parses Extension API stdin (config key-value pairs + JSON intervals); `Interval`/`Datetime` are the core data types. |
| `internal/edit/` | Bubbletea TUI model for `twe edit`. `Model` owns a slice of `Row`s; each `Row` wraps an `Interval` and a set of `cell`s (backed by `textinput.Model`). All mutations go through the `TimewarriorBackend` interface so the real `CLI` can be swapped for a mock in tests. |
| `internal/timecard/` | Stateless report logic for `twe timecard`. `Run()` takes a `*timew.Report` + `TimecardOptions` and returns a formatted string. Data is aggregated into `TimecardData` (a `map[tag]map[date]duration` wrapper). |
| `internal/styles/` | Shared lipgloss styles used by both the edit TUI and timecard table. |

### Data flow for `twe timecard`

1. `cmd/twe/cmd/timecard.go` calls `timew echo <args>` via `CLI.Report()`, which pipes the Extension API output (config + JSON intervals) to an `io.Reader`.
2. `timew.NewReport(reader)` parses that into a `*Report`.
3. `timecard.Run(tw, options)` aggregates intervals by tag × date into `TimecardData` and renders a lipgloss table.

### Data flow for `twe edit`

1. `cmd/twe/cmd/edit.go` creates a `timew.CLI` backend and a `time.Time` date, then passes both to `edit.NewModel()`.
2. The Bubbletea program runs; user keystrokes either navigate the table (`handleTableNavigation`) or edit a cell (`handleEditing`).
3. On every confirmed cell edit, `Model.UpdateRow()` calls the appropriate `COLUMNS[j].Action` which calls the backend (`Modify`, `Retag`, `Annotate`, `Stop`, `Track`).

### Testing notes

- `pkg/timewarrior` tests require a real `timew` binary; `TIMEWARRIORDB` is pointed at `testdata/db/` (static fixture data) or a temp dir created by `CLISuite.SetupSuite`.
- `internal/edit` tests use `cursor_test.go`; the `TimewarriorBackend` interface in `model.go` is the seam for mocking.
- `internal/timecard` tests in `timecard_test.go` work entirely in-process with no external dependencies.
