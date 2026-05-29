package openstack

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/gophercloud/gophercloud"
	osp "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

const DefaultCloud = "openstack"

type Service struct {
	provider *gophercloud.ProviderClient
	client   *gophercloud.ServiceClient
}

type Client struct {
	client         *gophercloud.ProviderClient
	BareMetalV1    *BareMetalV1
	BlockStorageV3 *BlockStorageV3
	ComputeV2      *ComputeV2
	IdentityV3     *IdentityV3
	ImageV2        *ImageV2
	NetworkingV2   *NetworkingV2
}

func NewClient(cloud string) (*Client, error) {
	if cloud == "" {
		cloud = DefaultCloud
	}
	if c, ok := os.LookupEnv("OS_CLOUD"); ok {
		slog.Debug("using custom cloud as per the OS_CLOUD environment variable", "cloud", c)
		cloud = c
	}

	opts, err := clientconfig.AuthOptions(&clientconfig.ClientOpts{
		Cloud: cloud,
	})
	if err != nil {
		slog.Error("failed to get auth options", "error", err)
		return nil, fmt.Errorf("failed to get auth options: %w", err)
	}

	return newClient(*opts)
}

func NewClientFromEnv() (*Client, error) {
	opts, err := osp.AuthOptionsFromEnv()
	if err != nil {
		slog.Error("failed to get auth options from enviroment", "error", err)
		return nil, fmt.Errorf("failed to get auth options from environment: %w", err)
	}
	return newClient(opts)
}

var DefaultRegion = "RegionOne"

func init() {
	if region, ok := os.LookupEnv("OS_REGION_NAME"); ok {
		slog.Info("using region from OS_REGION instead of default", "region", region, "default", DefaultRegion)
		DefaultRegion = region
	}
}

const (
	BareMetalV1MicroVersion    = "1.58"
	BlockStorageV3MicroVersion = "3.59"
	IdentityV3MicroVersion     = "3.13"
	ImageV2MicroVersion        = "2.9"
	NetworkingV2MicroVersion   = "2.35"
)

func newClient(opts gophercloud.AuthOptions) (*Client, error) {
	opts.AllowReauth = true
	provider, err := osp.AuthenticatedClient(opts)
	if err != nil {
		slog.Error("error creating authenticated client", "error", err)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	ar := provider.GetAuthResult()
	data, _ := json.MarshalIndent(ar, "", "  ")
	slog.Debug("auth result", "data", data)

	client := &Client{}

	// initialize Bare Metal client
	if service, err := osp.NewBareMetalV1(provider, gophercloud.EndpointOpts{Region: DefaultRegion}); err != nil {
		slog.Error("error creating bare metal v1 API client", "error", err)
		return nil, fmt.Errorf("failed to create bare metal v1 client: %w", err)
	} else {
		slog.Info("bare metal v1 client initialised")
		service.Microversion = BareMetalV1MicroVersion
		client.BareMetalV1 = &BareMetalV1{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}

	// initialise Cinder client
	if service, err := osp.NewBlockStorageV3(provider, gophercloud.EndpointOpts{Region: "RegionOne"}); err != nil {
		slog.Error("error creating block storage v3 API client", "error", err)
		return nil, fmt.Errorf("failed to create block storage v3 client: %w", err)
	} else {
		slog.Info("block storage v3 client initialised")
		service.Microversion = BlockStorageV3MicroVersion
		client.BlockStorageV3 = &BlockStorageV3{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}

	// initialize Nova client
	if service, err := osp.NewComputeV2(provider, gophercloud.EndpointOpts{Region: DefaultRegion}); err != nil {
		slog.Error("error creating compute v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create compute v2 client: %w", err)
	} else {
		slog.Info("compute v2 client initialised")
		service.Microversion = ComputeV2MicroVersion
		client.ComputeV2 = &ComputeV2{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}

	// initialize Keystone client
	if service, err := osp.NewIdentityV3(provider, gophercloud.EndpointOpts{Region: DefaultRegion}); err != nil {
		slog.Error("error creating identity v3 API client", "error", err)
		return nil, fmt.Errorf("failed to create identity v3 client: %w", err)
	} else {
		slog.Info("identity v3 client initialised")
		service.Microversion = IdentityV3MicroVersion
		client.IdentityV3 = &IdentityV3{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}

	// initialize Glance client
	if service, err := osp.NewImageServiceV2(provider, gophercloud.EndpointOpts{Region: DefaultRegion}); err != nil {
		slog.Error("error creating image service v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create image service v2 client: %w", err)
	} else {
		slog.Info("image service v2 client initialised")
		service.Microversion = ImageV2MicroVersion
		client.ImageV2 = &ImageV2{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}

	// initialize Neutron client
	if service, err := osp.NewNetworkV2(provider, gophercloud.EndpointOpts{Region: DefaultRegion}); err != nil {
		slog.Error("error creating network v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create network v2 client: %w", err)
	} else {
		slog.Info("network v2 client initialised")
		service.Microversion = NetworkingV2MicroVersion
		client.NetworkingV2 = &NetworkingV2{
			Service{
				provider: provider,
				client:   service,
			},
		}
	}
	slog.Info("client ready")
	return client, nil
}
