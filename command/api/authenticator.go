package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Authenticator interface {
	Authenticate(c *gin.Context, username, password string) bool
}

type StaticAuthenticator struct {
	Accounts map[string]string
}

func (a *StaticAuthenticator) Authenticate(c *gin.Context, username, password string) bool {
	if pass, exists := a.Accounts[username]; exists {
		slog.Debug("user successfuilly authenticated", "username", username, "password", password)
		return pass == password
	}
	slog.Debug("error authenticating user", "username", username)
	return false
}

type LDAPAuthenticator struct {
}

func (a *LDAPAuthenticator) Authenticate(c *gin.Context, username, password string) bool {
	return false
}
