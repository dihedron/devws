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
		ServerID string   `positional-arg-name:"SERVERID" required:"true"`
		Tags     []string `positional-arg-name:"TAG[S...]" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Delete) Execute(args []string) error {
	slog.Debug("running server tag delete command", "serverId", cmd.Args.ServerID, "tags", cmd.Args.Tags)

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	err = client.ComputeV2.DeleteTags(context.Background(), cmd.Args.ServerID, cmd.Args.Tags...)
	if err != nil {
		slog.Error("error deleting server tags", "error", err, "serverId", cmd.Args.ServerID, "tags", cmd.Args.Tags)
		return err
	}
	cmd.Output("ok")
	slog.Debug("all tags deleted")
	return nil
}
