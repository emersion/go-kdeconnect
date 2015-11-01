package plugin

import (
	"github.com/emersion/go-kdeconnect/network"
	"github.com/emersion/go-kdeconnect/protocol"
)

const SftpType protocol.PackageType = "kdeconnect.sftp"

type SftpBody struct {
	Ip         string   `json:"ip,omitempty"`
	Port       int      `json:"port,omitempty"`
	User       string   `json:"user,omitempty"`
	Password   string   `json:"password,omitempty"`
	MultiPaths []string `json:"multiPaths,omitempty"`
	PathNames  []string `json:"pathNames,omitempty"`

	StartBrowsing bool `json:"startBrowsing,omitempty"`
}

type SftpEvent struct {
	Event
	SftpBody
}

type Sftp struct {
	Incoming chan *SftpEvent
}

func (p *Sftp) GetDisplayName() string {
	return "SFTP"
}

func (p *Sftp) GetSupportedPackages() map[protocol.PackageType]BodyFactory {
	return map[protocol.PackageType]BodyFactory{
		SftpType: func() interface{} { return &SftpBody{} },
	}
}

func (p *Sftp) Handle(device *network.Device, pkg *protocol.Package) bool {
	if pkg.Type != SftpType {
		return false
	}

	event := &SftpEvent{SftpBody: *pkg.Body.(*SftpBody)}
	event.Device = device

	select {
	case p.Incoming <- event:
	default:
	}

	return true
}

func (p *Sftp) SendStartBrowsing(device *network.Device) error {
	return device.Send(SftpType, &SftpBody{StartBrowsing: true})
}

func NewSftp() *Sftp {
	return &Sftp{
		Incoming: make(chan *SftpEvent),
	}
}
