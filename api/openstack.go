package api

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/pauseunpause"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type OpenStackClient struct {
	client *gophercloud.ServiceClient
	mu     sync.RWMutex
	cache  map[string]string // userID -> serverID
}

func NewOpenStackClient() (*OpenStackClient, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth options: %w", err)
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: "RegionOne", // TODO: should probably be configurable
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return &OpenStackClient{
		client: client,
		cache:  make(map[string]string),
	}, nil
}

func (c *OpenStackClient) getServerID(ctx context.Context, userID string) (string, error) {
	c.mu.RLock()
	id, ok := c.cache[userID]
	c.mu.RUnlock()
	if ok {
		return id, nil
	}

	slog.Info("looking up server for user", "userID", userID)
	listOpts := servers.ListOpts{
		Tags: fmt.Sprintf("slumber-user-id=%s", userID),
	}

	allPages, err := servers.List(c.client, listOpts).AllPages()
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract servers: %w", err)
	}

	if len(allServers) == 0 {
		return "", fmt.Errorf("no server found for user %s", userID)
	}

	serverID := allServers[0].ID
	c.mu.Lock()
	c.cache[userID] = serverID
	c.mu.Unlock()

	return serverID, nil
}

func (c *OpenStackClient) Start(ctx context.Context, userID string) error {
	id, err := c.getServerID(ctx, userID)
	if err != nil {
		return err
	}
	res := startstop.Start(c.client, id)
	return res.ExtractErr()
}

func (c *OpenStackClient) Stop(ctx context.Context, userID string) error {
	id, err := c.getServerID(ctx, userID)
	if err != nil {
		return err
	}
	res := startstop.Stop(c.client, id)
	return res.ExtractErr()
}

func (c *OpenStackClient) Pause(ctx context.Context, userID string) error {
	id, err := c.getServerID(ctx, userID)
	if err != nil {
		return err
	}
	res := pauseunpause.Pause(c.client, id)
	return res.ExtractErr()
}

func (c *OpenStackClient) Unpause(ctx context.Context, userID string) error {
	id, err := c.getServerID(ctx, userID)
	if err != nil {
		return err
	}
	res := pauseunpause.Unpause(c.client, id)
	return res.ExtractErr()
}

func (c *OpenStackClient) Status(ctx context.Context, userID string) (string, error) {
	id, err := c.getServerID(ctx, userID)
	if err != nil {
		return "", err
	}
	server, err := servers.Get(c.client, id).Extract()
	if err != nil {
		return "", err
	}
	return server.Status, nil
}
