package show

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type Show struct {
	base.Command
	Args struct {
		ServerID string `positional-arg-name:"SERVERID" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *Show) Execute(args []string) error {
	slog.Debug("running vm list command")

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	server, err := client.ComputeV2.View(context.Background(), cmd.Args.ServerID)
	if err != nil {
		slog.Error("error viewing server details", "error", err, "serverId", cmd.Args.ServerID)
		return err
	}

	cmd.Output(server)
	slog.Debug("server show command completed")
	return nil
}
