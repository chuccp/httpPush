package main

import (
	"context"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/httpPush/api"
	"github.com/chuccp/httpPush/cluster"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/ex"
	"github.com/chuccp/httpPush/ws"
)

func main() {
	cfg := config.LoadAutoConfig()
	app := core.NewApp(cfg)

	builder := wf.NewBuilder(cfg)
	builder.Service(cluster.NewService(app))
	builder.Rest(api.NewController(app))
	builder.Rest(ex.NewController(app))
	builder.Rest(ws.NewController(app))
	builder.Build().Run(context.Background())
}
