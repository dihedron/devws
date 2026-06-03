package list

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dihedron/devws/command/vm/base"
	"github.com/dihedron/devws/openstack"
	"gopkg.in/yaml.v3"
)

type List struct {
	base.Command
	// Name is the optional name or ID of the virtual machine.
	Name *string `short:"n" long:"name" description:"Name or ID of the virtual machine." optional:"yes"`
	// Owner is the optional owner of the virtual machine.
	Owner *string `short:"o" long:"owner" description:"Login ID of the virtual machine's owner." optional:"yes"`
	// Status is the optional status of the virtual machine.
	Status *string `short:"s" long:"status" description:"Status the virtual machine." optional:"yes"`
	// UserID is the ID of the user who created the virtual machine.
	UserID *string `short:"u" long:"userid" description:"ID of the user who created the virtual machine." optional:"yes"`
	// Image is the URL of the image used to create the virtual machine.
	Image *string `short:"i" long:"image" description:"The URL of the image used to create the virtual machine." optional:"yes"`
	// Image is the URL of the flavour of the virtual machine.
	Flavour *string `short:"f" long:"flavour" description:"The URL of the flavour of the virtual machine." optional:"yes"`
}

func (cmd *List) Execute(args []string) error {
	slog.Debug("running vm list command")
	client, err := openstack.NewClient(cmd.Cloud)
	if err != nil {
		slog.Error("error creating client", "error", err)
		return err
	}

	options := []openstack.ComputeV2ListOption{}
	if cmd.Name != nil {
		options = append(options, openstack.WithName(*cmd.Name))
	}
	if cmd.Owner != nil {
		options = append(options, openstack.WithTags(fmt.Sprintf("devws::owner: %s", *cmd.Owner)))
	}
	if cmd.Status != nil {
		options = append(options, openstack.WithStatus(*cmd.Status))
	}
	if cmd.UserID != nil {
		options = append(options, openstack.WithUserID(*cmd.UserID))
	}
	if cmd.Image != nil {
		options = append(options, openstack.WithImage(*cmd.Image))
	}
	if cmd.Flavour != nil {
		options = append(options, openstack.WithFlavour(*cmd.Flavour))
	}

	servers, err := client.ComputeV2.List(context.Background(), options...)
	if err != nil {
		slog.Error("error listing servers", "error", err)
		return err
	}

	switch cmd.Format {
	case "yaml":
		data, _ := yaml.Marshal(servers)
		fmt.Printf("%s\n", data)
	case "json":
		data, _ := json.MarshalIndent(servers, "", "  ")
		fmt.Printf("%s", string(data))
	case "test":
		fmt.Printf("%+v\n", servers)
	}
	slog.Debug("vm list command completed")
	return nil
}
