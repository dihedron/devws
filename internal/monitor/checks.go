package monitor

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Check 1 - Shell background jobs
// A shell session has a background job when there are non-shell processes
// in the same SID that are not in an idle state (or are known-busy comms).
// ─────────────────────────────────────────────────────────────────────────────

func checkShellBackgroundJobs(
	sessions map[int]*shellSession,
	cfg Config,
	ignored map[int]bool,
) (Reason, bool) {

	var busyPIDs []int
	var details []string

	for _, sess := range sessions {
		if sess.Shell == nil {
			continue
		}
		for _, p := range sess.OtherPIDs {
			if ignored[p.PID] {
				continue
			}
			// Skip other shells (sub-shells are fine on their own).
			if isShell(p.Comm, cfg.ExtraShells...) {
				continue
			}
			// Flag if: known-busy comm OR process is not idle.
			if isBusyComm(p.Comm, cfg.ExtraBusyComms) || !isIdleState(p.State) {
				busyPIDs = append(busyPIDs, p.PID)
				details = append(details, fmt.Sprintf("%s(pid=%d,state=%s)", p.Comm, p.PID, p.State))
			}
		}
	}

	if len(busyPIDs) == 0 {
		return Reason{}, false
	}
	return Reason{
		Check:   "shell-bg-jobs",
		Details: "background jobs in shell sessions: " + strings.Join(details, ", "),
		PIDs:    busyPIDs,
	}, true
}

// ─────────────────────────────────────────────────────────────────────────────
// Check 2 - Shell foreground children
// A shell's foreground child is a process whose PGID equals the shell's PGID
// (or any PGID that is the terminal's foreground process group).
// We detect it by looking for direct children of the shell that are not
// themselves shells and are in an active state.
// ─────────────────────────────────────────────────────────────────────────────

func checkShellForegroundChildren(
	sessions map[int]*shellSession,
	snap *procSnapshot,
	cfg Config,
	ignored map[int]bool,
) (Reason, bool) {

	var busyPIDs []int
	var details []string

	for _, sess := range sessions {
		if sess.Shell == nil {
			continue
		}
		children := snap.children(sess.Shell.PID)
		for _, child := range children {
			if ignored[child.PID] {
				continue
			}
			if isShell(child.Comm, cfg.ExtraShells...) {
				continue
			}
			// A foreground child: PGID == child's own PGID and state is active.
			// In practice any direct child of the shell running in R/D is
			// a foreground process (shells wait() for them).
			if !isIdleState(child.State) || isBusyComm(child.Comm, cfg.ExtraBusyComms) {
				busyPIDs = append(busyPIDs, child.PID)
				details = append(details, fmt.Sprintf(
					"%s(pid=%d,ppid=%d,state=%s)",
					child.Comm, child.PID, child.PPID, child.State,
				))
			}
		}
	}

	if len(busyPIDs) == 0 {
		return Reason{}, false
	}
	return Reason{
		Check:   "shell-fg-children",
		Details: "foreground children in shell sessions: " + strings.Join(details, ", "),
		PIDs:    busyPIDs,
	}, true
}

// ─────────────────────────────────────────────────────────────────────────────
// Check 3 - Tmux panes
//
// uses tmux command: tmux list-panes -a -F #{pane_current_command}
// to detect running process in tmux terminal
// ─────────────────────────────────────────────────────────────────────────────

func checkTmuxPanes() (Reason, bool) {

	// List pane with current process
	out, err := exec.Command("tmux", "list-panes", "-a",
		"-F", "#{pane_current_command}").Output()
	if err != nil {
		slog.Info("no tmux process found")
		return Reason{
			Check:   "tmux-panes",
			Details: "tmux is not active",
		}, false // tmux is not active
	}

	activeCommands := strings.Split(strings.TrimSpace(string(out)), "\n")
	idleShells := map[string]bool{"bash": true, "zsh": true, "sh": true, "fish": true}

	for _, cmd := range activeCommands {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" && !idleShells[cmd] {
			slog.Info("activity detected in tmux process")
			return Reason{
				Check:   "tmux-panes",
				Details: "active processes inside tmux panes",
			}, true // something running
		}
	}

	return Reason{
		Check:   "tmux-panes",
		Details: "no active process in tmux",
	}, false
}
