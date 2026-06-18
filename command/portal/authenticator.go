package portal

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type AuthenticateOptions struct {
	username            string
	password            string
	skipPwdVerification bool
}

// Option is a function that configures a Server
type Option func(*AuthenticateOptions)

func WithCredentials(username, password string) func(*AuthenticateOptions) {
	return func(a *AuthenticateOptions) {
		a.username = username
		a.password = password
	}
}

func WithSkipPwdVerification(skipPwdVerification bool) func(*AuthenticateOptions) {
	return func(a *AuthenticateOptions) {
		a.skipPwdVerification = skipPwdVerification
	}
}

type Authenticator interface {
	// Authenticate will return true if the user could be successfully
	// authenticated; false (with no error) if the user's credentials
	// are invalid; false (with an error) if the authenticator encountered
	// and internal processing error.
	Authenticate(username, password string) (bool, error)
	// Authenticate a user an active session
	//
	// Variadics options:
	// WithCredentials(username, password string) - username e password
	// WithSkipPwdVerification(skip bool) - se true viene saltata la validazione della password
	Authenticate2(opts ...Option) (*UserSession, error)
	// Close can be used to perform cleanup operations.
	Close() error
}

// StaticAuthenticator authenticates users against an in memory, static map.
type StaticAuthenticator struct {
	accounts map[string]string
}

func NewStaticAuthenticator(options ...func(*StaticAuthenticator)) *StaticAuthenticator {
	auth := &StaticAuthenticator{
		accounts: map[string]string{},
	}
	for _, option := range options {
		option(auth)
	}
	return auth
}

func WithUser(username, password string) func(*StaticAuthenticator) {
	return func(a *StaticAuthenticator) {
		a.accounts[username] = password
	}
}

func (a *StaticAuthenticator) Authenticate(username, password string) (bool, error) {
	if pass, exists := a.accounts[username]; exists {
		slog.Debug("user successfully authenticated", "username", username, "password", password)
		return pass == password, nil
	}
	slog.Debug("error authenticating user", "username", username)
	return false, nil
}

func (a *StaticAuthenticator) Authenticate2(opts ...Option) (*UserSession, error) {
	authOpt := &AuthenticateOptions{
		username:            "",
		password:            "",
		skipPwdVerification: false,
	}
	for _, opt := range opts {
		opt(authOpt)
	}
	if pass, exists := a.accounts[authOpt.username]; exists {
		if pass == authOpt.password {
			slog.Debug("user successfully authenticated", "username", authOpt.username, "password", authOpt.password)
			var userRole []Role
			if authOpt.username == "admin" {
				for _, role := range []DomainRole{DomainRoleAdmin} {
					userRole = append(userRole, roles[role])
				}
			} else {
				for _, role := range []DomainRole{DomainRoleDeveloper} {
					userRole = append(userRole, roles[role])
				}
			}
			return &UserSession{ID: authOpt.username, Roles: userRole}, nil
		}
	}
	slog.Debug("error authenticating user", "username", authOpt.username)
	return nil, nil
}

func (a *StaticAuthenticator) Close() error {
	return nil
}

type LDAPAuthenticator struct {
	address    string
	account    string
	password   string
	basedn     string
	connection *ldap.Conn
	//filter     string
}

// NewLDAPAuthenticator initialises an LDAP authenticator using
// the given LDAP server address, service account and password;
// moreover is stores the BaseDN used for subsequent queries.
func NewLDAPAuthenticator(account, password, address, basedn string) (*LDAPAuthenticator, error) {

	slog.Debug("connecting to LDAP server", "address", address, "account", account, "password", "*********", "base DN", basedn)

	// connect to the LDAP server
	connection, err := ldap.DialURL(address)
	if err != nil {
		slog.Error("failed to connect to LDAP", "address", address, "error", err)
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	slog.Debug("connected to LDAP server", "address", address)

	// bind with the service account to search the directory
	if err = connection.Bind(account, password); err != nil {
		slog.Error("failed to bind service account", "address", address, "account", account)
		return nil, fmt.Errorf("failed to bind service account: %w", err)
	}

	slog.Info("successfully connected to LDAP server")

	return &LDAPAuthenticator{
		address:    address,
		account:    account,
		password:   password,
		basedn:     basedn,
		connection: connection,
	}, nil
}

func (a *LDAPAuthenticator) Close() error {
	if a.connection != nil {
		return a.connection.Close()
	}
	return nil
}

func (a *LDAPAuthenticator) Authenticate(username, password string) (bool, error) {

	// search for the user's Distinguished Name (DN)
	search := ldap.NewSearchRequest(
		a.basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=person)(|(uid=%s)(sAMAccountName=%s)))", ldap.EscapeFilter(username), ldap.EscapeFilter(username)),
		[]string{"dn"}, // We only need to retrieve the DN, no other attributes
		nil,
	)

	result, err := a.connection.Search(search)
	if err != nil {
		slog.Error("failed to search for user", "username", username)
		return false, fmt.Errorf("failed to search for user: %w", err)
	}

	// handle search results
	if len(result.Entries) == 0 {
		return false, fmt.Errorf("user not found")
	}

	if len(result.Entries) > 1 {
		return false, fmt.Errorf("multiple users found with the same username")
	}

	slog.Debug("user successfully retrieved")

	// extract the user's exact DN from the search result
	dn := result.Entries[0].DN

	slog.Debug("user's DN found", "username", username, "dn", dn)

	connection, err := ldap.DialURL(a.address)
	if err != nil {
		slog.Error("error connecting to LDAP server", "address", a.address, "error", err)
		return false, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	slog.Debug("successfully connected to LDAP server")
	defer connection.Close()

	// Step 4: Re-Bind as the specific user to verify their password
	err = connection.Bind(dn, password)
	if err != nil {
		// if the error is LDAP Result Code 49 (Invalid Credentials), the password was wrong;
		// we return false, but no error, as this is an expected authentication failure.
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			slog.Error("invalid credentials", "error", err)
			return false, nil
		}
		// any other error means the bind failed for a system reason (e.g., connection lost)
		slog.Error("failed to authenticate user", "username", username, "error", err)
		return false, fmt.Errorf("failed to bind as user: %w", err)
	}

	// if the second bind succeeds, the credentials are valid!
	slog.Info("user successfully authenticated", "username", username)
	return true, nil
}

