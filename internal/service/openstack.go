package service

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type OpenstackServiceI interface {
	List(ctx context.Context, options []openstack.ComputeV2ListOption) ([]openstack.Workstation, error)
	GetId(ctx context.Context, workstationNameOrID string) (string, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Reboot(ctx context.Context, id string, mode servers.RebootMethod) error
	Shelve(ctx context.Context, id string, offload bool) error
	Unshelve(ctx context.Context, id string, availabilityZone string) error
}

type OpenstackService struct {
	client *openstack.Client
}

func NewOpenstackService(ctx context.Context) (*OpenstackService, error) {
	client, err := openstack.NewClient(ctx)
	if err != nil {
		slog.Error("error creating client", "error", err)
		return nil, err
	}
	return &OpenstackService{
		client: client,
	}, nil
}

func (o *OpenstackService) List(ctx context.Context, options []openstack.ComputeV2ListOption) ([]openstack.Workstation, error) {
	workstations, err := o.client.ComputeV2.List(ctx, options...)
	if err != nil {
		slog.Error("error listing workstations", "error", err)
		return nil, err
	}
	return workstations, nil
}

func (o *OpenstackService) GetId(ctx context.Context, workstationNameOrID string) (string, error) {
	id, err := o.client.ComputeV2.GetFirstID(ctx, workstationNameOrID)
	if err != nil {
		slog.Debug("error getting safe ID", "value", workstationNameOrID, "error", err)
		return "", err
	}
	return id, nil
}

func (o *OpenstackService) Start(ctx context.Context, id string) error {
	err := o.client.ComputeV2.Start(ctx, id)
	if err != nil {
		slog.Error("error starting workstation", "error", err, "workstationId", id)
		return err
	}
	return nil
}

func (o *OpenstackService) Stop(ctx context.Context, id string) error {
	err := o.client.ComputeV2.Stop(ctx, id)
	if err != nil {
		slog.Error("error stopping workstation", "error", err, "workstationId", id)
		return err
	}
	return nil
}

func (o *OpenstackService) Reboot(ctx context.Context, id string, mode servers.RebootMethod) error {
	err := o.client.ComputeV2.Reboot(ctx, id, mode)
	if err != nil {
		slog.Error("error rebooting workstation", "error", err, "workstationId", id, "mode", mode)
		return err
	}
	return nil
}

func (o *OpenstackService) Shelve(ctx context.Context, id string, offload bool) error {
	var err error
	if offload {
		err = o.client.ComputeV2.ShelveOffload(ctx, id)
	} else {
		err = o.client.ComputeV2.Shelve(ctx, id)
	}
	if err != nil {
		slog.Error("error shelving workstation", "error", err, "workstationId", id, "offload", offload)
		return err
	}
	return nil
}

func (o *OpenstackService) Unshelve(ctx context.Context, id string, availabilityZone string) error {
	err := o.client.ComputeV2.Unshelve(ctx, id, availabilityZone)
	if err != nil {
		slog.Error("error unpausing workstation", "error", err, "workstationId", id, "availaibilityZone", availabilityZone)
		return err
	}
	return nil
}
