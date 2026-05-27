package daemon

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dihedron/slumberd/command/base"
	"github.com/dihedron/slumberd/internal/actions"
	"github.com/dihedron/slumberd/internal/detect"
	"github.com/dihedron/slumberd/timex"
)

type Daemon struct {
	base.Command
	Timeout  timex.Duration `short:"t" long:"timeout" description:"Timeout before the action execution when no activity found." default:"0s"`
	Interval timex.Duration `short:"i" long:"interval" description:"Interval between detections." default:"5s"`
	Action   string         `short:"a" long:"action" description:"Action to execute when no activity found." choice:"shutdown" choice:"hibernate" default:"shutdown"`
}

func (cmd *Daemon) Execute(args []string) error {
	slog.Debug("running run command")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(time.Duration(cmd.Interval))
	defer ticker.Stop()

	slog.Info("daemon started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return nil
		case <-ticker.C:
			if err := detectActivityHandler(cmd.Action, cmd.Timeout, cmd.Command.DryRun); err != nil {
				slog.Error("work failed", "error", err)
			}
		}
	}
}

func detectActivityHandler(action string, timeout timex.Duration, dryRun bool) error {
	slog.Debug("detectActivityHandler", "action", action, "timeout", timeout, "dryRun", dryRun)

	editors := detect.IsAnyEditorActive("/proc")
	if len(editors) == 0 {
		slog.Info("No active editors detected!")
		switch action {
		case "shutdown":
			callAction("PowerOff", timeout, dryRun)
		case "hibernate":
			callAction("Hibernate", timeout, dryRun)
		default:
			slog.Error("Action not supplied or implemented!")
		}

	}
	return nil
}

func callAction(command string, timeout timex.Duration, dryRun bool) error {
	if timeout > 0 {
		time.Sleep(time.Duration(timeout))
	}
	if dryRun {
		slog.Info("dry run:", "skipping", command)
		return nil
	}
	return actions.CallLogind(command)
}
