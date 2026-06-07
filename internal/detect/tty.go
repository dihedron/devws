package detect

type Process struct {
	PID string
}

/*
func main() {
	// 1. Figure out the TTY of the current shell running this Go program
	currentTTY, err := getTTY("self")
	if err != nil {
		fmt.Printf("Failed to get current TTY: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Your Current TTY Number: %d\n\n", currentTTY)

	// 2. Walk the /proc directory
	entries, err := os.ReadDir("/proc")
	if err != nil {
		fmt.Printf("Failed to read /proc: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-8s | %-15s | %s\n", "PID", "NAME", "STATUS")
	fmt.Println(strings.Repeat("-", 60))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid := entry.Name()

		// /proc contains non-process directories (like /proc/sys).
		// We only care about directories that are purely numeric (PIDs).
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		procTTY, err := getTTY(pid)
		if err != nil {
			// Process likely exited while we were reading; skip safely
			continue
		}

		// 3. Classify the process
		if procTTY != currentTTY {
			// Grab the process name for nicer output
			commData, _ := os.ReadFile(filepath.Join("/proc", pid, "comm"))
			name := strings.TrimSpace(string(commData))

			status := "Different Terminal (tmux/screen/ssh)"
			if procTTY == 0 {
				status = "No Terminal (Daemon/Background)"
			}

			// For the sake of demonstration, let's only print tmux/screen related ones
			// or processes attached to *another* active terminal (not daemons).
			if procTTY != 0 {
				fmt.Printf("%-8s | %-15s | %s (tty_nr: %d)\n", pid, name, status, procTTY)
			}
		}
	}
}
*/
