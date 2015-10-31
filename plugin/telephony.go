package plugin

import (
	"log"
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

const TelephonyType protocol.PackageType = "kdeconnect.telephony"

type TelephonyEventName string
const (
	TelephonyRinging TelephonyEventName = "ringing"
	TelephonyTalking = "talking"
	TelephonyMissedCall = "missedCall"
	TelephonySms = "sms"
)

type TelephonyAction string
const (
	TelephonyMute TelephonyAction = "mute"
)

type TelephonyBody struct {
	Event TelephonyEventName `json:"event",omitempty`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	ContactName string `json:"contactName,omitempty"`

	MessageBody string `json:"messageBody,omitempty"`

	isCancel string `json:"isCancel,omitempty"` // It is a string but must be a bool
	IsCancel bool `json:"-"`

	Action TelephonyAction `json:"action,omitempty"`
}

type TelephonyEvent struct {
	Event
	TelephonyBody
}

type Telephony struct {
	Incoming chan *TelephonyEvent
}

func (p *Telephony) GetDisplayName() string {
	return "Telephony"
}

func (p *Telephony) GetSupportedPackages() map[protocol.PackageType]BodyFactory {
	return map[protocol.PackageType]BodyFactory{
		TelephonyType: func() interface{} { return &TelephonyBody{} },
	}
}

func (p *Telephony) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != TelephonyType {
		return false
	}

	event := &TelephonyEvent{TelephonyBody: *pkg.Body.(*TelephonyBody)}
	event.Device = device

	// Workaround because isCancel is sent as a string, not a bool
	if event.isCancel != "" {
		event.IsCancel = true
	}

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Telephony) SendAction(device *network.Device, action TelephonyAction) error {
	return device.Send(TelephonyType, &TelephonyBody{Action: action})
}

func NewTelephony() *Telephony {
	return &Telephony{
		Incoming: make(chan *TelephonyEvent),
	}
}
