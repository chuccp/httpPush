package main

import (
	"context"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/component/cors"
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
	builder := wf.NewBuilder(cfg)
	builder.Service(core.NewApp())
	builder.Service(cluster.NewService())
	builder.Rest(api.NewController())
	builder.Rest(ex.NewController())
	builder.Rest(ws.NewController())
	builder.Filter(cors.NewCrosFilter())
	if err := builder.Build().Run(context.Background()); err != nil {
		wflog.Fatal("server stopped", zap.Error(err))
	}
}
