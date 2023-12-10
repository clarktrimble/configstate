package main

import (
	"context"
	"os"
	"sync"

	"github.com/clarktrimble/delish"
	"github.com/clarktrimble/delish/graceful"
	"github.com/clarktrimble/giant"
	"github.com/clarktrimble/hondo"
	"github.com/clarktrimble/launch"
	"github.com/clarktrimble/sabot"

	"configstate/chi"
	"configstate/consul"
	"configstate/discover"
)

const (
	appId     string = "demo-cfgstate"
	cfgPrefix string = "demo"
	blerb     string = "'demo' demonstrates ..."
)

var (
	version string
	wg      sync.WaitGroup
)

type Config struct {
	Version  string           `json:"version" ignored:"true"`
	Logger   *sabot.Config    `json:"logger"`
	Client   *giant.Config    `json:"http_client"`
	Server   *delish.Config   `json:"http_server"`
	Consul   *consul.Config   `json:"consul"`
	Discover *discover.Config `json:"discover"`
}

func main() {

	// load config and setup logger

	cfg := &Config{Version: version}
	launch.Load(cfg, cfgPrefix, blerb)

	lgr := cfg.Logger.New(os.Stdout)
	ctx := lgr.WithFields(context.Background(), "app_id", appId, "run_id", hondo.Rand(7))
	lgr.Info(ctx, "starting up", "config", cfg)

	// init graceful and create router

	ctx = graceful.Initialize(ctx, &wg, lgr)

	rtr := chi.New()
	rtr.Set("GET", "/config", delish.ObjHandler("config", cfg, lgr))

	// start discovery

	client := cfg.Client.NewWithTrippers(lgr)
	csl := cfg.Consul.New(client)
	dsc := cfg.Discover.New(lgr, csl)

	dsc.Start(ctx, &wg)
	rtr.Set("GET", "/services", dsc.GetServices)

	// delicious!

	server := cfg.Server.NewWithLog(ctx, rtr, lgr)
	server.Start(ctx, &wg)
	graceful.Wait(ctx)
}
