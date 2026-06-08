package list

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/dihedron/devws/command/workstation/base"
	"github.com/dihedron/devws/openstack"
)

type List struct {
	base.Command
	// Name is the optional name or ID of the virtual machine.
	With []string `short:"w" long:"with" description:"Whether to display only the given information items." choice:"id" choice:"name" choice:"address" choice:"addresses" choice:"volumes" optional:"yes"`
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

type brief struct {
	ID        *string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name      *string  `json:"name,omitempty" yaml:"name,omitempty"`
	Address   *string  `json:"address,omitempty" yaml:"address,omitempty"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Volumes   []string `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}

func (cmd *List) Execute(args []string) error {
	slog.Debug("running vm list command")

	cmd.Init()

	client, err := openstack.NewClient(context.Background())
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

	workstations, err := client.ComputeV2.List(context.Background(), options...)
	if err != nil {
		slog.Error("error listing workstations", "error", err)
		return err
	}

	if len(cmd.With) > 0 {
		results := []brief{}
		for _, workstation := range workstations {
			result := brief{}
			if slices.Contains(cmd.With, "id") {
				result.ID = new(workstation.ID)
			}
			if slices.Contains(cmd.With, "name") {
				result.Name = new(workstation.Name)
			}
			if slices.Contains(cmd.With, "address") {
				for _, address := range workstation.Addresses {
					for _, a := range address {
						result.Address = new(a.IPAddress)
						break
					}
				}
			}
			if slices.Contains(cmd.With, "addresses") {
				for _, address := range workstation.Addresses {
					for _, a := range address {
						result.Addresses = append(result.Addresses, a.IPAddress)
					}
				}
			}
			if slices.Contains(cmd.With, "volumes") {
				for _, volume := range workstation.AttachedVolumes {
					result.Volumes = append(result.Volumes, volume.ID)
				}
			}
			// TODO: add more information items if of interest
			results = append(results, result)
		}
		cmd.Output(results)
	} else {
		cmd.Output(workstations)
	}
	slog.Debug("server list command completed")
	return nil
}
