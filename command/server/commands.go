package server

import (
	"github.com/dihedron/devws/command/server/list"
	"github.com/dihedron/devws/command/server/tag"
)

type Server struct {
	// Login is the command that checks logins to an LDAP server.
	List list.List `command:"list" alias:"ls" description:"List the virtual machines."`
	// // API is the command that starts the API server.
	Tag tag.Tag `command:"tag" alias:"t" description:"Subcommands related to tags."`
}
