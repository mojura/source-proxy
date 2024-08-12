package main

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/hatchify/closer"
	"github.com/vroomy/vroomy"

	_ "go.uber.org/automaxprocs"

	_ "github.com/vroomy-ext/digitalocean-s3-plugin"
	_ "github.com/vroomy-ext/prometheus-plugin"

	_ "github.com/mojura/source-proxy/plugins/apikeys"
	_ "github.com/mojura/source-proxy/plugins/health"
	_ "github.com/mojura/source-proxy/plugins/proxy"
	_ "github.com/mojura/source-proxy/plugins/resources"
)

func main() {
	var (
		svc *vroomy.Vroomy
		err error
	)

	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = "./"
	}

	fullPath := path.Join(configPath, "config.toml")
	if svc, err = vroomy.New(fullPath); err != nil {
		log.Fatal(err)
	}

	c := closer.New()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = c.Wait()
		cancel()
	}()

	if err = svc.Listen(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}

	if err = svc.Close(); err != nil {
		log.Fatal(err)
	}
}
