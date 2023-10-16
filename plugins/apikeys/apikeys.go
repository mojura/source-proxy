package plugin

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/vroomy/vroomy"

	"github.com/mojura/source-proxy/libs/apikeys"
)

var p Plugin

func init() {
	if err := vroomy.Register("apikeys", &p); err != nil {
		log.Fatal(err)
	}
}

type Plugin struct {
	vroomy.BasePlugin

	apikeys *apikeys.APIKeys
}

// Load will initialize the APIKeys client
func (p *Plugin) Load(env vroomy.Environment) (err error) {
	configPath := env.Get("APIKEYS_PATH")
	if configPath == "" {
		configPath = "./"
	}

	fullPath := path.Join(configPath, "apikeys.json")

	var f *os.File
	if f, err = os.Open(fullPath); err != nil {
		return
	}
	defer f.Close()

	var as []apikeys.Entry
	if err = json.NewDecoder(f).Decode(&as); err != nil {
		return
	}

	p.apikeys = apikeys.New(as...)
	return
}

// Backend exposes this plugin's data layer to other plugins
func (p *Plugin) Backend() interface{} {
	return p.apikeys
}
