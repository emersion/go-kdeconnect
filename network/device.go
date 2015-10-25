package network

import (
	"log"
	"net"
	"bufio"
	"crypto/rsa"
	"github.com/emersion/go-kdeconnect/netpkg"
)

type Device struct {
	conn net.Conn
	Id string
	Name string
	ProtocolVersion int
	Type string
	PublicKey *rsa.PublicKey
	Incoming chan *netpkg.Package
}

func (d *Device) send(pkg *netpkg.Package) error {
	_, err := d.conn.Write(pkg.Serialize())
	return err
}

func (d *Device) Send(t netpkg.Type, b interface{}) error {
	pkg := &netpkg.Package{
		Type: t,
		Body: b,
	}

	if d.PublicKey != nil {
		var err error
		pkg, err = pkg.Encrypt(d.PublicKey)
		if err != nil {
			return err
		}
	}

	return d.send(pkg)
}

func (d *Device) Listen() {
	defer (func() {
		d.conn.Close()
		close(d.Incoming)
	})()

	scanner := bufio.NewScanner(d.conn)
	for scanner.Scan() {
		pkg, err := netpkg.Unserialize(scanner.Bytes())
		if err != nil {
			log.Println("Cannot parse package:", err)
			break
		}

		d.Incoming <- pkg
	}
}

func NewDevice(conn net.Conn) *Device {
	return &Device{
		conn: conn,
		Incoming: make(chan *netpkg.Package),
	}
}
