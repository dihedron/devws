package reboot

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type Reboot struct {
	base.Command
	// Mode indicates the reboot mode, whether hard or soft.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Mode string `short:"m" long:"mode" description:"Reboot mode, hard or soft." choice:"hard" choice:"soft" optional:"yes" default:"soft"`
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Reboot) Execute(args []string) error {
	slog.Debug("running workstation reboot command")

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	mode := servers.SoftReboot
	if cmd.Mode == "hard" {
		mode = servers.HardReboot
	}

	ctx := context.Background()
	id, err := client.ComputeV2.GetFirstID(ctx, cmd.Args.WorkstationNameOrID)
	if err != nil {
		slog.Debug("error getting safe ID", "value", cmd.Args.WorkstationNameOrID, "error", err)
	}

	err = client.ComputeV2.Reboot(ctx, id, mode)
	if err != nil {
		slog.Error("error rebooting workstation", "error", err, "workstationId", id, "mode", cmd.Mode)
		return err
	}

	cmd.Output("ok")
	slog.Debug("workstation reboot command completed")
	return nil
}
