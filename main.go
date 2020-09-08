package main

import (
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/server"
)

func main(){
	config.Init()
	server.InitServer()
}
