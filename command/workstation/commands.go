package workstation

import (
	"github.com/dihedron/devws/command/workstation/list"
	"github.com/dihedron/devws/command/workstation/show"
	"github.com/dihedron/devws/command/workstation/tag"
)

type Workstation struct {
	// List is the command that lists all available workstations.
	List list.List `command:"list" alias:"ls" alias:"l" description:"List all workstations."`
	// Show is the command that displays a workstation's details.
	Show show.Show `command:"show" alias:"s" description:"Show details about a workstation."`
	// Tag is the subcommand that manipulates a workstation's tags.
	Tag tag.Tag `command:"tag" alias:"t" description:"Subcommands related to workstation tags."`
}
