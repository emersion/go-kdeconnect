package plugin

import (
	"log"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
)

const BatteryType netpkg.Type = "kdeconnect.battery"

type BatteryBody struct {
	CurrentCharge int `json:"currentCharge,omitempty"`
	IsCharging bool `json:"isCharging,omitempty"`
	ThresholdEvent int `json:"thresholdEvent,omitempty"`
	Request bool `json:"request,omitempty"`
}

type Battery struct {}

func (p *Battery) GetDisplayName() string {
	return "Battery"
}

func (p *Battery) GetSupportedPackages() map[netpkg.Type]interface{} {
	return map[netpkg.Type]interface{}{
		BatteryType: new(BatteryBody),
	}
}

func (p *Battery) Handle(device *network.Device, pkg *netpkg.Package) bool {
	if pkg.Type != BatteryType {
		return false
	}

	log.Println("Battery:", pkg.Body)

	return true
}

func (p *Battery) SendRequest(device *network.Device) error {
	return device.Send(BatteryType, &BatteryBody{Request: true})
}
