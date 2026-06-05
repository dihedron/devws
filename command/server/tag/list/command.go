package list

import (
	"context"
	"log"
	"log/slog"

	"github.com/dihedron/devws/command/server/base"
	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/tags"
)

type List struct {
	base.Command
	//ServerID string `short:"s" long:"server-id" description:"Status the virtual machine." required:"yes"`
	Args struct {
		ServerID string `positional-arg-name:"SERVERID" required:"true"`
	} `positional-args:"true" required:"true"`
}

func (cmd *List) Execute(args []string) error {
	slog.Debug("running vm list command")

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	serverTags, err := tags.List(context.Background(), client.ComputeV2.Client(), cmd.Args.ServerID).Extract()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Output(serverTags)
	return nil
}
