package openstack

import (
	"log/slog"
	"os"
)

var (
	DefaultCloud  = "openstack"
	DefaultRegion = "RegionOne"
)

// place the package initialisation logic here
func init() {
	slog.Debug("initialising package openstack")
	if cloud, ok := os.LookupEnv("OS_CLOUD"); ok {
		slog.Info("using cloud from OS_CLOUD instead of default", "cloud", cloud, "default", DefaultCloud)
		DefaultCloud = cloud
	}

	if region, ok := os.LookupEnv("OS_REGION_NAME"); ok {
		slog.Info("using region from OS_REGION_NAME instead of default", "region", region, "default", DefaultRegion)
		DefaultRegion = region
	}
}
