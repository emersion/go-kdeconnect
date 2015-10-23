package plugin

import (
	"log"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
)

const PingType netpkg.Type = "kdeconnect.ping"

type Ping struct {}

func (p *Ping) GetDisplayName() string {
	return "Ping"
}

func (p *Ping) GetSupportedPackages() map[netpkg.Type]interface{} {
	return map[netpkg.Type]interface{}{}
}

func (p *Ping) Handle(device *network.Device, pkg *netpkg.Package) bool {
	if pkg.Type != PingType {
		return false
	}

	log.Println("Received a ping!")

	return true
}

func (p *Ping) SendPing(device *network.Device) error {
	return device.Send(PingType, nil)
}
