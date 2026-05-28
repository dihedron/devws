package monitor

import (
	"os/exec"
	"strconv"
	"strings"
)

// getClockTick returns the kernel SC_CLK_TCK.
// We read it from /proc/self/stat indirectly: parse getconf output, or
// fall back to the universal default of 100 Hz.
func getClockTick() uint64 {
	out, err := exec.Command("getconf", "CLK_TCK").Output()
	if err == nil {
		v, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err == nil && v > 0 {
			return v
		}
	}
	return 100
}

// tmuxPaneShellPIDs uses `tmux list-panes` to get pane shell PIDs.
func tmuxPaneShellPIDs() ([]int, error) {
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_pid}").Output()
	if err != nil {
		return nil, err
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}
