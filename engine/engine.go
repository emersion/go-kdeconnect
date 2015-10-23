package engine

import (
	"strconv"
	"net"
	"log"
	"crypto/rsa"
	"github.com/emersion/go-kdeconnect/crypto"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
	"github.com/emersion/go-kdeconnect/plugin"
)

const udpPort = 1714
const tcpPort = 1715
const protocolVersion = 5

func setDeviceIdentity(device *network.Device, identity *netpkg.Identity) {
	device.Id = identity.DeviceId
	device.Name = identity.DeviceName
	device.Type = identity.DeviceType
	device.ProtocolVersion = identity.ProtocolVersion
}

type Engine struct {
	udpServer *network.UdpServer
	tcpServer *network.TcpServer
	privateKey *rsa.PrivateKey
	handler *plugin.Handler
}

func (e *Engine) sendIdentity(device *network.Device) error {
	return device.Send(netpkg.IdentityType, &netpkg.Identity{
		DeviceId: "go",
		DeviceName: "go",
		ProtocolVersion: protocolVersion,
		DeviceType: "desktop",
		TcpPort: tcpPort,
	})
}

func (e *Engine) connect(addr *net.TCPAddr) (*network.Device, error) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	log.Println("New outgoing TCP connection")

	return network.NewDevice(conn), nil
}

func (e *Engine) handleDevice(device *network.Device) {
	go device.Listen()
	go e.sendIdentity(device)

	for {
		pkg := <-device.Incoming
		if pkg == nil {
			continue
		}

		if pkg.Type == netpkg.EncryptedType {
			var err error
			pkg, err = pkg.Body.(*netpkg.Encrypted).Decrypt(e.privateKey)
			if err != nil {
				log.Println("Cannot decrypt package:", err)
				continue
			}
		}

		if pkg.Type == netpkg.PairType {
			pair := pkg.Body.(*netpkg.Pair)
			rpub, err := crypto.UnmarshalPublicKey([]byte(pair.PublicKey))
			if err != nil {
				log.Println("Cannot parse public key:", err)
				break
			}
			log.Println("Received public key")

			lpub, _ := crypto.MarshalPublicKey(&e.privateKey.PublicKey)
			device.Send(netpkg.PairType, &netpkg.Pair{
				PublicKey: string(lpub),
				Pair: true,
			})

			device.PublicKey = rpub
		} else if pkg.Type == netpkg.IdentityType {
			setDeviceIdentity(device, pkg.Body.(*netpkg.Identity))
		} else {
			err := e.handler.Handle(device, pkg)
			if err != nil {
				log.Println("Error handling package:", err, pkg.Type, string(pkg.RawBody))
			}
		}
	}

	log.Println("Closed TCP connection")
}

func (e *Engine) broadcastIdentity() error {
	conn, err := net.Dial("udp", "255.255.255.255:"+strconv.Itoa(udpPort))
	if err != nil {
		return err
	}

	device := network.NewDevice(conn)
	return e.sendIdentity(device)
}

func (e *Engine) Listen() {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Could not generate private key", err)
	}
	e.privateKey = privateKey

	go e.udpServer.Listen()
	go e.tcpServer.Listen()
	go e.broadcastIdentity()

	for {
		select {
		case udpPkg := <-e.udpServer.Incoming:
			raddr := udpPkg.RemoteAddress
			pkg := udpPkg.Package

			if pkg.Type == netpkg.IdentityType {
				identity := pkg.Body.(*netpkg.Identity)
				log.Println("New device discovered by UDP:", identity)

				device, err := e.connect(&net.TCPAddr{
					IP: raddr.IP,
					Port: identity.TcpPort,
					Zone: raddr.Zone,
				})
				if err != nil {
					log.Println("Could open a TCP connection:", err)
					continue
				}

				setDeviceIdentity(device, identity)

				go e.handleDevice(device)
			} else {
				log.Println(pkg)
			}
		case client := <- e.tcpServer.Joins:
			log.Println("New incoming TCP connection")
			device := network.NewDevice(client.Conn)
			go e.handleDevice(device)
		}
	}
}

func New(handler *plugin.Handler) *Engine {
	return &Engine{
		udpServer: network.NewUdpServer(":"+strconv.Itoa(udpPort)),
		tcpServer: network.NewTcpServer(":"+strconv.Itoa(tcpPort)),
		handler: handler,
	}
}
