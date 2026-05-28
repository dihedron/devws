package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// procInfo holds the parsed fields we care about from /proc/<pid>/.
type procInfo struct {
	PID     int
	PPID    int
	PGID    int    // process group ID
	SID     int    // session ID  (from /proc/<pid>/stat field 6)
	Comm    string // basename of executable
	State   string // R, S, D, Z, T …
	UTime   uint64 // user-mode ticks
	STime   uint64 // kernel-mode ticks
	Environ []string
	Cmdline []string
}

// procSnapshot is a point-in-time view of all visible processes.
type procSnapshot struct {
	procs     map[int]*procInfo
	clockTick uint64 // sysconf(_SC_CLK_TCK), typically 100
}

// scanProc reads /proc once and returns a snapshot.
func scanProc() (*procSnapshot, error) {
	snap := &procSnapshot{
		procs:     make(map[int]*procInfo),
		clockTick: 100, // safe default; accurate value via getClockTick()
	}
	snap.clockTick = getClockTick()

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("readdir /proc: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // not a PID dir
		}
		info, err := readProcPID(pid)
		if err != nil {
			continue // process may have exited; skip
		}
		snap.procs[pid] = info
	}
	return snap, nil
}

// readProcPID parses the fields we need for a single PID.
func readProcPID(pid int) (*procInfo, error) {
	base := fmt.Sprintf("/proc/%d", pid)

	// ── /proc/<pid>/stat ──────────────────────────────────────────────────
	statRaw, err := os.ReadFile(base + "/stat")
	if err != nil {
		return nil, err
	}
	info, err := parseStat(pid, string(statRaw))
	if err != nil {
		return nil, err
	}

	// ── /proc/<pid>/cmdline ───────────────────────────────────────────────
	cmdRaw, _ := os.ReadFile(base + "/cmdline")
	if len(cmdRaw) > 0 {
		// args are NUL-separated
		parts := strings.Split(strings.TrimRight(string(cmdRaw), "\x00"), "\x00")
		info.Cmdline = parts
	}

	// ── /proc/<pid>/environ (optional, may be unreadable) ─────────────────
	envRaw, _ := os.ReadFile(base + "/environ")
	if len(envRaw) > 0 {
		info.Environ = strings.Split(strings.TrimRight(string(envRaw), "\x00"), "\x00")
	}

	return info, nil
}

// parseStat parses /proc/<pid>/stat.
// Format: pid (comm) state ppid pgrp session ...
func parseStat(pid int, raw string) (*procInfo, error) {
	// The comm field may contain spaces and parentheses; find its boundaries.
	lp := strings.Index(raw, "(")
	rp := strings.LastIndex(raw, ")")
	if lp < 0 || rp < 0 || rp <= lp {
		return nil, fmt.Errorf("pid %d: malformed stat", pid)
	}
	comm := raw[lp+1 : rp]
	rest := strings.Fields(raw[rp+2:]) // skip ") "

	// rest[0]=state [1]=ppid [2]=pgrp [3]=session [4]=tty_nr …
	// [11]=utime [12]=stime (0-indexed after the closing paren)
	if len(rest) < 13 {
		return nil, fmt.Errorf("pid %d: stat too short (%d fields)", pid, len(rest))
	}

	ppid, _ := strconv.Atoi(rest[1])
	pgid, _ := strconv.Atoi(rest[2])
	sid, _ := strconv.Atoi(rest[3])
	utime, _ := strconv.ParseUint(rest[11], 10, 64)
	stime, _ := strconv.ParseUint(rest[12], 10, 64)

	return &procInfo{
		PID:   pid,
		PPID:  ppid,
		PGID:  pgid,
		SID:   sid,
		Comm:  filepath.Base(comm), // strip any path prefix in exotic cases
		State: rest[0],
		UTime: utime,
		STime: stime,
	}, nil
}

// children returns all direct children of parentPID in the snapshot.
func (s *procSnapshot) children(parentPID int) []*procInfo {
	var out []*procInfo
	for _, p := range s.procs {
		if p.PPID == parentPID {
			out = append(out, p)
		}
	}
	return out
}

// sessionMembers returns all processes with the given session ID.
func (s *procSnapshot) sessionMembers(sid int) []*procInfo {
	var out []*procInfo
	for _, p := range s.procs {
		if p.SID == sid {
			out = append(out, p)
		}
	}
	return out
}
