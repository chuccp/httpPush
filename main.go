package main

import (
	"context"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
	wflog "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/httpPush/api"
	"github.com/chuccp/httpPush/cluster"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/ex"
	"github.com/chuccp/httpPush/ws"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		panic(err)
	}

	app := core.NewApp(cfg)

	builder := wf.NewBuilder(cfg)
	builder.Service(app)
	builder.Service(cluster.NewService(app))
	builder.Rest(api.NewController(app))
	builder.Rest(ex.NewController(app))
	builder.Rest(ws.NewController(app))
	if err := builder.Build().Run(context.Background()); err != nil {
		wflog.Fatal("server stopped", zap.Error(err))
	}
}
