package main

import (
	"github.com/emersion/go-kdeconnect/engine"
)

func main() {
	e := engine.New()
	e.Listen()
}
