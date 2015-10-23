package engine

import (
	"strconv"
	"net"
	"log"
	"crypto/rsa"
	"github.com/emersion/go-kdeconnect/crypto"
	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/server"
)

const udpPort = 1714
const tcpPort = 1715
const protocolVersion = 5

type Engine struct {
	udpServer *server.UdpServer
	tcpServer *server.TcpServer
	privateKey *rsa.PrivateKey
}

func (e *Engine) sendIdentity(conn net.Conn) error {
	packet := &netpkg.Package{
		Type: netpkg.IdentityType,
		Body: &netpkg.Identity{
			DeviceId: "go",
			DeviceName: "go",
			ProtocolVersion: protocolVersion,
			DeviceType: "desktop",
			TcpPort: tcpPort,
		},
	}
	_, err := conn.Write(packet.Serialize())
	return err
}

func (e *Engine) connect(addr *net.TCPAddr) (*server.TcpClient, error) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	log.Println("New outgoing TCP connection")

	return server.NewTcpClient(conn), nil
}

func (e *Engine) handleConnection(client *server.TcpClient) {
	go client.Listen()
	go e.sendIdentity(client.Conn)

	for {
		pkg := <-client.Incoming

		if pkg == nil {
			log.Println("Received null package")
			//client.Close()
			continue
		}

		if pkg.Type == netpkg.EncryptedType {
			// Decrypt package first
			var err error
			pkg, err = pkg.Body.(*netpkg.Encrypted).Decrypt(e.privateKey)
			if err != nil {
				log.Println("Cannot decrypt package:", err)
				continue
			}
		}

		switch pkg.Type {
		case netpkg.PairType:
			pair := pkg.Body.(*netpkg.Pair)
			pubkey, err := crypto.UnmarshalPublicKey([]byte(pair.PublicKey))
			if err != nil {
				log.Println("Cannot parse public key:", err)
				break
			}
			log.Println("Received public key:", pubkey)

			privkeyBin, _ := crypto.MarshalPublicKey(&e.privateKey.PublicKey)
			packet := &netpkg.Package{
				Type: netpkg.PairType,
				Body: &netpkg.Pair{
					PublicKey: string(privkeyBin),
					Pair: true,
				},
			}
			client.Conn.Write(packet.Serialize())
		default:
			log.Println("Unknown package type:", pkg.Type, string(pkg.RawBody))
		}
	}

	log.Println("Closed TCP connection")
}

func (e *Engine) broadcastIdentity() error {
	conn, err := net.Dial("udp", "255.255.255.255:"+strconv.Itoa(udpPort))
	if err != nil {
		return err
	}

	return e.sendIdentity(conn)
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

				client, err := e.connect(&net.TCPAddr{
					IP: raddr.IP,
					Port: identity.TcpPort,
					Zone: raddr.Zone,
				})
				if err != nil {
					log.Println("Could open a TCP connection:", err)
					continue
				}

				go e.handleConnection(client)
			} else {
				log.Println(pkg)
			}
		case client := <- e.tcpServer.Joins:
			log.Println("New incoming TCP connection")
			go e.handleConnection(client)
		}
	}
}

func New() *Engine {
	return &Engine{
		udpServer: server.NewUdpServer(":"+strconv.Itoa(udpPort)),
		tcpServer: server.NewTcpServer(":"+strconv.Itoa(tcpPort)),
	}
}
