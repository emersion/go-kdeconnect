package engine

import (
	"strconv"
	"net"
	"log"
	"os"
	"github.com/emersion/go-kdeconnect/crypto"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/network"
	"github.com/emersion/go-kdeconnect/plugin"
)

type Config struct {
	UdpPort int
	TcpPort int
	DeviceId string
	DeviceName string
	DeviceType string
	PrivateKey *crypto.PrivateKey
}

func DefaultConfig() *Config {
	hostname, _ := os.Hostname()

	return &Config{
		UdpPort: 1714,
		TcpPort: 1715,
		DeviceId: hostname,
		DeviceName: hostname,
		DeviceType: "desktop",
	}
}

func setDeviceIdentity(device *network.Device, identity *netpkg.Identity) {
	device.Id = identity.DeviceId
	device.Name = identity.DeviceName
	device.Type = identity.DeviceType
	device.ProtocolVersion = identity.ProtocolVersion
}

type Engine struct {
	config *Config
	handler *plugin.Handler
	udpServer *network.UdpServer
	tcpServer *network.TcpServer
	devices map[string]*network.Device
	Joins chan *network.Device
	Paired chan *network.Device
	Leaves chan *network.Device
}

func (e *Engine) sendIdentity(device *network.Device) error {
	return device.Send(netpkg.IdentityType, &netpkg.Identity{
		DeviceId: e.config.DeviceId,
		DeviceName: e.config.DeviceName,
		ProtocolVersion: netpkg.ProtocolVersion,
		DeviceType: e.config.DeviceType,
		TcpPort: e.config.TcpPort,
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
	e.devices[device.Id] = device

	select {
	case e.Joins <- device:
	default:
	}

	go device.Listen()
	go e.sendIdentity(device)

	for pkg := range device.Incoming {
		if pkg == nil {
			continue
		}

		if pkg.Type == netpkg.EncryptedType {
			var err error
			pkg, err = pkg.Body.(*netpkg.Encrypted).Decrypt(e.config.PrivateKey)
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

			lpub, _ := e.config.PrivateKey.PublicKey().Marshal()
			device.Send(netpkg.PairType, &netpkg.Pair{
				PublicKey: string(lpub),
				Pair: true,
			})

			device.PublicKey = rpub

			select {
			case e.Paired <- device:
			default:
			}
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

	delete(e.devices, device.Id)

	select {
	case e.Leaves <- device:
	default:
	}
}

func (e *Engine) broadcastIdentity() error {
	conn, err := net.Dial("udp", "255.255.255.255:"+strconv.Itoa(e.config.UdpPort))
	if err != nil {
		return err
	}

	device := network.NewDevice(conn)
	return e.sendIdentity(device)
}

func (e *Engine) Listen() {
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
				if identity.DeviceId == e.config.DeviceId {
					// Do not try to connect with ourselves
					continue
				}
				if _, ok := e.devices[identity.DeviceId]; ok {
					// Device already known
					continue
				}

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

func New(handler *plugin.Handler, config *Config) *Engine {
	if config.PrivateKey == nil {
		log.Println("No private key specified, generating a new one...")
		privateKey, err := crypto.GeneratePrivateKey()
		if err != nil {
			log.Fatal("Could not generate private key", err)
		}
		config.PrivateKey = privateKey
	}

	return &Engine{
		config: config,
		handler: handler,
		udpServer: network.NewUdpServer(":"+strconv.Itoa(config.UdpPort)),
		tcpServer: network.NewTcpServer(":"+strconv.Itoa(config.TcpPort)),
		devices: map[string]*network.Device{},
		Joins: make(chan *network.Device),
		Paired: make(chan *network.Device),
		Leaves: make(chan *network.Device),
	}
}
