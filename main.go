package main

import (
	"context"
	"log"

	"github.com/hatchify/closer"
	"github.com/vroomy/vroomy"

	_ "github.com/vroomy-ext/digitalocean-s3-plugin"

	_ "github.com/mojura/source-proxy/plugins/apikeys"
	_ "github.com/mojura/source-proxy/plugins/proxy"
	_ "github.com/mojura/source-proxy/plugins/resources"
)

func main() {
	var (
		svc *vroomy.Vroomy
		err error
	)

	if svc, err = vroomy.New("./config.toml"); err != nil {
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
