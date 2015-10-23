package network

import (
	"net"
	"log"
)

// Client holds info about connection
type TcpClient struct {
	Conn net.Conn
	Server *TcpServer
	id int
}

// TCP server
type TcpServer struct {
	clients []*TcpClient
	address string // Address to open connection, e.g. localhost:9999
	joins chan net.Conn // Channel for new connections
	Joins chan *TcpClient
}

// Creates new Client instance and starts listening
func (s *TcpServer) newClient(conn net.Conn) {
	client := &TcpClient{
		Conn: conn,
		Server: s,
		id: len(s.clients),
	}
	s.clients = append(s.clients, client)
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
