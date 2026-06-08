package pause

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Pause struct {
	base.Command
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Pause) Execute(args []string) error {
	slog.Debug("running workstation pause command")

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

	err = client.ComputeV2.Pause(ctx, id)
	if err != nil {
		slog.Error("error pausing workstation", "error", err, "workstationId", id)
		return err
	}

	cmd.Output("ok")
	slog.Debug("workstation pause command completed")
	return nil
}
