package server

import (
	"bufio"
	"net"
	"log"
	"github.com/emersion/go-kdeconnect/netpkg"
)

// Client holds info about connection
type TcpClient struct {
	Conn net.Conn
	Server *TcpServer
	Incoming chan *netpkg.Package // Channel for incoming data from client
	id int
}

// TCP server
type TcpServer struct {
	clients []*TcpClient
	address string // Address to open connection, e.g. localhost:9999
	joins chan net.Conn // Channel for new connections
	Joins chan *TcpClient
}

// Read client data from channel
func (c *TcpClient) Listen() {
	defer c.Close()

	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		pkg, err := netpkg.Unserialize(scanner.Bytes())
		if err != nil {
			log.Fatal("Cannot parse package:", err)
		}

		c.Incoming <- pkg
	}
}

func (c *TcpClient) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}

	if c.Server != nil {
		c.Server.clients[c.id] = nil
	}

	return nil
}

// Creates new Client instance and starts listening
func (s *TcpServer) newClient(conn net.Conn) {
	client := &TcpClient{
		Conn: conn,
		Server: s,
		id: len(s.clients),
		Incoming: make(chan *netpkg.Package),
	}
	s.clients = append(s.clients, client)
	go client.Listen()
	s.Joins <- client
}

// Listens new connections channel and creating new client
func (s *TcpServer) listenChannels() {
	for {
		select {
		case conn := <-s.joins:
			s.newClient(conn)
		}
	}
}

// Start network server
func (s *TcpServer) Listen() {
	log.Println("Creating TCP server with address", s.address)

	go s.listenChannels()

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}

	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		s.joins <- conn
	}
}

// Creates new tcp server instance
func NewTcpServer(address string) *TcpServer {
	return &TcpServer{
		address: address,
		joins: make(chan net.Conn),
		Joins: make(chan *TcpClient),
	}
}

func NewTcpClient(conn net.Conn) *TcpClient {
	return &TcpClient{
		Conn: conn,
		Incoming: make(chan *netpkg.Package),
	}
}
