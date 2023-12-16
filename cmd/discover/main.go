// Package main demonstrates a service that discovers other services available to it.
package main

// Todo: update modses!

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
	appId     string = "dsc-demo"
	cfgPrefix string = "dsc"
	blerb     string = "'discover' demonstrates service discovery"
)

var (
	version string
	wg      sync.WaitGroup
)

type Config struct {
	Version      string         `json:"version" ignored:"true"`
	Logger       *sabot.Config  `json:"logger"`
	ConsulClient *giant.Config  `json:"consul_http_client"`
	Consul       *consul.Config `json:"consul"`
	Server       *delish.Config `json:"http_server"`
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

	// start discovery and register handler

	client := cfg.ConsulClient.NewWithTrippers(lgr)
	csl := cfg.Consul.New(client)
	dsc := &discover.Discover{Poller: csl, Logger: lgr}

	dsc.Start(ctx, &wg)
	dsc.Register(rtr)

	// start server and wait for shutdown

	server := cfg.Server.NewWithLog(ctx, rtr, lgr)
	server.Start(ctx, &wg)
	graceful.Wait(ctx)
}
