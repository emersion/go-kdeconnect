package plugin

import (
	"log"
	"github.com/emersion/go-kdeconnect/netpkg"
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

func (p *Battery) Handle(pkg *netpkg.Package) bool {
	if pkg.Type != BatteryType {
		return false
	}

	log.Println("Battery:", pkg.Body)

	return true
}
