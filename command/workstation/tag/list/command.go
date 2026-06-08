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
		ServerID string `positional-arg-name:"SERVERID" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *List) Execute(args []string) error {
	slog.Debug("running server tag list command", "serverId", cmd.Args.ServerID)

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	tags, err := client.ComputeV2.ListTags(context.Background(), cmd.Args.ServerID)
	if err != nil {
		slog.Error("error listing server tags", "error", err, "serverId", cmd.Args.ServerID)
		return err
	}
	cmd.Output(tags)
	return nil
}
