package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/mojura/kiroku"
	"github.com/mojura/source-proxy/libs/apikeys"
	"github.com/mojura/source-proxy/libs/resources"
	"github.com/vroomy/httpserve"
	"github.com/vroomy/vroomy"
)

var p Plugin

const defaultMatch = "[0-9]+"

var forbidden = errors.New("forbidden")

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
	Source    kiroku.Source        `vroomy:"mojura-source"`
	APIKeys   *apikeys.APIKeys     `vroomy:"apikeys"`
	Resources *resources.Resources `vroomy:"resources"`
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

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		err = fmt.Errorf("error seeking within temporary file: %v", err)
		ctx.WriteJSON(500, err)
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
	var (
		apikey string
		err    error
	)

	if apikey, err = getAPIKey(ctx); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	var resource string
	filename := ctx.Param("filename")
	if resource, err = getResource(filename); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	method := ctx.Request().Method
	groups := p.APIKeys.Groups(apikey)

	if !p.Resources.Can(method, resource, groups...) {
		ctx.WriteJSON(401, forbidden)
		return
	}

	ctx.Put("resource", resource)
}
