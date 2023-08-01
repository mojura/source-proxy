package plugin

import (
	"encoding/json"
	"log"
	"os"

	"github.com/vroomy/vroomy"

	"github.com/mojura/source-proxy/libs/resources"
)

var p Plugin

func init() {
	if err := vroomy.Register("resources", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin

	resources *resources.Resources
}

// Load will initialize the APIKeys client
func (p *Plugin) Load(env vroomy.Environment) (err error) {
	var f *os.File
	if f, err = os.Open("./resources.json"); err != nil {
		return
	}
	defer f.Close()

	var rs []resources.Entry
	if err = json.NewDecoder(f).Decode(&rs); err != nil {
		return
	}

	p.resources = resources.New(rs...)
	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p.resources
}
