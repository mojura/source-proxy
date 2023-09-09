package plugin

import (
	"log"

	"github.com/vroomy/httpserve"
	"github.com/vroomy/vroomy"
)

func init() {
	var (
		p   Plugin
		err error
	)

	if err = vroomy.Register("health", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin
}

// Login is the handler for serving the login page
func (p *Plugin) Ping(ctx *httpserve.Context) {
	ctx.WriteString(200, "text/plain", "PONG")
}
