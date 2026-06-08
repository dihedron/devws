package resume

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Resume struct {
	base.Command
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Resume) Execute(args []string) error {
	slog.Debug("running workstation resume command")

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

	err = client.ComputeV2.Resume(ctx, id)
	if err != nil {
		slog.Error("error resuming workstation", "error", err, "workstationId", id)
		return err
	}

	cmd.Output("ok")
	slog.Debug("workstation resume command completed")
	return nil
}
