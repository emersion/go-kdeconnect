package plugin

import (
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

const NotificationType protocol.PackageType = "kdeconnect.notification"

type NotificationBody struct {
	Id string `json:"id,omitempty"`
	AppName string `json:"appName,omitempty"`
	IsClearable bool `json:"isClearable,omitempty"`
	IsCancel bool `json:"isCancel,omitempty"`
	Ticker string `json:"ticker,omitempty"`
	Time string `json:"time,omitempty"`
	Silent bool `json:"silent,omitempty"`

	Request bool `json:"request,omitempty"`
	Cancel string `json:"cancel,omitempty"`
}

type NotificationEvent struct {
	Event
	NotificationBody
}

func (e *NotificationEvent) Cancel() {
	e.Device.Send(NotificationType, &NotificationBody{Cancel: e.Id})
}

type Notification struct {
	Incoming chan *NotificationEvent
}

func (p *Notification) GetDisplayName() string {
	return "Notification"
}

func (p *Notification) GetSupportedPackages() map[protocol.PackageType]BodyFactory {
	return map[protocol.PackageType]BodyFactory{
		NotificationType: func() interface{} { return &NotificationBody{} },
	}
}

func (p *Notification) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != NotificationType {
		return false
	}

	event := &NotificationEvent{NotificationBody: *pkg.Body.(*NotificationBody)}
	event.Device = device

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Notification) SendRequest(device *network.Device) error {
	return device.Send(NotificationType, &NotificationBody{Request: true})
}

func NewNotification() *Notification {
	return &Notification{
		Incoming: make(chan *NotificationEvent),
	}
}
