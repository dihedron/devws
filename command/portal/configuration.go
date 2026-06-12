package portal

import (
	"log/slog"

	"github.com/dihedron/defaults"
	"github.com/dihedron/rawdata"
)

type Configuration struct {
	Address string `json:"address,omitempty" yaml:"address,omitempty" default:":3000"`
	LDAP    struct {
		Server   string `json:"server,omitempty" yaml:"server,omitempty" default:"ldaps://ldap.example.com:636"`
		Account  string `json:"account,omitempty" yaml:"account,omitempty" default:"CN=admin,OU=ITDept,DC=example,DC=com"`
		Password string `json:"password,omitempty" yaml:"password,omitempty" default:"IF0rg0tMyP4$$w0rd?|"`
		BaseDN   string `json:"basedn,omitempty" yaml:"basedn,omitempty" default:"OU=ITDept,DC=example,DC=com"`
	} `json:",inline" yaml:",inline"`
}

// UnmarshalFlag unmarshals the data on the command line into
// a Configuration object.
func (c *Configuration) UnmarshalFlag(value string) error {
	if err := rawdata.UnmarshalInto(value, c); err != nil {
		return err
	}
	if err := defaults.Set(c); err != nil {
		return err
	}
	slog.Debug("configuration loaded", "configuration", *c)
	return nil
}
