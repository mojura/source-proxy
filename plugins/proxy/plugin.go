package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/mojura/kiroku"
	"github.com/mojura/source-proxy/libs/apikeys"
	"github.com/mojura/source-proxy/libs/resources"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/vroomy/httpserve"
	"github.com/vroomy/vroomy"
)

var p Plugin

const defaultMatch = "[0-9]+"

var errForbidden = errors.New("forbidden")

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

	getsStarted   prometheus.Counter
	getsCompleted prometheus.Counter
	getsErrored   prometheus.Counter

	getNextsStarted   prometheus.Counter
	getNextsCompleted prometheus.Counter
	getNextsErrored   prometheus.Counter

	getNextListsStarted   prometheus.Counter
	getNextListsCompleted prometheus.Counter
	getNextListsErrored   prometheus.Counter

	getHeadStarted   prometheus.Counter
	getHeadCompleted prometheus.Counter
	getHeadErrored   prometheus.Counter

	exportsStarted   prometheus.Counter
	exportsCompleted prometheus.Counter
	exportsErrored   prometheus.Counter
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

	p.getsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_gets_started_total",
		Help: "The number of Get events started",
	})

	p.getsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_gets_completed_total",
		Help: "The number of Get events completed",
	})

	p.getsErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_gets_errored_total",
		Help: "The number of Get events with errors",
	})

	p.getNextsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_nexts_started_total",
		Help: "The number of GetNext events started",
	})

	p.getNextsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_nexts_completed_total",
		Help: "The number of GetNext events completed",
	})

	p.getNextListsErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_next_lists_errored_total",
		Help: "The number of GetNextList events with errors",
	})

	p.getNextListsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_next_lists_started_total",
		Help: "The number of GetNextList events started",
	})

	p.getNextListsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_next_lists_completed_total",
		Help: "The number of GetNextList events completed",
	})

	p.getNextsErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_get_nexts_errored_total",
		Help: "The number of GetNext events with errors",
	})

	p.exportsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_exports_started_total",
		Help: "The number of Export events started",
	})

	p.exportsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_exports_completed_total",
		Help: "The number of Export events completed",
	})

	p.exportsErrored = promauto.NewCounter(prometheus.CounterOpts{
		Name: "source_proxy_exports_errored_total",
		Help: "The number of exExportport events with errors",
	})
	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p
}

// Ingest will ingest logs and set IDs if necessary
func (p *Plugin) Export(ctx *httpserve.Context) {
	var (
		newFilename string
		err         error
	)

	p.exportsStarted.Add(1)
	req := ctx.Request()
	prefix := ctx.Param("prefix")
	p.mux.Lock()
	filename := updateFilename(ctx.Param("filename"))
	p.mux.Unlock()

	// We need to copy the request body to a file so that the s3 library can determine the max content length
	if err = copyToTemp(req.Body, func(f *os.File) (err error) {
		if newFilename, err = p.Source.Export(req.Context(), prefix, filename, f); err != nil {
			err = fmt.Errorf("error exporting: %v", err)
			return
		}

		return
	}); err != nil {
		ctx.WriteJSON(400, err)
		log.Printf("error exporting prefix %v: %v req: %v", prefix, err, req)
		p.exportsErrored.Add(1)
		return
	}

	ctx.WriteString(200, "text/plain", newFilename)
	p.exportsCompleted.Add(1)
}

// Get will get a file by name
func (p *Plugin) Get(ctx *httpserve.Context) {
	p.getsStarted.Add(1)
	req := ctx.Request()
	prefix := ctx.Param("prefix")
	filename := ctx.Param("filename")
	if err := p.Source.Import(req.Context(), prefix, filename, ctx.Writer()); err != nil {
		log.Printf("error getting: %v: %v req: %v", prefix, err, req)
		err = fmt.Errorf("error getting: %v", err)
		ctx.WriteJSON(400, err)
		p.getsErrored.Add(1)
		return
	}

	p.getsCompleted.Add(1)
}

// Get will get a file by name
func (p *Plugin) GetNext(ctx *httpserve.Context) {
	var (
		nextFilename string
		err          error
	)

	p.getNextsStarted.Add(1)
	req := ctx.Request()
	prefix := ctx.Param("prefix")
	lastFilename := ctx.Param("filename")
	if nextFilename, err = p.Source.GetNext(req.Context(), prefix, lastFilename); err != nil {
		log.Printf("error getting next filename: %v: %v req: %v", prefix, err, req)
		err = fmt.Errorf("error getting next filename: %v", err)
		ctx.WriteJSON(400, err)
		p.getNextsErrored.Add(1)
		return
	}

	if lastFilename >= nextFilename {
		ctx.WriteJSON(400, io.EOF)
		p.getNextsErrored.Add(1)
	}

	ctx.WriteJSON(200, nextFilename)
	p.getNextsCompleted.Add(1)
}

// GetNextList will get a list of next files
func (p *Plugin) GetNextList(ctx *httpserve.Context) {
	var (
		nextFilenames []string
		err           error
	)

	p.getNextListsStarted.Inc()
	req := ctx.Request()
	prefix := ctx.Param("prefix")
	lastFilename := ctx.Param("filename")
	maxKeysStr := ctx.Param("maxKeys")

	var maxKeys int64
	if maxKeys, err = strconv.ParseInt(maxKeysStr, 10, 64); err != nil {
		log.Printf("error parsing maxKeys parameter <%s>: %v", maxKeysStr, err)
		err = fmt.Errorf("error parsing maxKeys parameter: %v", err)
		ctx.WriteJSON(400, err)
		p.getNextsErrored.Add(1)
		return
	}

	if nextFilenames, err = p.Source.GetNextList(req.Context(), prefix, lastFilename, maxKeys); err != nil {
		log.Printf("error getting next filename: %v: %v req: %v", prefix, err, req)
		err = fmt.Errorf("error getting next filename: %v", err)
		ctx.WriteJSON(400, err)
		p.getNextsErrored.Add(1)
		return
	}

	ctx.WriteJSON(200, nextFilenames)
	p.getNextsCompleted.Inc()
}

// GetHead will get the info for a file
func (p *Plugin) GetHead(ctx *httpserve.Context) {
	var (
		info kiroku.Info
		err  error
	)

	p.getHeadStarted.Inc()
	req := ctx.Request()
	prefix := ctx.Param("prefix")
	filename := ctx.Param("filename")

	if info, err = p.Source.GetHead(req.Context(), prefix, filename); err != nil {
		log.Printf("error getting head for filename: %v: %v req: %v", prefix, err, req)
		err = fmt.Errorf("error getting head for filename: %v", err)
		ctx.WriteJSON(400, err)
		p.getHeadErrored.Add(1)
		return
	}

	ctx.WriteJSON(200, info)
	p.getHeadCompleted.Inc()
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
	prefix := ctx.Param("prefix")
	filename := ctx.Param("filename")
	if resource, err = getResource(prefix, filename); err != nil {
		ctx.WriteJSON(400, err)
		return
	}

	method := ctx.Request().Method
	groups := p.APIKeys.Groups(apikey)

	if !p.Resources.Can(method, resource, groups...) {
		fmt.Printf("forbidden request: Prefix: <%s> / Filename: <%s> / Resource <%s> / Last 4 API Key <%s>\n", prefix, filename, resource, apikey[len(apikey)-4:])
		ctx.WriteJSON(401, errForbidden)
		return
	}

	ctx.Put("resource", resource)
}
