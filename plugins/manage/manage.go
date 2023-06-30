package manage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gdbu/jump"
	"github.com/gdbu/jump/apikeys"
	"github.com/gdbu/jump/permissions"
	"github.com/gdbu/jump/users"
	"github.com/vroomy/httpserve"
	"github.com/vroomy/vroomy"
)

var p Plugin

func init() {
	if err := vroomy.Register("manage", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin

	// Backend
	Jump *jump.Jump `vroomy:"jump"`
}

// New ensures Profiles Database is built and open for access
func (p *Plugin) Load(env vroomy.Environment) (err error) {
	groupsLocation := env["groups-location"]
	if len(groupsLocation) == 0 {
		groupsLocation = "./groups.json"
	}

	var f *os.File
	if f, err = os.Open(groupsLocation); err != nil {
		err = fmt.Errorf("error opening groups file: %v", err)
		return
	}
	defer f.Close()

	var schemas []GroupSchema
	if err = json.NewDecoder(f).Decode(&schemas); err != nil {
		return
	}

	if err = p.Jump.Permissions().SetPermissions("manage", "admins", jump.PermRWD); err != nil {
		return
	}

	for _, schema := range schemas {
		for _, pair := range schema.Permissions {
			var action permissions.Action
			switch {
			case pair.CanRead && pair.CanWrite:
				action = jump.PermRW
			case pair.CanRead:
				action = jump.PermR
			case pair.CanWrite:
				action = permissions.ActionWrite
			}

			if err = p.Jump.SetPermission(pair.Resource, schema.Group, action, jump.PermRWD); err != nil {
				return
			}
		}
	}

	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p
}

// GetCurrentUser will get the current user
func (p *Plugin) GetCurrentUser(ctx *httpserve.Context) {
	var (
		user *users.User
		err  error
	)

	userID := ctx.Get("userID")

	if user, err = p.Jump.GetUser(userID); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	ctx.WriteJSON(200, user)
}

// Create will create a new proxy user
func (p *Plugin) Create(ctx *httpserve.Context) {
	var (
		req CreateRequest
		err error
	)

	if err = ctx.Bind(&req); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	var resp CreateResponse
	if resp.UserID, resp.APIKey, err = p.Jump.CreateUser(req.Username, "", req.Groups...); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	ctx.WriteJSON(200, resp)
}

// RefreshAPIKey will refresh the API key of a user
func (p *Plugin) RefreshAPIKey(ctx *httpserve.Context) {
	var err error
	userID := ctx.Param("userID")
	if err = p.removeKeys(userID); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	var resp CreateResponse
	if resp.APIKey, err = p.Jump.APIKeys().New(userID, ""); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	resp.UserID = userID
	ctx.WriteJSON(200, resp)
}

// List gets a users list
func (p *Plugin) List(ctx *httpserve.Context) {
	var (
		us  []*users.User
		err error
	)

	if us, err = p.Jump.GetUsersList(); err != nil {
		return
	}

	ctx.WriteJSON(200, us)
}

func (p *Plugin) removeKeys(userID string) (err error) {
	var keys []*apikeys.APIKey
	a := p.Jump.APIKeys()
	if keys, err = a.GetByUser(userID); err != nil {
		err = fmt.Errorf("error gettig keys for user <%s>: %v", userID, err)
		return
	}

	for _, k := range keys {
		if _, err = a.Remove(k.Key); err != nil {
			err = fmt.Errorf("error removing key of <%s>: %v", k.Key, err)
			return
		}
	}

	return
}

type CreateRequest struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

type CreateResponse struct {
	Username string `json:"username,omitempty"`

	UserID string `json:"userID"`
	APIKey string `json:"apiKey"`
}

type GroupSchema struct {
	Group       string        `json:"group"`
	Permissions []Permissions `json:"permissions"`
}

type Permissions struct {
	Resource string `json:"resource"`
	CanRead  bool   `json:"canRead"`
	CanWrite bool   `json:"canWrite"`
}
