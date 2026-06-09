package unshelve

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Unshelve struct {
	base.Command
	AvailabilityZone string `short:"a" long:"availability-zone" description:"The availability zone into which to restore the workstation." required:"yes"`
	Args             struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Unshelve) Execute(args []string) error {
	slog.Debug("running workstation unshelve command")

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	ctx := context.Background()
	id, err := client.ComputeV2.GetFirstID(ctx, cmd.Args.WorkstationNameOrID)
	if err != nil {
		slog.Debug("error getting safe ID", "value", cmd.Args.WorkstationNameOrID, "error", err)
	}

	err = client.ComputeV2.Unshelve(ctx, id, cmd.AvailabilityZone)
	if err != nil {
		slog.Error("error unpausing workstation", "error", err, "workstationId", id, "availaibilityZone", cmd.AvailabilityZone)
		return err
	}

	cmd.Output("ok")
	slog.Debug("workstation unshelve command completed")
	return nil
}
