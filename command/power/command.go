package power

import (
	"log/slog"
	"time"

	"github.com/dihedron/devws/command/base"
	"github.com/dihedron/devws/internal/actions"
	"github.com/dihedron/devws/timex"
)

// Shutdown is the command that shuts down the machine.
type Shutdown struct {
	base.Command
	// Timeout is the duration to wait before shutting down.
	Timeout timex.Duration `short:"t" long:"timeout" description:"Timeout before shutting down." default:"0s"`
}

// Execute is the real implementation of the Shutdown command.
func (cmd *Shutdown) Execute(args []string) error {
	slog.Info("requesting system shutdown", "timeout", cmd.Timeout.String())
	if cmd.Timeout > 0 {
		time.Sleep(time.Duration(cmd.Timeout))
	}
	if cmd.DryRun {
		slog.Info("dry run: skipping shutdown")
		return nil
	}
	return actions.CallLogind("PowerOff")
}

// Hibernate is the command that hibernates the machine.
type Hibernate struct {
	base.Command
	// Timeout is the duration to wait before hibernating.
	Timeout timex.Duration `short:"t" long:"timeout" description:"Timeout before hibernating." default:"0s"`
}

// Execute is the real implementation of the Hibernate command.
func (cmd *Hibernate) Execute(args []string) error {
	slog.Info("requesting system hibernation", "timeout", cmd.Timeout.String())
	if cmd.Timeout > 0 {
		time.Sleep(time.Duration(cmd.Timeout))
	}
	if cmd.DryRun {
		slog.Info("dry run: skipping hibernation")
		return nil
	}
	return actions.CallLogind("Hibernate")
}
