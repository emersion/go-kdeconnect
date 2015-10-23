package engine

import (
	"crypto/rsa"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
)

type Device struct {
	Remote *network.TcpClient
	PublicKey *rsa.PublicKey
}

func (d *Device) WritePackage(pkg *netpkg.Package) error {
	return nil
}

func (d *Device) Listen() {}
