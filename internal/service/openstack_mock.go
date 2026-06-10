package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/dihedron/devws/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type OpenstackMockService struct {
	client *openstack.Client
}

func NewOpenstackMockService(ctx context.Context) (*OpenstackMockService, error) {
	return &OpenstackMockService{}, nil
}

func (o *OpenstackMockService) List(ctx context.Context, options []openstack.ComputeV2ListOption) ([]openstack.Workstation, error) {
	slog.Info("List", "options", options)
	return readOpenstackData(), nil
}

func (o *OpenstackMockService) GetId(ctx context.Context, workstationNameOrID string) (string, error) {
	slog.Info("GetId", "workstationNameOrID", workstationNameOrID)
	return workstationNameOrID, nil
}

func (o *OpenstackMockService) Start(ctx context.Context, id string) error {
	slog.Info("mock - started workstation", "workstationId", id)
	return nil
}

func (o *OpenstackMockService) Stop(ctx context.Context, id string) error {
	slog.Info("mock - stopped workstation", "workstationId", id)
	return nil
}

func (o *OpenstackMockService) Reboot(ctx context.Context, id string, mode servers.RebootMethod) error {
	slog.Info("mock - rebooted workstation", "workstationId", id, "mode", mode)
	return nil
}

func (o *OpenstackMockService) Shelve(ctx context.Context, id string, offload bool) error {
	if offload {
		slog.Info("mock - rebooted workstation", "workstationId", id, "offload", offload)
	} else {
		slog.Info("mock - rebooted workstation", "workstationId", id)
	}
	return nil
}

func (o *OpenstackMockService) Unshelve(ctx context.Context, id string, availabilityZone string) error {
	slog.Info("mock - unshelve workstation", "workstationId", id, "availabilityZone", availabilityZone)
	return nil
}

func readOpenstackData() []openstack.Workstation {

	mock_path, found := os.LookupEnv("MOCK_OS_VM_COLLECTION")
	if !found {
		slog.Error("MOCK_OS_VM_COLLECTION env not found!")
		return []openstack.Workstation{}
	}

	jsonFile, err := os.Open(mock_path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var workstations []openstack.Workstation
	err = json.Unmarshal(byteValue, &workstations)
	if err != nil {
		log.Fatal(err)
	}

	return workstations
}
