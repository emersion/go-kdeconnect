package main

import (
	"github.com/emersion/go-kdeconnect/engine"
	"github.com/emersion/go-kdeconnect/plugin"
)

func main() {
	p := plugin.NewHandler()
	p.Register(&plugin.Battery{})
	e := engine.New(p)
	e.Listen()
}
