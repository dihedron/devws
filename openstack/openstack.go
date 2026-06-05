package openstack

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/config"
	"github.com/gophercloud/gophercloud/v2/openstack/config/clouds"
)

var DefaultCloud = "openstack"

type Service struct {
	provider *gophercloud.ProviderClient
	client   *gophercloud.ServiceClient
}

func (s *Service) Client() *gophercloud.ServiceClient {
	return s.client
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

func NewClient(ctx context.Context) (*Client, error) {
	// fetch coordinates from a `cloud.yaml` in the current directory, or
	// in the well-known config directories (different for each operating
	// system).
	authOpts, _, tlsConfig, err := clouds.Parse()
	if err != nil {
		slog.Error("failed to get options from cloud.yaml", "error", err)
		return nil, fmt.Errorf("failed to get auth options from environment: %w", err)
	}

	// call Keystone to get an authentication token and construct a ProviderClient
	authOpts.AllowReauth = true
	provider, err := config.NewProviderClient(ctx, authOpts, config.WithTLSConfig(tlsConfig))
	if err != nil {
		slog.Error("error creating authenticated client", "error", err)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return newClient(ctx, provider)
}

func NewClientFromEnv(ctx context.Context) (*Client, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		slog.Error("failed to get auth options from environment", "error", err)
		return nil, fmt.Errorf("failed to get auth options from environment: %w", err)
	}
	// call Keystone to get an authentication token and construct a ProviderClient
	provider, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		slog.Error("error creating authenticated client", "error", err)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return newClient(ctx, provider)
}

func newClient(ctx context.Context, provider *gophercloud.ProviderClient /*, endpointOpts gophercloud.EndpointOpts*/) (*Client, error) {

	var err error
	//slog.debug("checking token scope", "scope", fmt.Sprintf("%+v", provider.AuthenticatedProjectID))

	// debug
	// ar := provider.GetAuthResult()
	// data, _ := yaml.Marshal(ar)
	// os.WriteFile("catalog.yml", data, 0644)

	client := &Client{
		client: provider,
	}

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
