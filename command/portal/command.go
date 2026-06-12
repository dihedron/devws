package portal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/dihedron/devws/command/base"
	"github.com/dihedron/devws/command/portal/dto"
	"github.com/dihedron/devws/internal/service"
	"github.com/dihedron/devws/openstack"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type Portal struct {
	base.Command
	Configuration *Configuration `short:"c" long:"configuration" description:"Path to the configuration file (prefixed by @)" default:"@devws.yaml"`
	// Address       string         `short:"a" long:"address" description:"Address to bind the API to." default:":3000"`
}

func (cmd *Portal) Execute(args []string) error {
	slog.Info("starting portal and API server", "address", cmd.Configuration.Address)

	var openstackService service.OpenstackServiceI
	var authenticator Authenticator
	var err error

	if cmd.Configuration != nil {
		authenticator, err = NewLDAPAuthenticator(
			cmd.Configuration.LDAP.Account,
			cmd.Configuration.LDAP.Password,
			cmd.Configuration.LDAP.Server,
			cmd.Configuration.LDAP.BaseDN,
		)
	} else {
		authenticator = NewStaticAuthenticator(
			WithUser("admin", "QWERTY"),
			WithUser("developer", "QWERTY"),
		)
	}

	openstackService, err = service.NewOpenstackService(context.Background())
	if err != nil {
		slog.Error("Unable to use openstack service", "error", err.Error())
	}

	mock, found := os.LookupEnv("MOCK_SERVICES")
	if found && mock == "y" {
		openstackService, err = service.NewOpenstackMockService(context.Background())
		if err != nil {
			slog.Error("Unable to use openstack mock service", "error", err.Error())
		}
	}

	router := gin.New()
	router.SetTrustedProxies(nil)

	// generate a session key from random bytes
	// this is used to secure the session cookie
	// authenticationKey := make([]byte, 32)
	// rand.Read(authenticationKey)
	// encryptionKey := make([]byte, 32)
	// rand.Read(encryptionKey)
	// store := cookie.NewStore(authenticationKey, encryptionKey)

	router.Use(
		Logger(),
		gin.Recovery(),
		sessions.Sessions("api_session", cookie.NewStore([]byte("super-secret-key"))),
	)

	router.SetFuncMap(template.FuncMap{})
	// router.LoadHTMLGlob("command/portal/assets/*.html")
	router.LoadHTMLGlob("command/portal/templates/*.html")

	unauthenticated := router.Group("")
	{
		unauthenticated.StaticFile("/favicon.ico", "command/server/assets/favicon.ico")
		unauthenticated.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			if username := session.Get("username"); username != nil {
				slog.Debug("user already logged in, redirecting to main page...")
				c.Redirect(http.StatusFound, "/api/v1/vm/")
			} else {
				slog.Debug("user not logged in yet, redirecting to login page")
				c.Redirect(http.StatusFound, "/api/v1/auth/login")
			}
		})
		// authentication endpoints: the /api/v1/auth/login and
		// /api/v1/auth/logout routes do not need authentication
		unauthenticated.GET("/api/v1/auth/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", nil)
		})
		unauthenticated.POST("/api/v1/auth/login", func(c *gin.Context) {
			username := c.PostForm("username")
			password := c.PostForm("password")
			slog.Debug("logging out user first...", "username", username)
			session := sessions.Default(c)
			if u := session.Get("username"); u == username {
				slog.Debug("user already logged in, redirecting to main page")
				c.Redirect(http.StatusFound, "/api/v1/vm")
			} else {
				slog.Debug("logging in user...", "username", username, "password", "*******")
				if ok, err := authenticator.Authenticate(username, password); ok {
					slog.Info("user successfully logged in", "username", username)
					session.Set("username", username)
					session.Save()
					c.Redirect(http.StatusFound, "/api/v1/vm")
					return
				} else {
					slog.Error("failed to authenticate user", "username", username, "error", err)
					c.HTML(http.StatusUnauthorized, "login_error.html", gin.H{
						"Error": "Invalid credentials",
					})
				}
			}

		})

	}

	authenticated := router.Group("/api/v1/vm", SessionAuthMiddleware("Developer Workstations Realm", authenticator))
	{

		authenticated.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			c.Header("HX-Redirect", "/")
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Username": session.Get("username"),
			})
		})

		authenticated.GET("/vms", func(c *gin.Context) {
			pageStr := c.DefaultQuery("page", "1")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				page = 1
			}

			// Filters
			filterName := strings.ToLower(c.Query("name"))
			filterOwner := c.Query("owner")
			filterUserid := c.Query("userid")
			filterImage := c.Query("image")
			filterStatus := c.Query("status")

			options := []openstack.ComputeV2ListOption{}
			if filterName != "" {
				options = append(options, openstack.WithName(filterName))
			}
			if filterOwner != "" {
				options = append(options, openstack.WithTags(fmt.Sprintf("devws.owner=%s", filterOwner)))
			}
			if filterUserid != "" {
				options = append(options, openstack.WithUserID(filterUserid))
			}
			if filterImage != "" {
				options = append(options, openstack.WithImage(filterImage))
			}
			if filterStatus != "" {
				options = append(options, openstack.WithStatus(filterStatus))
			}

			vms, err := openstackService.List(context.Background(), options)

			// vms := retrieveVms(c)
			data := dto.NewTableData(vms, page)

			c.HTML(http.StatusOK, "_table.html", data)
		})

		authenticated.GET("/vms/detail/:id", func(c *gin.Context) {
			pageStr := c.DefaultQuery("page", "1")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				page = 1
			}

			id := c.Param("id")

			vm, err := openstackService.View(context.Background(), id)

			if vm != nil {
				c.HTML(http.StatusOK, "_detail.html", vm)
				return
			}

			c.Status(http.StatusNotFound)
		})

		// POST /api/v1/vm/vms/:id/:action
		authenticated.POST("/vms/:id/:action", func(c *gin.Context) {
			id := c.Param("id")
			action := c.Param("action")
			slog.Info("POST requested", "id", id, "action", action)

			switch action {
			case "stop":
				slog.Debug("stop requested", "id", id)
				openstackService.Stop(context.Background(), id)
			case "start":
				slog.Debug("start requested", "id", id)
				openstackService.Start(context.Background(), id)
			case "reboot":
				slog.Debug("reboot requested", "id", id)
				openstackService.Reboot(context.Background(), id, servers.HardReboot)
			case "shelve":
				slog.Debug("shelve requested", "id", id)
				openstackService.Shelve(context.Background(), id, false)
			case "unshelve":
				slog.Debug("unshelve requested", "id", id)
				openstackService.Unshelve(context.Background(), id, "cdm")
			}

			options := []openstack.ComputeV2ListOption{}
			vms, _ := openstackService.List(context.Background(), options)
			data := dto.NewTableData(vms, 1)

			c.HTML(http.StatusOK, "_table.html", data)
		})

		authenticated.POST("/logout", func(c *gin.Context) {
			slog.Debug(("LOGOUT"))
			c.Header("HX-Redirect", "/")
			session := sessions.Default(c)
			if username := session.Get("username"); username != nil {
				slog.Debug("logging out user...", "username", username)
				session.Clear()
				session.Save()
			}
			c.Redirect(http.StatusFound, "/api/v1/auth/login")
		})

	}

	// /login
	// https://github.com/puikinsh/login-forms/tree/main/forms/glassmorphism
	// https://github.com/puikinsh/login-forms/tree/main/forms/material

	slog.Info("portal and API server running", "address", cmd.Configuration.Address)
	err = router.Run(cmd.Configuration.Address)
	if err != nil {
		slog.Error("portal and API server failed", "error", err)
		return fmt.Errorf("portal and API server failed: %w", err)
	}
	return nil
}
