package engine

import (
	"strconv"
	"net"
	"log"
	"os"
	"github.com/emersion/go-kdeconnect/crypto"
	"github.com/emersion/go-kdeconnect/protocol"
	"github.com/emersion/go-kdeconnect/network"
	"github.com/emersion/go-kdeconnect/plugin"
)

type KnownDevice struct {
	Id string
	Name string
	Type string
	PublicKey *crypto.PublicKey
}

func NewKnownDeviceFromDevice(device *network.Device) *KnownDevice {
	return &KnownDevice{
		Id: device.Id,
		Name: device.Name,
		Type: device.Type,
		PublicKey: device.PublicKey,
	}
}

type Config struct {
	UdpPort int
	TcpPort int
	DeviceId string
	DeviceName string
	DeviceType string
	PrivateKey *crypto.PrivateKey
	KnownDevices []*KnownDevice
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

func setDeviceIdentity(device *network.Device, identity *protocol.Identity) {
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
	RequestsPairing chan *network.Device
	Paired chan *network.Device
	Unpaired chan *network.Device
	Leaves chan *network.Device
}

func (e *Engine) sendIdentity(device *network.Device) error {
	return device.Send(protocol.IdentityType, &protocol.Identity{
		DeviceId: e.config.DeviceId,
		DeviceName: e.config.DeviceName,
		ProtocolVersion: protocol.Version,
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

func (e *Engine) PairDevice(device *network.Device) error {
	if device.Paired {
		return nil
	}

	if !device.PairRequestSent {
		pair := &protocol.Pair{
			Pair: true,
		}

		pub, err := e.config.PrivateKey.PublicKey().Marshal()
		if err != nil {
			return err
		}
		pair.PublicKey = string(pub)

		err = device.Send(protocol.PairType, pair)
		if err != nil {
			return err
		}

		log.Println("Sent pairing request to", device.Name)

		device.PairRequestSent = true
	}

	if device.PairRequestReceived && device.PairRequestSent {
		device.Paired = true

		log.Println("Device", device.Name, "paired")

		if e.getKnownDevice(device) == -1 {
			// Add it to the list of known devices
			e.config.KnownDevices = append(e.config.KnownDevices, NewKnownDeviceFromDevice(device))
		}

		select {
		case e.Paired <- device:
		default:
		}
	}

	return nil
}

func (e *Engine) UnpairDevice(device *network.Device) error {
	if device.Paired {
		err := device.Send(protocol.PairType, &protocol.Pair{
			Pair: false,
		})
		if err != nil {
			return err
		}

		device.Paired = false
	}

	select {
	case e.Unpaired <- device:
	default:
	}

	return nil
}

func (e *Engine) getKnownDevice(device *network.Device) int {
	for i, knownDevice := range e.config.KnownDevices {
		if knownDevice.Id == device.Id {
			return i
		}
	}
	return -1
}

func (e *Engine) handleDevice(device *network.Device) {
	e.devices[device.Id] = device

	if i := e.getKnownDevice(device); i != -1 {
		device.Paired = true
		//device.PublicKey = e.config.KnownDevices[i].PublicKey
	}

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

		// Decrypt package if encrypted
		if pkg.Type == protocol.EncryptedType {
			var err error
			pkg, err = pkg.Body.(*protocol.Encrypted).Decrypt(e.config.PrivateKey)
			if err != nil {
				log.Println("Cannot decrypt package:", err)
				continue
			}
		}

		if pkg.Type == protocol.PairType {
			pair := pkg.Body.(*protocol.Pair)

			if pair.Pair {
				// Remote asks pairing
				device.PairRequestReceived = true

				if len(pair.PublicKey) > 0 {
					// Remote sent its public key
					pub := &crypto.PublicKey{}

					if err := pub.Unmarshal([]byte(pair.PublicKey)); err != nil {
						log.Println("Cannot parse public key:", err)
						break
					}
					log.Println("Received public key")

					device.PublicKey = pub
				}

				if device.PairRequestSent {
					// We asked for pairing first, remote accepted
					log.Println("Device accepted pairing")

					e.PairDevice(device)
				} else {
					log.Println("Device requested pairing")

					select {
					case e.RequestsPairing <- device:
					default:
					}
				}
			} else {
				log.Println("Device requested unpairing")

				device.Paired = false
				e.UnpairDevice(device)
			}
		} else if pkg.Type == protocol.IdentityType {
			setDeviceIdentity(device, pkg.Body.(*protocol.Identity))
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
	err = e.sendIdentity(device)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
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

			if pkg.Type == protocol.IdentityType {
				identity := pkg.Body.(*protocol.Identity)
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
					log.Println("Could not open a TCP connection:", err)
					continue
				}

				setDeviceIdentity(device, identity)

				go e.handleDevice(device)
			} else {
				log.Println("Received a non-identity package on UDP connection", pkg)
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
		priv := &crypto.PrivateKey{}
		if err := priv.Generate(); err != nil {
			log.Fatal("Could not generate private key", err)
		}
		config.PrivateKey = priv
	}

	return &Engine{
		config: config,
		handler: handler,
		udpServer: network.NewUdpServer(":"+strconv.Itoa(config.UdpPort)),
		tcpServer: network.NewTcpServer(":"+strconv.Itoa(config.TcpPort)),
		devices: map[string]*network.Device{},
		Joins: make(chan *network.Device),
		RequestsPairing: make(chan *network.Device),
		Paired: make(chan *network.Device),
		Unpaired: make(chan *network.Device),
		Leaves: make(chan *network.Device),
	}
}
