package main

import (
	"context"

	httpproxy "github.com/devdevaraj/firestarter/http_proxy"
	init_app "github.com/devdevaraj/firestarter/init_app"
	"github.com/devdevaraj/firestarter/launcher"
	wsserver "github.com/devdevaraj/firestarter/ws_server"
)

var WSSPort = "8080"
var HTTPPort = "80"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, args := init_app.Init()

	go func() {
		wsserver.WebSockerServer(WSSPort, *cfg)
	}()

	go func() {
		httpproxy.HTTPProxy(HTTPPort)
	}()

	launcher.Launcher(*cfg, ctx, args)
	select {}
}
