package engine

import (
	"crypto/rsa"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/server"
)

type Device struct {
	Remote *server.TcpClient
	PublicKey *rsa.PublicKey
}

func (d *Device) WritePackage(pkg *netpkg.Package) error {
	return nil
}
