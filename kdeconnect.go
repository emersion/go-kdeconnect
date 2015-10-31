package main

import (
	"log"
	"github.com/emersion/go-kdeconnect/engine"
	"github.com/emersion/go-kdeconnect/plugin"
)

func main() {
	config := engine.DefaultConfig()

	battery := plugin.NewBattery()
	ping := plugin.NewPing()

	go (func() {
		for {
			select {
			case event := <-ping.Incoming:
				log.Println("New ping from device:", event.Device.Name)
			case event := <-battery.Incoming:
				log.Println("Battery:", event.Device.Name, event.BatteryBody)
			}
		}
	})()

	hdlr := plugin.NewHandler()
	hdlr.Register(battery)
	hdlr.Register(ping)

	e := engine.New(hdlr, config)

	go (func() {
		for {
			select {
			case device := <-e.RequestsPairing:
				e.PairDevice(device)
			}
		}
	})()

	e.Listen()
}
