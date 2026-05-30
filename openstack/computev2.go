package openstack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gophercloud/gophercloud"
	osp "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

const ComputeV2MicroVersion = "2.79"

type ComputeV2 struct {
	Service
	// mu    sync.RWMutex
	// cache map[string][]string // userID -> serverID(s)
}

// initialise the ComputeV2 client
func newComputeV2(provider *gophercloud.ProviderClient) (*ComputeV2, error) {
	slog.Debug("initialising compute v2 client")
	if service, err := osp.NewComputeV2(provider, gophercloud.EndpointOpts{}); err != nil {
		slog.Error("error creating compute v2 API client", "error", err)
		return nil, fmt.Errorf("failed to create compute v2 client: %w", err)
	} else {
		slog.Info("compute v2 client initialised")
		service.Microversion = ComputeV2MicroVersion
		return &ComputeV2{
			Service{
				provider: provider,
				client:   service,
			},
		}, nil
	}
}

// c.mu.RLock()
// ids, ok := c.cache[userID]
// c.mu.RUnlock()
// if ok {
// 	return ids, nil
// }

type ComputeV2ListOption func(*servers.ListOpts)

func WithName(pattern string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		o.Name = pattern
	}
}

func WithStatus(status string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		// TODO: check status value against enum
		o.Status = status
	}
}

func WithUserID(userID string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		o.UserID = userID
	}
}

func WithImage(imageURL string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		o.Image = imageURL
	}
}

func WithFlavor(flavorURL string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		o.Flavor = flavorURL
	}
}

func WithTags(tags ...string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		if len(o.Tags) == 0 {
			o.Tags = strings.Join(tags, ",")
		} else {
			o.Tags = o.Tags + "," + strings.Join(tags, ",")
		}
	}
}

func WithAnyTag(tags ...string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		if len(o.TagsAny) == 0 {
			o.TagsAny = strings.Join(tags, ",")
		} else {
			o.TagsAny = o.TagsAny + "," + strings.Join(tags, ",")
		}
	}
}

func (c *ComputeV2) List(ctx context.Context, options ...ComputeV2ListOption) ([]servers.Server, error) {
	slog.Info("looking up servers")
	listOpts := servers.ListOpts{}
	for _, option := range options {
		option(&listOpts)
	}

	allPages, err := servers.List(c.client, listOpts).AllPages()
	if err != nil {
		slog.Error("error listing servers", "error", err)
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract servers: %w", err)
	}

	if len(allServers) == 0 {
		return nil, nil
	}
	return allServers, nil

	// serverID := allServers[0].ID
	// c.mu.Lock()
	// c.cache[userID] = serverID
	// c.mu.Unlock()

	// return serverID, nil
}

// TODO: migrate these
// func (c *Client) Start(ctx context.Context, userID string) error {
// 	id, err := c.getServerID(ctx, userID)
// 	if err != nil {
// 		return err
// 	}
// 	res := startstop.Start(c.client, id)
// 	return res.ExtractErr()
// }

// func (c *Client) Stop(ctx context.Context, userID string) error {
// 	id, err := c.getServerID(ctx, userID)
// 	if err != nil {
// 		return err
// 	}
// 	res := startstop.Stop(c.client, id)
// 	return res.ExtractErr()
// }

// func (c *Client) Pause(ctx context.Context, userID string) error {
// 	id, err := c.getServerID(ctx, userID)
// 	if err != nil {
// 		return err
// 	}
// 	res := pauseunpause.Pause(c.client, id)
// 	return res.ExtractErr()
// }

// func (c *Client) Unpause(ctx context.Context, userID string) error {
// 	id, err := c.getServerID(ctx, userID)
// 	if err != nil {
// 		return err
// 	}
// 	res := pauseunpause.Unpause(c.client, id)
// 	return res.ExtractErr()
// }

// func (c *Client) Status(ctx context.Context, userID string) (string, error) {
// 	id, err := c.getServerID(ctx, userID)
// 	if err != nil {
// 		return "", err
// 	}
// 	server, err := servers.Get(c.client, id).Extract()
// 	if err != nil {
// 		return "", err
// 	}
// 	return server.Status, nil
// }
