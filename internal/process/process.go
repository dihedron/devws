package process

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Process struct {
	PID     string
	CmdLine string
	State   string
	PPID    string
	PGRP    string
	SID     string
	TTY     int
}

func ListAll(pattern string) ([]*Process, error) {
	var (
		re        *regexp.Regexp
		processes []*Process
		err       error
	)
	if len(pattern) > 0 {
		re, err = regexp.Compile(pattern)
		if err != nil {
			slog.Error("invalid regexp pattern", "pattern", pattern, "error", err)
			return nil, fmt.Errorf("invalid regexp pattern: %s: %w", pattern, err)
		}
	}

	files, err := os.ReadDir("/proc")
	if err != nil {
		slog.Error("failed to read /proc", "error", err)
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() || !isPID(f.Name()) {
			slog.Warn("ignoring directory or invalid name (non PID)", "name", f.Name())
			continue
		}
		slog.Debug("checking process", "pid", f.Name())
		filename := path.Clean(filepath.Join("/proc", f.Name(), "cmdline"))
		data, err := os.ReadFile(filename)
		if err != nil {
			slog.Warn("failed to read process command line", "pid", f.Name(), "error", err)
			continue
		}

		cmdline := strings.Replace(string(data), "\x00", " ", -1)
		if re == nil || re.MatchString(cmdline) {
			slog.Debug("found process", "filename", filename, "cmdline", cmdline)
			process, err := New(f.Name())
			if err != nil {
				slog.Error("error reading process statstics", "pid", f.Name(), "error", err)
				continue
			}

			processes = append(processes, process)
		}
	}

	slog.Debug("processes found", "count", len(processes))
	return processes, nil
}

func New(pid string) (*Process, error) {
	p := &Process{
		PID: pid,
	}
	if err := p.getStats(); err != nil {
		return nil, err
	}
	if err := p.getCmdLine(); err != nil {
		return nil, err
	}
	return p, nil
}

func isPID(name string) bool {
	_, err := strconv.Atoi(name)
	return err == nil
}

func (p *Process) getCmdLine() error {
	filename := path.Clean(filepath.Join("/proc", p.PID, "cmdline"))
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read cmdline file for process %s: %w", p.PID, err)
	}
	p.CmdLine = strings.Replace(string(data), "\x00", " ", -1)
	return nil
}

// getTTY extracts the tty_nr (7th field overall) from /proc/[pid]/stat
func (p *Process) getStats() error {
	statPath := filepath.Join("/proc", p.PID, "stat")
	data, err := os.ReadFile(statPath)
	if err != nil {
		return fmt.Errorf("failed to read stat file for process %s: %w", p.PID, err)
	}

	statStr := string(data)

	// Safely bypass the executable name which may contain spaces or parentheses
	endOfComm := strings.LastIndex(statStr, ")")
	if endOfComm == -1 {
		return fmt.Errorf("malformed stat file for process %s", p.PID)
	}

	// Extract everything after the ") "
	fields := strings.Fields(statStr[endOfComm+1:])
	if len(fields) < 5 {
		return fmt.Errorf("not enough fields in stat for process %s", p.PID)
	}

	// fields[0] = State (3rd field overall)
	// fields[1] = PPID  (4th field overall)
	// fields[2] = PGRP  (5th field overall)
	// fields[3] = SID   (6th field overall)
	// fields[4] = tty_nr (7th field overall)
	ttyNr, err := strconv.Atoi(fields[4])
	if err != nil {
		return fmt.Errorf("failed to parse tty_nr for process %s: %w", p.PID, err)
	}

	p.State = fields[0]
	p.PPID = fields[1]
	p.PGRP = fields[2]
	p.SID = fields[3]
	p.TTY = ttyNr

	return nil
}
