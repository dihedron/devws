package process

import (
	"strings"
	"testing"
)

func TestListAll(t *testing.T) {
	//processes, err := ListAll(".*test.*")
	processes, err := ListAll("")
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}

	t.Logf("processes found: %d", len(processes))
	for _, process := range processes {
		t.Logf("%s %s %s %s %s %s %d", process.PID, strings.TrimSpace(process.CmdLine), process.State, process.PPID, process.PGRP, process.SID, process.TTY)
	}
}