func (a *LDAPAuthenticator) Authenticate2(opts ...Option) (*UserSession, error) {

	authOpt := &AuthenticateOptions{
		username:            "",
		password:            "",
		skipPwdVerification: false,
	}
	for _, opt := range opts {
		opt(authOpt)
	}

	// search for the user's Distinguished Name (DN)
	search := ldap.NewSearchRequest(
		a.basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=person)(|(uid=%s)(sAMAccountName=%s)))", ldap.EscapeFilter(authOpt.username), ldap.EscapeFilter(authOpt.username)),
		[]string{"dn"}, // We only need to retrieve the DN, no other attributes
		nil,
	)

	result, err := a.connection.Search(search)
	if err != nil {
		slog.Error("failed to search for user", "username", authOpt.username)
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}

	// handle search results
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("multiple users found with the same username")
	}

	slog.Debug("user successfully retrieved")

	// extract the user's exact DN from the search result
	dn := result.Entries[0].DN

	slog.Debug("user's DN found", "username", authOpt.username, "dn", dn)

	connection, err := ldap.DialURL(a.address)
	if err != nil {
		slog.Error("error connecting to LDAP server", "address", a.address, "error", err)
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	slog.Debug("successfully connected to LDAP server")
	defer connection.Close()

	if !authOpt.skipPwdVerification {
		// Step 4: Re-Bind as the specific user to verify their password
		err = connection.Bind(dn, authOpt.password)
		if err != nil {
			// if the error is LDAP Result Code 49 (Invalid Credentials), the password was wrong;
			// we return false, but no error, as this is an expected authentication failure.
			if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
				slog.Error("invalid credentials", "error", err)
				return nil, nil
			}
			// any other error means the bind failed for a system reason (e.g., connection lost)
			slog.Error("failed to authenticate user", "username", authOpt.username, "error", err)
			return nil, fmt.Errorf("failed to bind as user: %w", err)
		}
	}

	// extract the user's exact DN from the search result
	ownedRoles := []DomainRole{}
	memberOf := getAttribute(result.Entries[0], "memberOf")
	memberSet := make(map[string]struct{}, len(memberOf))
	for _, group := range memberOf {
		cn := strings.Split(strings.Split(group, ",")[0], "=")[1]
		memberSet[cn] = struct{}{}
	}
	for key := range memberSet {
		slog.Debug("user's attributes found", "username", authOpt.username, "memberOf", key)
	}

	// Check application allowed roles
	if _, ok := memberSet[string(DomainRoleAdmin)]; ok {
		ownedRoles = append(ownedRoles, DomainRoleAdmin)
	}
	if _, ok := memberSet[string(DomainRoleDeveloper)]; ok {
		ownedRoles = append(ownedRoles, DomainRoleDeveloper)
	}

	// if the second bind succeeds, the credentials are valid!
	slog.Info("user successfully authenticated", "username", authOpt.username)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve ldap user roles")
	}
	var userRole []Role
	for _, role := range ownedRoles {
		userRole = append(userRole, roles[role])
	}
	return &UserSession{
		ID:    authOpt.username,
		Roles: userRole,
	}, nil
}

func getAttribute(entry *ldap.Entry, name string) []string {
	for _, attr := range entry.Attributes {
		if attr.Name == name {
			return attr.Values
		}
	}
	return nil
}
