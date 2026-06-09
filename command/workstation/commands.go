package workstation

import (
	"github.com/dihedron/devws/command/workstation/list"
	"github.com/dihedron/devws/command/workstation/pause"
	"github.com/dihedron/devws/command/workstation/reboot"
	"github.com/dihedron/devws/command/workstation/resume"
	"github.com/dihedron/devws/command/workstation/shelve"
	"github.com/dihedron/devws/command/workstation/start"
	"github.com/dihedron/devws/command/workstation/stop"
	"github.com/dihedron/devws/command/workstation/suspend"
	"github.com/dihedron/devws/command/workstation/tag"
	"github.com/dihedron/devws/command/workstation/unpause"
	"github.com/dihedron/devws/command/workstation/unshelve"
	"github.com/dihedron/devws/command/workstation/view"
)

type Workstation struct {
	// List is the command that lists all available workstations.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	List list.List `command:"list" alias:"ls" alias:"l" description:"List all workstations."`
	// Show is the command that displays a workstation's details.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	View view.View `command:"view" alias:"v" alias:"show" description:"Show details about a workstation."`
	// Start is the command that starts a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Start start.Start `command:"start" alias:"s" description:"Start a workstation."`
	// Stop is the command that stops a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Stop stop.Stop `command:"stop" alias:"t" description:"Stop a workstation."`
	// Pause is the command that pauses a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Pause pause.Pause `command:"pause" alias:"p" description:"Pause a workstation."`
	// Unpause is the command that unpauses a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Unpause unpause.Unpause `command:"unpause" alias:"u" description:"Unpause a workstation."`
	// Suspend is the command that suspends a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Suspend suspend.Suspend `command:"suspend" alias:"n" description:"Suspend a workstation."`
	// Suspend is the command that suspends a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Resume resume.Resume `command:"resume" alias:"m" description:"Resume a workstation."`
	// Reboot is the command that reboots a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Reboot reboot.Reboot `command:"reboot" alias:"r" description:"Reboot a workstation."`
	// Shelve is the command that shelves a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Shelve shelve.Shelve `command:"shelve" alias:"e" description:"Shelve a workstation."`
	// Unshelve is the command that unshelves a workstation.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Unshelve unshelve.Unshelve `command:"unshelve" alias:"x" description:"Unshelve a workstation."`
	// Tag is the subcommand that manipulates a workstation's tags.
	Tag tag.Tag `command:"tag" alias:"t" description:"Subcommands related to workstation tags."`
}
