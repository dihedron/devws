package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/dihedron/devws/openstack"
)

type OpenstackMockService struct {
	client *openstack.Client
}

func NewOpenstackMockService(ctx context.Context) (*OpenstackMockService, error) {
	return &OpenstackMockService{}, nil
}

func (o *OpenstackMockService) List(ctx context.Context, options []openstack.ComputeV2ListOption) ([]openstack.Workstation, error) {
	return readOpenstackData(), nil
}

func readOpenstackData() []openstack.Workstation {
	// Apri il file
	jsonFile, err := os.Open("mocked.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	// Leggi il contenuto
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	// Mappa nella struct
	var workstations []openstack.Workstation
	err = json.Unmarshal(byteValue, &workstations)
	if err != nil {
		log.Fatal(err)
	}

	return workstations
}
