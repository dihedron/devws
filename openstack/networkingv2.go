package openstack

import (
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

const NetworkingV2MicroVersion = "2.35"

type NetworkingV2 struct {
	Service
}

func newNetworkingV2(provider *gophercloud.ProviderClient) (*NetworkingV2, error) {
	slog.Debug("initialising networking v2 client")
	if service, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating network v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create network v2 client: %w", err)
	} else {
		slog.Info("network v2 client initialised")
		service.Microversion = NetworkingV2MicroVersion
		return &NetworkingV2{
			Service{
				provider: provider,
				client:   service,
			},
		}, nil
	}
}
