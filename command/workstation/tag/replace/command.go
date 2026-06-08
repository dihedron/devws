package replace

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Replace struct {
	base.Command
	Args struct {
		WorkstationNameOrID string   `positional-arg-name:"WORKSTATION" required:"true"`
		Tags                []string `positional-arg-name:"TAG[S...]" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Replace) Execute(args []string) error {
	slog.Debug("running server tag replace command", "serverId", cmd.Args.WorkstationNameOrID, "tags", cmd.Args.Tags)

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
	//tags, err := tags.ReplaceAll(context.Background(), client.ComputeV2.Client(), cmd.Args.ServerID, tags.ReplaceAllOpts{Tags: cmd.Args.Tags}).Extract()
	tags, err := client.ComputeV2.ReplaceTags(ctx, id, cmd.Args.Tags...)
	if err != nil {
		slog.Error("error replacing all server tags", "error", err, "workstationId", id, "tags", cmd.Args.Tags)
		return err
	}
	cmd.Output(tags)
	slog.Debug("all tags replaced")
	return nil
}
