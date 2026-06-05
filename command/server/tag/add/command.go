package add

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/server/base"
	"github.com/dihedron/devws/openstack"
)

type Add struct {
	base.Command
	Args struct {
		ServerID string   `positional-arg-name:"SERVERID" required:"true"`
		Tags     []string `positional-arg-name:"TAG[S...]" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Add) Execute(args []string) error {
	slog.Debug("running server tag add command", "serverId", cmd.Args.ServerID, "tags", cmd.Args.Tags)

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	err = client.ComputeV2.AddTags(context.Background(), cmd.Args.ServerID, cmd.Args.Tags...)
	if err != nil {
		slog.Error("error adding server tags", "error", err, "serverId", cmd.Args.ServerID, "tags", cmd.Args.Tags)
		return err
	}
	cmd.Output("ok")
	slog.Debug("all tags added")
	return nil
}
