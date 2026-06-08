package shelve

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Shelve struct {
	base.Command
	Offload bool `short:"o" long:"offload" description:"Whether to also offload the workstation storage data." optional:"yes"`
	Args    struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Shelve) Execute(args []string) error {
	slog.Debug("running workstation shelve command")

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

	if cmd.Offload {
		err = client.ComputeV2.ShelveOffload(ctx, id)
	} else {
		err = client.ComputeV2.Shelve(ctx, id)
	}
	if err != nil {
		slog.Error("error shelving workstation", "error", err, "workstationId", id, "offload", cmd.Offload)
		return err
	}

	cmd.Output("ok")
	slog.Debug("workstation shelve command completed")
	return nil
}
