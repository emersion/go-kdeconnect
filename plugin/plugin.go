package plugin

import (
	"errors"
	"encoding/json"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
)

type Plugin interface {
	GetDisplayName() string
	GetSupportedPackages() map[netpkg.Type]interface{}
	Handle(*network.Device, *netpkg.Package) bool
}

type Event struct {
	Device *network.Device
}

type Handler struct {
	plugins []Plugin
	registeredPackages map[netpkg.Type]interface{}
}

func (h *Handler) Register(plugin Plugin) {
	pkgs := plugin.GetSupportedPackages()
	for t, b := range pkgs {
		h.registeredPackages[t] = b
	}

	h.plugins = append(h.plugins, plugin)
}

func (h *Handler) Handle(device *network.Device, pkg *netpkg.Package) error {
	for t, b := range h.registeredPackages {
		if pkg.Type == t {
			pkg.Body = *&b
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
		registeredPackages: map[netpkg.Type]interface{}{},
	}
}
