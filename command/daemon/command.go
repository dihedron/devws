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
	Timeout      timex.Duration `short:"t" long:"timeout" description:"Timeout before the action execution when no activity found." default:"0s"`
	Interval     timex.Duration `short:"i" long:"interval" description:"Interval between detections." default:"5s"`
	Action       string         `short:"a" long:"action" description:"Action to execute when no activity found." choice:"shutdown" choice:"hibernate" default:"shutdown"`
	CheckCPU     bool           `short:"c" long:"check-cpu" description:"Check CPU process pressure. If no --cpu-treshold is provided default 10% will be used."`
	CPUThreshold float64        `long:"cpu-threshold" description:"CPU threshold. if not provided default 10% will be used."`
}

func (cmd *Daemon) Execute(args []string) error {
	slog.Debug("running run command")

	err := cmd.Validate()
	if err != nil {
		fmt.Printf("%v\n", err.Error())
		os.Exit(1)
	}

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
			if err := detectActivityHandler(cmd.Action, cmd.Timeout, cmd.CheckCPU, cmd.CPUThreshold, cmd.Command.DryRun); err != nil {
				slog.Error("work failed", "error", err)
			}
		}
	}
}

func detectActivityHandler(action string, timeout timex.Duration, checkCPU bool, cpuThreshold float64, dryRun bool) error {
	slog.Debug("detectActivityHandler", "action", action, "timeout", timeout, "dryRun", dryRun)

	cfg := monitor.Config{}
	if checkCPU {
		cfg.CheckCPUProcesses = checkCPU
		if cfg.CPUThreshold > 0 {
			cfg.CPUThreshold = cpuThreshold
		}
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

func (d *Daemon) Validate() error {

	// Validate durations
	if d.Interval <= 0 {
		return fmt.Errorf("interval must be positive, got %v", d.Interval)
	}

	if d.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative, got %v", d.Timeout)
	}

	// Validate CPU configuration
	if d.CPUThreshold != 0 && !d.CheckCPU {
		return fmt.Errorf("%w: threshold=%.2f", ErrThresholdWithoutCPU, d.CPUThreshold)
	}

	if d.CheckCPU {
		// Validate threshold value if explicitly set
		if d.CPUThreshold != 0 {
			if d.CPUThreshold <= 0 {
				return fmt.Errorf("%w: got %.2f", ErrNegativeThreshold, d.CPUThreshold)
			}
			if d.CPUThreshold > 100 {
				return fmt.Errorf("%w: got %.2f", ErrThresholdTooHigh, d.CPUThreshold)
			}
		}
	}

	return nil
}
