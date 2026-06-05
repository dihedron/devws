// Package monitor scans /proc to decide if the system is safe to shutdown or hibernate.
//
// Decision pipeline:
//
//	scan all /proc once → build session map
//	         │
//	         ├─ che if is any editor active
//	         │
//	         ├─ shell session has non-idle OtherPIDs? → busy (bg job)
//	         │
//	         ├─ any shell has foreground children?    → busy
//	         │
//	         ├─ tmux panes running non-shell?         → busy
//	         │
//	         └─ CPU sampling > threshold?             → busy (setsid/detached)
//	                    │
//	                    └─ all clear → safe to shutdown/hibernate

package monitor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/dihedron/devws/internal/detect"
)

// Decision is the outcome of a system activity check.
type Decision int

const (
	Safe    Decision = iota // no activity detected – ok to shutdown/hibernate
	Busy                    // activity detected – must wait
	Unknown                 // could not determine (e.g. permission error)
)

func (d Decision) String() string {
	switch d {
	case Safe:
		return "SAFE"
	case Busy:
		return "BUSY"
	default:
		return "UNKNOWN"
	}
}

// Reason records why a Decision was made.
type Reason struct {
	Check   string // which check fired
	Details string // human-readable explanation
	PIDs    []int  // relevant PIDs (if any)
}

// Result is returned by Monitor.Check.
type Result struct {
	Decision Decision
	Reasons  []Reason
	At       time.Time
}

type Config struct {
	CheckCPUProcesses bool

	// CPUThreshold is the per-process CPU % above which a process is
	// considered "busy". Measured over CPUSampleInterval. Default: 10.0
	CPUThreshold float64

	// CPUSampleInterval is how long we wait between two /proc/stat reads
	// to compute CPU %. Default: 500 ms.
	CPUSampleInterval time.Duration

	// ExtraShells lists additional shell binary names beyond the built-in set.
	ExtraShells []string

	// ExtraBusyComms lists additional comm names that are always considered busy
	// (e.g. "make", "ninja"). The built-in list already covers common compilers.
	ExtraBusyComms []string

	// IgnorePIDs lists PIDs that should never be flagged as busy.
	IgnorePIDs []int

	// Logger is used for debug output. If nil, logs are discarded.
	Logger *slog.Logger
}

func (c Config) withDefaults() Config {
	out := c
	if out.CPUThreshold == 0 {
		out.CPUThreshold = 10.0
	}
	if out.CPUSampleInterval == 0 {
		out.CPUSampleInterval = 500 * time.Millisecond
	}
	if out.Logger == nil {
		out.Logger = slog.Default()
	}
	return out
}

// Monitor - top-level checker.
type Monitor struct {
	cfg Config
}

// New creates a Monitor with the given configuration.
func New(cfg Config) *Monitor {
	return &Monitor{cfg: cfg.withDefaults()}
}

func (m *Monitor) Check() (Result, error) {
	result := Result{At: time.Now()}

	// Step 1: scan /proc once and build the session map
	snapshot, err := scanProc()
	if err != nil {
		result.Decision = Unknown
		return result, fmt.Errorf("scanProc: %w", err)
	}
	slog.Debug("proc scan complete", "processes", len(snapshot.procs))

	sessions := buildSessionMap(snapshot)
	slog.Debug("session map built", "sessions", len(sessions))

	ignoredSet := make(map[int]bool, len(m.cfg.IgnorePIDs))
	for _, pid := range m.cfg.IgnorePIDs {
		ignoredSet[pid] = true
	}

	// Step 2: check editors currently connected
	if editors := detect.IsAnyEditorActive("/proc"); len(editors) > 0 {
		result.Reasons = append(result.Reasons, Reason{
			Check:   "editors",
			Details: fmt.Sprintf("%v editors active", len(editors)),
		})
	}

	// Step 3: shell background jobs
	if r, busy := checkShellBackgroundJobs(sessions, m.cfg, ignoredSet); busy {
		result.Reasons = append(result.Reasons, r)
	}

	// Step 4: shell foreground children
	if r, busy := checkShellForegroundChildren(sessions, snapshot, m.cfg, ignoredSet); busy {
		result.Reasons = append(result.Reasons, r)
	}

	// Step 5: tmux panes running non-shell processes
	if r, busy := checkTmuxPanes(); busy {
		result.Reasons = append(result.Reasons, r)
	}

	if m.cfg.CheckCPUProcesses {
		// Step 6: [OPTIONAL] CPU sampling for detached/setsid processes
		if r, busy := checkCPUUsage(snapshot, m.cfg, ignoredSet); busy {
			result.Reasons = append(result.Reasons, r)
		}
	}

	if len(result.Reasons) == 0 {
		result.Decision = Safe
	} else {
		result.Decision = Busy
	}
	return result, nil
}
