package check

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/server/base"
	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/tags"
)

type Check struct {
	base.Command
	Args struct {
		ServerID string `positional-arg-name:"SERVERID" required:"true"`
		Tag      string `positional-arg-name:"TAG" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Check) Execute(args []string) error {
	slog.Debug("running server tag check command", "serverId", cmd.Args.ServerID, "tag", cmd.Args.Tag)

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	exists, err := tags.Check(context.Background(), client.ComputeV2.Client(), cmd.Args.ServerID, cmd.Args.Tag).Extract()
	if err != nil {
		slog.Error("error checking server tag", "error", err, "serverId", cmd.Args.ServerID, "tag", cmd.Args.Tag)
		return err
	}

	cmd.Output(exists)
	return nil
}
