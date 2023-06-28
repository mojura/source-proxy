package plugin

import (
	"log"

	"github.com/mojura/source-proxy/client"
	"github.com/vroomy/vroomy"
)

var p Plugin

func init() {
	if err := vroomy.Register("mojura-source", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin

	client *client.Client
}

// Load will initialize the s3 client
func (p *Plugin) Load(env vroomy.Environment) (err error) {
	host := env["source-proxy-host"]
	apiKey := env["source-proxy-apikey"]

	p.client, err = client.New(host, apiKey)
	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p.client
}
