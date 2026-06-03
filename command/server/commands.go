package server

import "github.com/dihedron/devws/command/server/list"

type Server struct {
	// Login is the command that checks logins to an LDAP server.
	List list.List `command:"list" alias:"ls" description:"List the virtual machines."`
	// // API is the command that starts the API server.
	// Server server.Server `command:"server" alias:"a" description:"Start the API server." `
}
