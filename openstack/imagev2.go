package openstack

import (
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud"
	osp "github.com/gophercloud/gophercloud/openstack"
)

const ImageV2MicroVersion = "2.9"

type ImageV2 struct {
	Service
}

func newImageV2(provider *gophercloud.ProviderClient) (*ImageV2, error) {
	if service, err := osp.NewImageServiceV2(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating image service v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create image service v2 client: %w", err)
	} else {
		slog.Info("image service v2 client initialised")
		service.Microversion = ImageV2MicroVersion
		return &ImageV2{
			Service{
				provider: provider,
				client:   service,
			},
		}, nil
	}
}
