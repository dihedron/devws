package openstack

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/gophercloud/gophercloud"
	osp "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

type Service struct {
	provider *gophercloud.ProviderClient
	client   *gophercloud.ServiceClient
}

type Client struct {
	client *gophercloud.ProviderClient
	// BareMetalV1    *BareMetalV1
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
		slog.Error("failed to get auth options from environment", "error", err)
		return nil, fmt.Errorf("failed to get auth options from environment: %w", err)
	}
	return newClient(opts)
}

func newClient(opts gophercloud.AuthOptions) (*Client, error) {
	opts.AllowReauth = true
	provider, err := osp.AuthenticatedClient(opts)
	if err != nil {
		slog.Error("error creating authenticated client", "error", err)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	//slog.debug("checking token scope", "scope", fmt.Sprintf("%+v", provider.AuthenticatedProjectID))

	// debug
	ar := provider.GetAuthResult()
	data, _ := yaml.Marshal(ar)
	os.WriteFile("catalog.yml", data, 0644)

	client := &Client{}

	// // initialize Bare Metal client
	// if client.BareMetalV1, err = newBareMetalV1(provider); err != nil {
	// 	return nil, err
	// }

	// initialise Cinder client
	if client.BlockStorageV3, err = newBlockStorageV3(provider); err != nil {
		return nil, err
	}

	// initialize Nova client
	if client.ComputeV2, err = newComputeV2(provider); err != nil {
		return nil, err
	}

	// initialize Keystone client
	if client.IdentityV3, err = newIdentityV3(provider); err != nil {
		return nil, err
	}

	// initialize Glance client
	if client.ImageV2, err = newImageV2(provider); err != nil {
		return nil, err
	}

	// initialize Neutron client
	if client.NetworkingV2, err = newNetworkingV2(provider); err != nil {
		return nil, err
	}

	slog.Info("client ready")
	return client, nil
}
