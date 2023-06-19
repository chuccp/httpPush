package main

import (
	"github.com/chuccp/httpPush/api"
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/core"
	"log"
)

func main() {

	cfg, err := config.LoadFile("config.ini")
	if err != nil {
		log.Print(err)
		return
	}
	register := core.NewRegister(cfg)
	register.AddServer(&api.Server{})
	httpPush := register.Create()
	err = httpPush.Start()
	if err != nil {
		log.Print(err)
	}
}
