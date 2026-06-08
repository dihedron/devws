package delete

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Delete struct {
	base.Command
	Args struct {
		WorkstationNameOrID string   `positional-arg-name:"WORKSTATION" required:"true"`
		Tags                []string `positional-arg-name:"TAG[S...]" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Delete) Execute(args []string) error {
	slog.Debug("running server tag delete command", "serverId", cmd.Args.WorkstationNameOrID, "tags", cmd.Args.Tags)

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

	err = client.ComputeV2.DeleteTags(ctx, id, cmd.Args.Tags...)
	if err != nil {
		slog.Error("error deleting server tags", "error", err, "workstationId", id, "tags", cmd.Args.Tags)
		return err
	}
	cmd.Output("ok")
	slog.Debug("all tags deleted")
	return nil
}
