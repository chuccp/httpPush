package main

import (
	"github.com/chuccp/httpPush/cluster"
	"github.com/chuccp/httpPush/ex"
	"log"
	"runtime"

	"github.com/chuccp/httpPush/api"
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/core"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg, err := config.LoadFile("config.ini")
	if err != nil {
		log.Panic(err)
		return
	}
	register := core.NewRegister(cfg)
	register.AddServer(api.NewServer())
	register.AddServer(cluster.NewServer())
	register.AddServer(ex.NewServer())
	httpPush := register.Create()
	httpPush.Start()

}
