package main

import (
	"DurakGo/server"
	"DurakGo/config"
)

func main() {
	// TODO Replace this to get env from file
	conf := config.GetConfiguration("DEV")
	server.InitServer(conf)
}
