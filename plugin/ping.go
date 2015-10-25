package plugin

import (
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

const PingType protocol.PackageType = "kdeconnect.ping"

type PingEvent struct {
	Event
}

type Ping struct {
	Incoming chan *PingEvent
}

func (p *Ping) GetDisplayName() string {
	return "Ping"
}

func (p *Ping) GetSupportedPackages() map[protocol.PackageType]BodyFactory {
	return nil
}

func (p *Ping) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != PingType {
		return false
	}

	event := &PingEvent{}
	event.Device = device

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Ping) SendPing(device *network.Device) error {
	return device.Send(PingType, nil)
}

func NewPing() *Ping {
	return &Ping{
		Incoming: make(chan *PingEvent),
	}
}
