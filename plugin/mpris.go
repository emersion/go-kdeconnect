package plugin

import (
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

const MprisType protocol.PackageType = "kdeconnect.mpris"

const (
	MprisActionPlayPause = "PlayPause"
)

type MprisBody struct {
	Player string `json:"player,omitempty"`

	NowPlaying string `json:"nowPlaying,omitempty"`
	Volume int `json:"volume,omitempty"`
	IsPlaying bool `json:"isPlaying,omitempty"`
	Length float32 `json:"length,omitempty"`
	Pos float32 `json:"pos,omitempty"`
	PlayerList []string `json:"playerList,omitempty"`

	Action string `json:"action,omitempty"`
	SetVolume int `json:"setVolume,omitempty"`
	SetPosition float32 `json:"SetPosition,omitempty"`
	Seek float32 `json:"Seek,omitempty"`
	RequestPlayerList bool `json:"requestPlayerList,omitempty"`
	RequestNowPlaying bool `json:"requestNowPlaying,omitempty"`
	RequestVolume bool `json:"requestVolume,omitempty"`
}

type MprisEvent struct {
	Event
	MprisBody
}

type Mpris struct {
	Incoming chan *MprisEvent
}

func (p *Mpris) GetDisplayName() string {
	return "Media control"
}

func (p *Mpris) GetSupportedPackages() map[protocol.PackageType]interface{} {
	return map[protocol.PackageType]interface{}{
		MprisType: new(MprisBody),
	}
}

func (p *Mpris) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != MprisType {
		return false
	}

	event := &MprisEvent{MprisBody: *pkg.Body.(*MprisBody)}
	event.Device = device

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Mpris) SendAction(device *network.Device, action string) error {
	return device.Send(MprisType, &MprisBody{Action: action})
}

func NewMpris() *Mpris {
	return &Mpris{
		Incoming: make(chan *MprisEvent),
	}
}
