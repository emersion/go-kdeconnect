package plugin

import (
	"errors"
	"encoding/json"
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
)

type BodyFactory func() interface{}

type Plugin interface {
	GetDisplayName() string
	GetSupportedPackages() map[protocol.PackageType]BodyFactory
	Handle(*network.Device, *protocol.Package) bool
}

type Event struct {
	Device *network.Device
}

type Handler struct {
	plugins []Plugin
	registeredPackages map[protocol.PackageType]BodyFactory
}

func (h *Handler) Register(plugin Plugin) {
	pkgs := plugin.GetSupportedPackages()
	if pkgs != nil {
		for t, factory := range pkgs {
			h.registeredPackages[t] = factory
		}
	}

	h.plugins = append(h.plugins, plugin)
}

func (h *Handler) Handle(device *network.Device, pkg *protocol.Package) error {
	for t, factory := range h.registeredPackages {
		if pkg.Type == t {
			pkg.Body = factory()
			break
		}
	}

	if pkg.Body != nil {
		err := json.Unmarshal(pkg.RawBody, pkg.Body)
		if err != nil {
			return err
		}
	}

	for _, plugin := range h.plugins {
		if plugin.Handle(device, pkg) {
			return nil
		}
	}

	return errors.New("Unknown message type")
}

func NewHandler() *Handler {
	return &Handler{
		registeredPackages: map[protocol.PackageType]BodyFactory{},
	}
}
