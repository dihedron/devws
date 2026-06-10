package service

import (
	"context"
	"log/slog"

	"github.com/dihedron/devws/openstack"
)

type OpenstackServiceI interface {
	List(ctx context.Context, options []openstack.ComputeV2ListOption) ([]openstack.Workstation, error)
}

type OpenstackService struct {
	client *openstack.Client
}

func NewOpenStackService(ctx context.Context) (*OpenstackService, error) {
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
