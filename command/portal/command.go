package portal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/dihedron/devws/command/base"
	"github.com/dihedron/devws/internal/service"
	"github.com/dihedron/devws/openstack"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type Portal struct {
	base.Command

	Address string `short:"a" long:"address" description:"Address to bind the API to." default:":3000"`
}

type Link struct {
	Relation string `json:"rel"`
	Href     string `json:"href"`
}

func (l Link) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"rel":  l.Relation,
		"href": l.Href,
	})
}

type VM struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  *struct {
		Self     *Link `json:"self,omitempty"`
		Stop     *Link `json:"stop,omitempty"`
		Start    *Link `json:"start,omitempty"`
		Restart  *Link `json:"restart,omitempty"`
		Pause    *Link `json:"pause,omitempty"`
		Unpause  *Link `json:"unpause,omitempty"`
		Shelve   *Link `json:"shelve,omitempty"`
		Unshelve *Link `json:"unshelve,omitempty"`
	} `json:"links,omitempty"`
}

func NewVM(base string, id string, status string) *VM {
	return &VM{
		ID:     id,
		Status: status,
		//Link:   &Link{Relation: "self", Href: base + "/" + id},
		Links: &struct {
			Self     *Link `json:"self,omitempty"`
			Stop     *Link `json:"stop,omitempty"`
			Start    *Link `json:"start,omitempty"`
			Restart  *Link `json:"restart,omitempty"`
			Pause    *Link `json:"pause,omitempty"`
			Unpause  *Link `json:"unpause,omitempty"`
			Shelve   *Link `json:"shelve,omitempty"`
			Unshelve *Link `json:"unshelve,omitempty"`
		}{
			Self:     &Link{Relation: "self", Href: base + "/" + id},
			Stop:     &Link{Relation: "stop", Href: base + "/" + id + "/stop"},
			Start:    &Link{Relation: "start", Href: base + "/" + id + "/start"},
			Restart:  &Link{Relation: "restart", Href: base + "/" + id + "/restart"},
			Pause:    &Link{Relation: "pause", Href: base + "/" + id + "/pause"},
			Unpause:  &Link{Relation: "unpause", Href: base + "/" + id + "/unpause"},
			Shelve:   &Link{Relation: "shelve", Href: base + "/" + id + "/shelve"},
			Unshelve: &Link{Relation: "unshelve", Href: base + "/" + id + "/unshelve"},
		},
	}
}

func (cmd *Portal) Execute(args []string) error {
	slog.Info("starting portal and API server", "address", cmd.Address)

	var openstackService service.OpenstackServiceI
	var authenticator Authenticator
	var err error
	mock, found := os.LookupEnv("MOCK_SERVICES")
	if found {
		if mock == "y" {
			// define an authenticator
			authenticator = NewStaticAuthenticator(
				WithUser("admin", "QWERTY"),
				WithUser("developer", "QWERTY"),
			)
			openstackService, err = service.NewOpenstackMockService(context.Background())
		} else {
			err = fmt.Errorf("env variable MOCK_SERVICES must be 'y'")
		}
	} else {
		authenticator, err = NewLDAPAuthenticatorFromEnvs()
		openstackService, err = service.NewOpenstackService(context.Background())
	}

	if err != nil {
		slog.Error("Unable to use openstack service", "error", err.Error())
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
		unauthenticated.StaticFile("/favicon.ico", "./command/server/assets/favicon.ico")
		unauthenticated.StaticFile("/devws.png", "./command/server/assets/devws.png")
		unauthenticated.StaticFile("/style.css", "./command/server/assets/style.css")
		unauthenticated.StaticFile("/background.jpg", "./command/server/assets/background.jpg")
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
				}
				c.Redirect(http.StatusFound, "/api/v1/auth/login")
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

			options := []openstack.ComputeV2ListOption{}
			vms, err := openstackService.List(context.Background(), options)

			// vms := retrieveVms(c)
			data := NewTableData(vms, page)

			c.HTML(http.StatusOK, "_table.html", data)
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
			data := NewTableData(vms, 1)

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

	slog.Info("portal and API server running", "address", cmd.Address)
	err = router.Run(cmd.Address)
	if err != nil {
		slog.Error("portal and API server failed", "error", err)
		return fmt.Errorf("portal and API server failed: %w", err)
	}
	return nil
}

type TableData struct {
	Records      []openstack.Workstation
	Page         int
	TotalPages   int
	TotalRecords int
	PrevPage     int
	NextPage     int
	Pages        []int
}

func NewTableData(vms []openstack.Workstation, page int) *TableData {

	td := &TableData{}
	return td.Paginate(vms, page)

}

const pageSize = 10

func (t *TableData) Paginate(vms []openstack.Workstation, page int) *TableData {
	total := len(vms)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	// Build page number slice
	pages := make([]int, totalPages)
	for i := range pages {
		pages[i] = i + 1
	}

	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	next := page + 1
	if next > totalPages {
		next = totalPages
	}

	t.Records = vms[start:end]
	t.Page = page
	t.TotalPages = totalPages
	t.TotalRecords = total
	t.PrevPage = prev
	t.NextPage = next
	t.Pages = pages

	return t
}
