package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dihedron/slumberd/command/base"
	"github.com/dihedron/slumberd/internal/actions"
	"github.com/dihedron/slumberd/internal/monitor"
	"github.com/dihedron/slumberd/timex"
)

type Daemon struct {
	base.Command
	// Timeout before the action execution when no activity found.
	Timeout timex.Duration `short:"t" long:"timeout" description:"Timeout before the action execution when no activity found." default:"0s"`
	// Interval between detections.
	Interval timex.Duration `short:"i" long:"interval" description:"Interval between detections." default:"5s"`
	// Action to execute when no activity found.
	Action string `short:"a" long:"action" description:"Action to execute when no activity found." choice:"shutdown" choice:"hibernate" default:"shutdown"`
	// CPUThreshold CPU percentage threshold (0-100); 0 means no detection.
	CPUThreshold uint8 `short:"c" long:"cpu-threshold" description:"CPU threshold (0-100)." default:"0"`
}

func (cmd *Daemon) Execute(args []string) error {
	slog.Debug("running daemon command")

	if cmd.CPUThreshold > 100 {
		return fmt.Errorf("CPU threshold must be between 0 and 100, got %v", cmd.CPUThreshold)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(time.Duration(cmd.Interval))
	defer ticker.Stop()

	slog.Info("daemon started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("signal caught, shutting down...")
			return nil
		case <-ticker.C:
			if err := detectActivityHandler(cmd.Action, cmd.Timeout, cmd.CPUThreshold, cmd.Command.DryRun); err != nil {
				slog.Error("work failed", "error", err)
			}
		}
	}
}

func detectActivityHandler(action string, timeout timex.Duration, cpuThreshold uint8, dryRun bool) error {
	slog.Debug("detectActivityHandler", "action", action, "timeout", timeout, "dryRun", dryRun)

	cfg := monitor.Config{}
	if cpuThreshold > 0 {
		cfg.CPUThreshold = cpuThreshold
	}

	mon := monitor.New(cfg)
	res, err := mon.Check()
	if err != nil {
		slog.Error("error in monitor check", "error:", err)
		return err
	}

	if res.Decision == monitor.Busy {
		slog.Info("monitor check", "decision", res.Decision, "reason", res.Reasons)
		return nil
	}
	slog.Info("monitor check", "decision", res.Decision)

	if res.Decision == monitor.Safe {
		slog.Info("no activity detected!")

		switch action {
		case "shutdown":
			callAction("PowerOff", timeout, dryRun)
		case "hibernate":
			callAction("Hibernate", timeout, dryRun)
		default:
			slog.Error("Action not supplied or implemented!")
			return fmt.Errorf("Action not supplied or implemented!")
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

var (
	ErrNegativeThreshold   = errors.New("CPU threshold must be greater than 0")
	ErrThresholdTooHigh    = errors.New("CPU threshold cannot exceed 100")
	ErrThresholdWithoutCPU = errors.New("cpu-threshold provided but cpu check not enabled")
)
