package server

import (
	"github.com/dihedron/devws/command/server/list"
	"github.com/dihedron/devws/command/server/show"
	"github.com/dihedron/devws/command/server/tag"
)

type Server struct {
	// List is the command that lists all available servers.
	List list.List `command:"list" alias:"ls" description:"List the virtual machines."`
	// Show is the command that displays a virtual machine's details.
	Show show.Show `command:"show" alias:"s" description:"Show a virtual machine details."`
	// // API is the command that starts the API server.
	Tag tag.Tag `command:"tag" alias:"t" description:"Subcommands related to tags."`
}
