package base

// Command is the base command.
type Command struct {
	// Name is the optional name or ID of the virtual machine.
	Cloud string `short:"c" long:"cloud" description:"The cloud profile to use." default:"openstack" optional:"yes" env:"OS_CLOUD"`
	// Format is the output format of the result.
	Format string `short:"f" long:"format" description:"The format of the result." optional:"true" choice:"json" choice:"yaml" choice:"text" default:"yaml"`
}
