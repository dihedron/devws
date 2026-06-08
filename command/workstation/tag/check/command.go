package check

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Check struct {
	base.Command
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
		Tag                 string `positional-arg-name:"TAG" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Check) Execute(args []string) error {
	slog.Debug("running server tag check command", "serverId", cmd.Args.WorkstationNameOrID, "tag", cmd.Args.Tag)

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

	exists, err := client.ComputeV2.CheckTag(ctx, id, cmd.Args.Tag)
	if err != nil {
		slog.Error("error checking server tag", "error", err, "workstationId", id, "tag", cmd.Args.Tag)
		return err
	}

	cmd.Output(exists)
	return nil
}
