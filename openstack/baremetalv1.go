package openstack

import (
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud"
	osp "github.com/gophercloud/gophercloud/openstack"
)

const BareMetalV1MicroVersion = "1.58"

type BareMetalV1 struct {
	Service
}

func newBareMetalV1(provider *gophercloud.ProviderClient) (*BareMetalV1, error) {
	if service, err := osp.NewBareMetalV1(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating bare metal v1 API client", "error", err)
		return nil, fmt.Errorf("failed to create bare metal v1 client: %w", err)
	} else {
		slog.Info("bare metal v1 client initialised")
		service.Microversion = BareMetalV1MicroVersion
		return &BareMetalV1{
			Service{
				provider: provider,
				client:   service,
			},
		}, err
	}
}
