package list

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type List struct {
	base.Command
	//ServerID string `short:"s" long:"server-id" description:"Status the virtual machine." required:"yes"`
	Args struct {
		WorkstationNameOrID string `positional-arg-name:"WORKSTATION" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *List) Execute(args []string) error {
	slog.Debug("running server tag list command", "serverId", cmd.Args.WorkstationNameOrID)

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

	tags, err := client.ComputeV2.ListTags(ctx, id)
	if err != nil {
		slog.Error("error listing server tags", "error", err, "workstationId", id)
		return err
	}
	cmd.Output(tags)
	return nil
}
