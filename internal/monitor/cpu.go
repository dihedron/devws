package monitor

import (
	"fmt"
	"strings"
	"time"
)

// cpuSample holds two tick readings separated in time.
type cpuSample struct {
	utime, stime uint64
}

// cpuPct returns the CPU percentage for a process between two snapshots.
//
//	elapsed = wall clock seconds between the two samples
//	pct     = (Δutime + Δstime) / (elapsed * clockTick) * 100
func cpuPct(before, after cpuSample, elapsed time.Duration, clockTick uint64) float64 {
	delta := float64((after.utime + after.stime) - (before.utime + before.stime))
	totalTicks := elapsed.Seconds() * float64(clockTick)
	if totalTicks <= 0 {
		return 0
	}
	return delta / totalTicks * 100.0
}

// checkCPUUsage finds processes that are NOT in any shell session (i.e. they
// are detached / setsid'd / started by a daemon manager) and that consume
// more than cfg.CPUThreshold % CPU over cfg.CPUSampleInterval.
//
// This catches:
//   - `nohup make` or `setsid cargo build` left running detached
//   - systemd user services doing heavy work
//   - any other orphaned computation
func checkCPUUsage(
	snap *procSnapshot,
	cfg Config,
	ignored map[int]bool,
) (Reason, bool) {

	// Collect candidate PIDs: processes that are NOT session leaders of a
	// known shell session AND not a shell themselves.
	// We sample CPU for ALL non-shell, non-ignored processes to be safe.

	type candidate struct {
		info   *procInfo
		before cpuSample
	}
	var candidates []candidate

	for _, p := range snap.procs {
		if ignored[p.PID] {
			continue
		}
		if isShell(p.Comm, cfg.ExtraShells...) {
			continue
		}
		// Kernel threads have no cmdline; skip them.
		if len(p.Cmdline) == 0 && p.State == "S" {
			continue
		}
		candidates = append(candidates, candidate{
			info:   p,
			before: cpuSample{p.UTime, p.STime},
		})
	}

	if len(candidates) == 0 {
		return Reason{}, false
	}

	// Sleep for the sampling interval.
	time.Sleep(cfg.CPUSampleInterval)

	// Re-read stat for each candidate.
	var busyPIDs []int
	var details []string

	for _, c := range candidates {
		after, err := readProcPID(c.info.PID)
		if err != nil {
			continue // process exited during sleep – not busy
		}
		afterSample := cpuSample{after.UTime, after.STime}
		pct := cpuPct(c.before, afterSample, cfg.CPUSampleInterval, snap.clockTick)

		cfg.Logger.Debug("cpu sample",
			"pid", c.info.PID, "comm", c.info.Comm, "cpu_pct", fmt.Sprintf("%.1f", pct))

		if pct >= cfg.CPUThreshold {
			busyPIDs = append(busyPIDs, c.info.PID)
			details = append(details, fmt.Sprintf(
				"%s(pid=%d,cpu=%.1f%%,state=%s)",
				c.info.Comm, c.info.PID, pct, c.info.State,
			))
		}
	}

	if len(busyPIDs) == 0 {
		return Reason{}, false
	}
	return Reason{
		Check:   "cpu-sampling",
		Details: "CPU-intensive processes detected: " + strings.Join(details, ", "),
		PIDs:    busyPIDs,
	}, true
}
