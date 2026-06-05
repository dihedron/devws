package tag

import "github.com/dihedron/devws/command/server/tag/list"

type Tag struct {
	// List is the command that lists tags on a virtual machine.
	List list.List `command:"list" alias:"ls" description:"List the tags on a virtual machine."`
}
