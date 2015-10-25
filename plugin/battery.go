package plugin

import (
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

const BatteryType protocol.PackageType = "kdeconnect.battery"

const (
	BatteryThresholdEventNone = 0
	BatteryThresholdEventLow = 1
)

type BatteryBody struct {
	CurrentCharge int `json:"currentCharge,omitempty"`
	IsCharging bool `json:"isCharging,omitempty"`
	ThresholdEvent int `json:"thresholdEvent,omitempty"`

	Request bool `json:"request,omitempty"`
}

type BatteryEvent struct {
	Event
	BatteryBody
}

type Battery struct {
	Incoming chan *BatteryEvent
}

func (p *Battery) GetDisplayName() string {
	return "Battery"
}

func (p *Battery) GetSupportedPackages() map[protocol.PackageType]interface{} {
	return map[protocol.PackageType]interface{}{
		BatteryType: new(BatteryBody),
	}
}

func (p *Battery) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != BatteryType {
		return false
	}

	event := &BatteryEvent{BatteryBody: *pkg.Body.(*BatteryBody)}
	event.Device = device

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Battery) SendRequest(device *network.Device) error {
	return device.Send(BatteryType, &BatteryBody{Request: true})
}

func NewBattery() *Battery {
	return &Battery{
		Incoming: make(chan *BatteryEvent),
	}
}
