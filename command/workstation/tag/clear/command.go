package clear

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Clear struct {
	base.Command
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Clear) Execute(args []string) error {
	slog.Debug("running server tag clear command", "serverId", cmd.Args.WorkstationNameOrID)

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

	err = client.ComputeV2.ClearTags(ctx, id)
	if err != nil {
		slog.Error("error clearing server tags", "error", err, "workstationId", id)
		return err
	}
	cmd.Output("ok")
	return nil
}
