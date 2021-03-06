package network

import (
	"bufio"
	"github.com/emersion/go-kdeconnect/crypto"
	"github.com/emersion/go-kdeconnect/protocol"
	"log"
	"net"
)

type Device struct {
	conn                net.Conn
	Id                  string
	Name                string
	ProtocolVersion     int
	Type                string
	PublicKey           *crypto.PublicKey
	Paired              bool
	PairRequestReceived bool
	PairRequestSent     bool
	Incoming            chan *protocol.Package
}

func (d *Device) send(pkg *protocol.Package) error {
	_, err := d.conn.Write(pkg.Serialize())
	return err
}

func (d *Device) Send(t protocol.PackageType, b interface{}) error {
	pkg := &protocol.Package{
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
		pkg, err := protocol.Unserialize(scanner.Bytes())
		if err != nil {
			log.Println("Cannot parse package:", err)
			break
		}

		d.Incoming <- pkg
	}
}

func (d *Device) Close() error {
	return d.conn.Close()
}

func (d *Device) Addr() net.Addr {
	return d.conn.RemoteAddr()
}

func NewDevice(conn net.Conn) *Device {
	return &Device{
		conn:     conn,
		Incoming: make(chan *protocol.Package),
	}
}
