package base

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
)

// Command is the base command.
type Command struct {
	// Name is the optional name or ID of the virtual machine.
	Cloud *string `short:"C" long:"cloud" description:"The cloud profile to use." default:"openstack" optional:"yes" env:"OS_CLOUD"`
	// Format is the output format of the result.
	Format string `short:"F" long:"format" description:"The format of the result." optional:"true" choice:"json" choice:"yaml" choice:"text" default:"yaml"`
}

// Init initialises the OS_CLOUD environment variable.
func (cmd *Command) Init() {
	if _, ok := os.LookupEnv("OS_CLOUD"); !ok && cmd.Cloud != nil {
		slog.Debug("setting OS_CLOUD environment variable", "value", *cmd.Cloud)
		os.Setenv("OS_CLOUD", *cmd.Cloud)
	}
}

// Output writes the command's output in the correct format to output.
func (cmd *Command) Output(result any) {
	switch cmd.Format {
	case "yaml":
		data, _ := yaml.Marshal(result)
		fmt.Printf("%s", data)
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("%s\n", string(data))
	case "text":
		fmt.Printf("%+v\n", result)
	}
}
