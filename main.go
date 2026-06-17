package main

import (
	"github.com/chuccp/httpPush/cluster"
	"github.com/chuccp/httpPush/ex"
	"log"
	"os"
	"runtime"

	"github.com/chuccp/httpPush/api"
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/core"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configFile := "config.ini"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	cfg, err := config.LoadFile(configFile)
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
