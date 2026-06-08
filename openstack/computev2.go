package openstack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/tags"
)

const ComputeV2MicroVersion = "2.79"

type ComputeV2 struct {
	Service
}

// initialise the ComputeV2 client
func newComputeV2(provider *gophercloud.ProviderClient) (*ComputeV2, error) {
	slog.Debug("initialising compute v2 client")
	if service, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{}); err != nil {
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

func WithFlavour(flavourURL string) ComputeV2ListOption {
	return func(o *servers.ListOpts) {
		o.Flavor = flavourURL
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

// Instance is an internal type used to unmarshal more data from the API
// response than would usually be possible through the ordinary gophercloud
// struct. OpenStack API microversions enable more response data that is not
// taken into account by the gophercloud library, which unmarshals only what
// is available at the base level for each API version, for backward compatibility.
// This is also why there is an ExtractInto function that allows you to pass in
// an arbitrary struct to marshal the response data into.
type Workstation struct {
	ID           string `json:"id,omitempty" yaml:"id,omitempty"`
	TenantID     string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty"`
	UserID       string `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	CreatedAt    *Time  `json:"created,omitempty" yaml:"created,omitempty"`
	LaunchedAt   *Time  `json:"OS-SRV-USG:launched_at,omitempty" yaml:"OS-SRV-USG:launched_at,omitempty"`
	UpdatedAt    *Time  `json:"updated,omitempty" yaml:"updated,omitempty"`
	TerminatedAt *Time  `json:"OS-SRV-USG:terminated_at,omitempty" yaml:"OS-SRV-USG:terminated_at,omitempty"`
	HostID       string `json:"hostid,omitempty" yaml:"hostid,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
	Progress     int    `json:"progress,omitempty" yaml:"progress,omitempty"`
	AccessIPv4   string `json:"accessIPv4,omitempty" yaml:"accessIPv4,omitempty"`
	AccessIPv6   string `json:"accessIPv6,omitempty" yaml:"accessIPv6,omitempty"`
	Image        any    `json:"image,omitempty" yaml:"image,omitempty"`
	Flavor       struct {
		Name          string          `json:"original_name,omitempty" yaml:"original_name,omitempty"`
		Disk          int             `json:"disk,omitempty" yaml:"disk,omitempty"`
		RAM           int             `json:"ram,omitempty" yaml:"ram,omitempty"`
		Swap          int             `json:"-" yaml:"-"`
		VCPUs         int             `json:"vcpus,omitempty" yaml:"vcpus,omitempty"`
		Ephemeral     int             `json:"OS-FLV-EXT-DATA:ephemeral,omitempty" yaml:"OS-FLV-EXT-DATA:ephemeral,omitempty"`
		ExtraSpecsRaw json.RawMessage `json:"extra_specs,omitempty" yaml:"extra_specs,omitempty"`
		ExtraSpecsObj *struct {
			CPUCores        string `json:"hw:cpu_cores,omitempty" yaml:"hw:cpu_cores,omitempty"`
			CPUSockets      string `json:"hw:cpu_sockets,omitempty" yaml:"hw:cpu_sockets,omitempty"`
			RNGAllowed      string `json:"hw_rng:allowed,omitempty" yaml:"hw_rng:allowed,omitempty"`
			WatchdogAction  string `json:"hw:watchdog_action,omitempty" yaml:"hw:watchdog_action,omitempty"`
			VGPUs           string `json:"resources:VGPU,omitempty" yaml:"resources:VGPU,omitempty"`
			TraitCustomVGPU string `json:"trait:CUSTOM_VGPU,omitempty" yaml:"trait:CUSTOM_VGPU,omitempty"`
		} `json:"-" yaml:"-"`
		ExtraSpecsMap *map[string]string `json:"-" yaml:"-"`
	} `json:"flavor,omitempty" yaml:"flavor,omitempty"`
	Addresses map[string][]struct {
		Network    *string `json:"-" yaml:"-"`
		MACAddress string  `json:"OS-EXT-IPS-MAC:mac_addr,omitempty" yaml:"OS-EXT-IPS-MAC:mac_addr,omitempty"`
		IPType     string  `json:"OS-EXT-IPS:type,omitempty" yaml:"OS-EXT-IPS:type,omitempty"`
		IPAddress  string  `json:"addr,omitempty" yaml:"addr,omitempty"`
		IPVersion  int     `json:"version,omitempty" yaml:"version,omitempty"`
	} `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Links    []struct {
		Href string `json:"href,omitempty" yaml:"href,omitempty"`
		Rel  string `json:"rel,omitempty" yaml:"rel,omitempty"`
	} `json:"links,omitempty" yaml:"links,omitempty"`
	KeyName        string `json:"key_name,omitempty" yaml:"key_name,omitempty"`
	AdminPass      string `json:"adminPass,omitempty" yaml:"adminPass,omitempty"`
	SecurityGroups []struct {
		Name string `json:"name,omitempty" yaml:"name,omitempty"`
	} `json:"security_groups,omitempty" yaml:"security_groups,omitempty"`
	AttachedVolumes    []servers.AttachedVolume `json:"os-extended-volumes:volumes_attached,omitempty" yaml:"os-extended-volumes:volumes_attached,omitempty"`
	Tags               *[]string                `json:"tags,omitempty" yaml:"tags,omitempty"`
	ServerGroups       *[]string                `json:"server_groups,omitempty" yaml:"server_groups,omitempty"`
	DiskConfig         string                   `json:"OS-DCF:diskConfig,omitempty" yaml:"OS-DCF:diskConfig,omitempty"`
	AvailabilityZone   string                   `json:"OS-EXT-AZ:availability_zone,omitempty" yaml:"OS-EXT-AZ:availability_zone,omitempty"`
	Host               string                   `json:"OS-EXT-SRV-ATTR:host,omitempty" yaml:"OS-EXT-SRV-ATTR:host,omitempty"`
	HostName           string                   `json:"OS-EXT-SRV-ATTR:hostname,omitempty" yaml:"OS-EXT-SRV-ATTR:hostname,omitempty"`
	HypervisorHostname string                   `json:"OS-EXT-SRV-ATTR:hypervisor_hostname,omitempty" yaml:"OS-EXT-SRV-ATTR:hypervisor_hostname,omitempty"`
	InstanceName       string                   `json:"OS-EXT-SRV-ATTR:instance_name,omitempty" yaml:"OS-EXT-SRV-ATTR:instance_name,omitempty"`
	KernelID           string                   `json:"OS-EXT-SRV-ATTR:kernel_id,omitempty" yaml:"OS-EXT-SRV-ATTR:kernel_id,omitempty"`
	LaunchIndex        int                      `json:"OS-EXT-SRV-ATTR:launch_index,omitempty" yaml:"OS-EXT-SRV-ATTR:launch_index,omitempty"`
	RAMDiskID          string                   `json:"OS-EXT-SRV-ATTR:ramdisk_id,omitempty" yaml:"OS-EXT-SRV-ATTR:ramdisk_id,omitempty"`
	ReservationID      string                   `json:"OS-EXT-SRV-ATTR:reservation_id,omitempty" yaml:"OS-EXT-SRV-ATTR:reservation_id,omitempty"`
	RootDeviceName     string                   `json:"OS-EXT-SRV-ATTR:root_device_name,omitempty" yaml:"OS-EXT-SRV-ATTR:root_device_name,omitempty"`
	UserData           string                   `json:"OS-EXT-SRV-ATTR:user_data,omitempty" yaml:"OS-EXT-SRV-ATTR:user_data,omitempty"`
	PowerState         int                      `json:"OS-EXT-STS:power_state,omitempty" yaml:"OS-EXT-STS:power_state,omitempty"`
	VMState            string                   `json:"OS-EXT-STS:vm_state,omitempty" yaml:"OS-EXT-STS:vm_state,omitempty"`
	ConfigDrive        string                   `json:"config_drive,omitempty" yaml:"config_drive,omitempty"`
	Description        string                   `json:"description,omitempty" yaml:"description,omitempty"`
	// NO!!! Fault              servers.Fault            `json:"fault,omitempty" yaml:"fault,omitempty"`
	// NO!!! TaskState          interface{}              `json:"OS-EXT-STS:task_state,omitempty" yaml:"OS-EXT-STS:task_state,omitempty"`
}

var (
	uuid        = regexp.MustCompile(`(?i)^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	ErrNotFound = errors.New("item not found")
)

func (c *ComputeV2) GetFirstID(ctx context.Context, value string) (string, error) {
	ids, err := c.GetAllIDs(ctx, value)
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}
	return "", ErrNotFound
}

func (c *ComputeV2) GetAllIDs(ctx context.Context, value string) ([]string, error) {
	if uuid.MatchString(value) {
		slog.Debug("value is already an ID", "value", value)
		return []string{value}, nil
	}
	slog.Debug("need to resolve name to ID", "value", value)
	wks, err := c.List(ctx, WithName(strings.ReplaceAll(value, ".", "\\.")))
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, w := range wks {
		ids = append(ids, w.ID)
	}
	return ids, nil
}

// List lists all workstations, possibly filtering using a set of criteria.
func (c *ComputeV2) List(ctx context.Context, options ...ComputeV2ListOption) ([]Workstation, error) {
	slog.Info("looking up servers")
	listOpts := servers.ListOpts{}
	for _, option := range options {
		option(&listOpts)
	}

	allPages, err := servers.List(c.client, listOpts).AllPages(ctx)
	if err != nil {
		slog.Error("error listing servers", "error", err)
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	var allServers []Workstation
	err = servers.ExtractServersInto(allPages, &allServers)
	if err != nil {
		return nil, fmt.Errorf("failed to extract servers: %w", err)
	}

	if len(allServers) == 0 {
		return nil, nil
	}
	return allServers, nil
}

// View retrieves details about a specific server, by id.
func (c *ComputeV2) View(ctx context.Context, serverId string) (*Workstation, error) {
	slog.Info("looking up single server", "serverId", serverId)
	wks := &Workstation{}
	err := servers.Get(ctx, c.client, serverId).ExtractInto(wks)
	return wks, err
}

func (c *ComputeV2) Reboot(ctx context.Context, serverId string, method servers.RebootMethod) error {
	return servers.Reboot(ctx, c.client, serverId, servers.RebootOpts{Type: method}).ExtractErr()
}

func (c *ComputeV2) SoftReboot(ctx context.Context, serverId string) error {
	return c.Reboot(ctx, serverId, servers.SoftReboot)
}

func (c *ComputeV2) HardReboot(ctx context.Context, serverId string) error {
	return c.Reboot(ctx, serverId, servers.HardReboot)
}

// Start starts a specific server, by id.
func (c *ComputeV2) Start(ctx context.Context, serverId string) error {
	slog.Info("starting a server", "serverId", serverId)
	return servers.Start(ctx, c.client, serverId).ExtractErr()
}

// Stops starts a specific server, by id.
func (c *ComputeV2) Stop(ctx context.Context, serverId string) error {
	slog.Info("stopping a server", "serverId", serverId)
	return servers.Stop(ctx, c.client, serverId).ExtractErr()
}

// Suspend suspends a specific server, by id.
func (c *ComputeV2) Suspend(ctx context.Context, serverId string) error {
	slog.Info("suspending a server", "serverId", serverId)
	return servers.Suspend(ctx, c.client, serverId).ExtractErr()
}

// Resume resumes a specific server, by id.
func (c *ComputeV2) Resume(ctx context.Context, serverId string) error {
	slog.Info("resuming a server", "serverId", serverId)
	return servers.Resume(ctx, c.client, serverId).ExtractErr()
}

// Pause pauses a specific server, by id.
func (c *ComputeV2) Pause(ctx context.Context, serverId string) error {
	slog.Info("pausing a server", "serverId", serverId)
	return servers.Pause(ctx, c.client, serverId).ExtractErr()
}

// Unpause unpauses a specific server, by id.
func (c *ComputeV2) Unpause(ctx context.Context, serverId string) error {
	slog.Info("unpausing a server", "serverId", serverId)
	return servers.Unpause(ctx, c.client, serverId).ExtractErr()
}

// Shelve is the operation responsible for shelving a server.
func (c *ComputeV2) Shelve(ctx context.Context, serverId string) error {
	slog.Info("shelving a server", "serverId", serverId)
	return servers.Shelve(ctx, c.client, serverId).ExtractErr()
}

// ShelveOffload is the operation responsible for Shelve-Offload a server.
func (c *ComputeV2) ShelveOffload(ctx context.Context, serverId string) error {
	slog.Info("shelve-offloading a server", "serverId", serverId)
	return servers.ShelveOffload(ctx, c.client, serverId).ExtractErr()
}

func (c *ComputeV2) Unshelve(ctx context.Context, serverId string, availabilityZone string) error {
	slog.Info("unshelving a server", "serverId", serverId, "availabilityZone", availabilityZone)
	return servers.Unshelve(ctx, c.client, serverId, servers.UnshelveOpts{AvailabilityZone: availabilityZone}).ExtractErr()
}

func (c *ComputeV2) AddTags(ctx context.Context, serverId string, values ...string) error {
	for _, tag := range values {
		err := tags.Add(ctx, c.client, serverId, tag).ExtractErr()
		if err != nil {
			slog.Error("error adding server tag", "error", err, "serverId", serverId, "tag", tag)
			return err
		}
	}
	return nil
}

func (c *ComputeV2) DeleteTags(ctx context.Context, serverId string, values ...string) error {
	for _, tag := range values {
		err := tags.Delete(ctx, c.client, serverId, tag).ExtractErr()
		if err != nil {
			slog.Error("error deleting server tag", "error", err, "serverId", serverId, "tag", tag)
			return err
		}
	}
	return nil
}

func (c *ComputeV2) CheckTag(ctx context.Context, serverId string, tag string) (bool, error) {
	exists, err := tags.Check(ctx, c.client, serverId, tag).Extract()
	if err != nil {
		slog.Error("error checking server tag", "error", err, "serverId", serverId, "tag", tag)
		return false, err
	}
	return exists, nil
}

func (c *ComputeV2) ReplaceTags(ctx context.Context, serverId string, values ...string) ([]string, error) {
	result, err := tags.ReplaceAll(ctx, c.client, serverId, tags.ReplaceAllOpts{Tags: values}).Extract()
	if err != nil {
		slog.Error("error replacing all server tags", "error", err, "serverId", serverId, "tags", values)
		return nil, err
	}
	return result, nil
}

func (c *ComputeV2) ClearTags(ctx context.Context, serverId string) error {
	err := tags.DeleteAll(ctx, c.client, serverId).ExtractErr()
	if err != nil {
		slog.Error("error clearing server tags", "error", err, "serverId", serverId)
		return err
	}
	return nil
}

func (c *ComputeV2) ListTags(ctx context.Context, serverId string) ([]string, error) {
	values, err := tags.List(ctx, c.client, serverId).Extract()
	if err != nil {
		slog.Error("error listing server tags", "error", err, "serverId", serverId)
		return nil, err
	}
	return values, nil
}
