package proxy

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gdbu/jump"
	"github.com/gdbu/jump/permissions"
	"github.com/mojura/kiroku"
	"github.com/vroomy/httpserve"
	"github.com/vroomy/vroomy"
)

var p Plugin

const defaultMatch = "[0-9]+"

func init() {
	if err := vroomy.Register("proxy", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	mux sync.Mutex
	vroomy.BasePlugin

	match *regexp.Regexp

	// Backend
	Jump   *jump.Jump    `vroomy:"jump"`
	Source kiroku.Source `vroomy:"mojura-source"`
}

// New ensures Profiles Database is built and open for access
func (p *Plugin) Load(env vroomy.Environment) (err error) {
	var (
		matchExpression string
		ok              bool
	)

	if matchExpression, ok = env["matchExpression"]; !ok {
		matchExpression = defaultMatch
	}

	if p.match, err = regexp.Compile(matchExpression); err != nil {
		err = fmt.Errorf("error compiling match expression of <%s>", matchExpression)
		return
	}

	err = p.Jump.Permissions().SetMultiPermissions("proxy",
		permissions.NewPair("admins", jump.PermRWD),
		permissions.NewPair("proxy-write", jump.PermRW),
		permissions.NewPair("proxy-read", jump.PermR),
	)

	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p
}

// Ingest will ingest logs and set IDs if necessary
func (p *Plugin) Export(ctx *httpserve.Context) {
	p.mux.Lock()
	defer p.mux.Unlock()
	req := ctx.Request()
	filename := updateFilename(ctx.Param("filename"))

	var (
		f   *os.File
		err error
	)

	if f, err = os.CreateTemp("", "exporting"); err != nil {
		err = fmt.Errorf("error creating temporary file: %v", err)
		ctx.WriteJSON(400, err)
		return
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()

	if _, err = io.Copy(f, req.Body); err != nil {
		err = fmt.Errorf("error copying to temporary file: %v", err)
		ctx.WriteJSON(400, err)
		return
	}

	if err := p.Source.Export(req.Context(), filename, f); err != nil {
		err = fmt.Errorf("error exporting: %v", err)
		ctx.WriteJSON(400, err)
		return
	}

	ctx.WriteNoContent()
}

// Get will get a file by name
func (p *Plugin) Get(ctx *httpserve.Context) {
	req := ctx.Request()
	filename := ctx.Param("filename")
	if err := p.Source.Import(req.Context(), filename, ctx.Writer()); err != nil {
		err = fmt.Errorf("error getting: %v", err)
		ctx.WriteJSON(400, err)
		return
	}
}

// Get will get a file by name
func (p *Plugin) GetNext(ctx *httpserve.Context) {
	var (
		nextFilename string
		err          error
	)

	req := ctx.Request()
	resource := ctx.Get("resource")
	lastFilename := ctx.Param("filename")
	if nextFilename, err = p.Source.GetNext(req.Context(), resource, lastFilename); err != nil {
		err = fmt.Errorf("error getting next filename: %v", err)
		ctx.WriteJSON(400, err)
		return
	}

	ctx.WriteJSON(200, nextFilename)
}

func (p *Plugin) CheckPermissionsMW(ctx *httpserve.Context) {
	filename := ctx.Param("filename")
	partEnd := strings.Index(filename, ".")
	if partEnd == -1 {
		msg := fmt.Errorf("invalid filename <%s> does not have a part separator", filename)
		ctx.WriteJSON(400, msg)
		return
	}

	resource, err := url.PathUnescape(filename[0:partEnd])
	if err != nil {
		msg := fmt.Errorf("error unescaping filename: %v", err)
		ctx.WriteJSON(400, msg)
		return
	}

	resource = strings.Replace(resource, "_latestSnapshots/", "", 1)
	ctx.Put("resource", resource)
	p.Jump.NewCheckPermissionsMW(resource, "")(ctx)
}
