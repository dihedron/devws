package command

import (
	"github.com/dihedron/devws/command/login"
	"github.com/dihedron/devws/command/portal"
	"github.com/dihedron/devws/command/power"
	"github.com/dihedron/devws/command/version"
	"github.com/dihedron/devws/command/workstation"
)

// Commands is the set of root command groups.
type Commands struct {
	// Login is the command that checks logins to an LDAP server.
	Login login.Login `command:"login" alias:"l" description:"Log in to an LDAP server." hidden:"true"`
	// Workstation is a set of commands to manipulate workstations on OpenStack.
	//lint:ignore SA5008 multiple struct alias tags are allowed and useful
	Workstation workstation.Workstation `command:"workstation" alias:"ws" alias:"w" description:"Manipulate workstations in OpenStack."`
	// API is the command that starts the API server.
	Portal portal.Portal `command:"portal" alias:"p" description:"Start the API server." `
	// Shutdown is the command that shuts down the machine.
	Shutdown power.Shutdown `command:"shutdown" alias:"s" description:"Shut down the machine."`
	// Hibernate is the command that hibernates the machine.
	Hibernate power.Hibernate `command:"hibernate" alias:"h" description:"Hibernate the machine."`
	// Version prints overlay version information and exits.
	Version version.Version `command:"version" alias:"v" description:"Show the command version and exit."`
}
