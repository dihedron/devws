package openstack

import (
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

const BlockStorageV3MicroVersion = "3.59"

type BlockStorageV3 struct {
	Service
}

func newBlockStorageV3(provider *gophercloud.ProviderClient) (*BlockStorageV3, error) {
	slog.Debug("initialising block storage v3 client")
	if service, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating block storage v3 API client", "error", err)
		return nil, fmt.Errorf("failed to create block storage v3 client: %w", err)
	} else {
		slog.Info("block storage v3 client initialised")
		service.Microversion = BlockStorageV3MicroVersion
		return &BlockStorageV3{
			Service{
				provider: provider,
				client:   service,
			},
		}, nil
	}
}
