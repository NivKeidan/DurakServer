package main

import (
	"DurakGo/output"
	"DurakGo/server"
	"DurakGo/config"
	"flag"

)

func main() {
	// TODO Replace this to get env from file
	var debug = flag.Bool("debug", false, "Verbose output")
	flag.Parse()
	output.SetOutput(*debug)

	conf := config.GetConfiguration("DEV")
	server.InitServer(conf)
}
