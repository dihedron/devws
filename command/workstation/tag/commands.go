package tag

import (
	"github.com/dihedron/devws/command/workstation/tag/add"
	"github.com/dihedron/devws/command/workstation/tag/check"
	"github.com/dihedron/devws/command/workstation/tag/clear"
	"github.com/dihedron/devws/command/workstation/tag/delete"
	"github.com/dihedron/devws/command/workstation/tag/list"
	"github.com/dihedron/devws/command/workstation/tag/replace"
)

type Tag struct {
	// List is the command that lists tags on a virtual machine.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	List list.List `command:"list" alias:"ls" alias:"l" description:"List the tags on a virtual machine."`
	// List is the command that lists tags on a virtual machine.
	Check check.Check `command:"check" alias:"c" description:"Check the existence of a tags on a virtual machine."`
	// Add is the command that adds tags to a virtual machine.
	Add add.Add `command:"add" alias:"a" description:"Add one or more tags to a virtual machine."`
	// Delete is the command that removes tags from a virtual machine.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Delete delete.Delete `command:"delete" alias:"del" alias:"remove" alias:"rm" alias:"d" description:"Remove one or more tags from a virtual machine."`
	// Clear is the command that clears all tags from a virtual machine.
	Clear clear.Clear `command:"clear" alias:"clr" alias:"x" description:"Clear all tags from a virtual machine."`
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	// Replace is the command that replaces all tags on a virtual machine.
	Replace replace.Replace `command:"replace" alias:"r" description:"Replace all tags on a virtual machine."`
}
