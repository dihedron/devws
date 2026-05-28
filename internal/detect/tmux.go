package detect

import (
	"log/slog"
	"os/exec"
	"strings"
)

// Check if tmux has active sessions with process in execution
func HasTmuxActivity() bool {

	// List pane with current process
	out, err := exec.Command("tmux", "list-panes", "-a",
		"-F", "#{pane_current_command}").Output()
	if err != nil {
		slog.Info("no tmux process found")
		return false // tmux is not active
	}

	activeCommands := strings.Split(strings.TrimSpace(string(out)), "\n")
	idleShells := map[string]bool{"bash": true, "zsh": true, "sh": true, "fish": true}

	for _, cmd := range activeCommands {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" && !idleShells[cmd] {
			slog.Info("activity detected in tmux process")
			return true // something running
		}
	}
	slog.Debug("no active process in tmux")
	return false
}
