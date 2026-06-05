package openstack

import (
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

const IdentityV3MicroVersion = "3.13"

type IdentityV3 struct {
	Service
}

func newIdentityV3(provider *gophercloud.ProviderClient) (*IdentityV3, error) {
	slog.Debug("initialising identity v3 client")
	if service, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating identity v3 API client", "error", err)
		return nil, fmt.Errorf("failed to create identity v3 client: %w", err)
	} else {
		slog.Info("identity v3 client initialised")
		service.Microversion = IdentityV3MicroVersion
		return &IdentityV3{
			Service{
				provider: provider,
				client:   service,
			},
		}, nil
	}
}
