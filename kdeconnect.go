package main

import (
	"github.com/emersion/go-kdeconnect/engine"
	"github.com/emersion/go-kdeconnect/plugin"
)

func main() {
	battery := &plugin.Battery{}
	ping := &plugin.Ping{}

	p := plugin.NewHandler()
	p.Register(battery)
	p.Register(ping)

	e := engine.New(p)
	e.Listen()
}
