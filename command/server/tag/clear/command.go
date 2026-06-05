package clear

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/server/base"
	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/tags"
)

type Clear struct {
	base.Command
	Args struct {
		ServerID string `positional-arg-name:"SERVERID" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Clear) Execute(args []string) error {
	slog.Debug("running server tag clear command", "serverId", cmd.Args.ServerID)

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	err = tags.DeleteAll(context.Background(), client.ComputeV2.Client(), cmd.Args.ServerID).ExtractErr()
	if err != nil {
		slog.Error("error clearing server tags", "error", err, "serverId", cmd.Args.ServerID)
		return err
	}

	cmd.Output("ok")
	return nil
}
