package view

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type View struct {
	base.Command
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *View) Execute(args []string) error {
	slog.Debug("running workstation view command")

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

	server, err := client.ComputeV2.View(ctx, id)
	if err != nil {
		slog.Error("error viewing workstation details", "error", err, "workstation", cmd.Args.WorkstationNameOrID)
		return err
	}

	cmd.Output(server)
	slog.Debug("workstation view command completed")
	return nil
}
