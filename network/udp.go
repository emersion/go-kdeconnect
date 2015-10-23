package network

import (
	"net"
	"log"
	"github.com/emersion/go-kdeconnect/netpkg"
	"bytes"
)

type UdpServer struct {
	address string
	Incoming chan UdpPackage
}

type UdpPackage struct {
	RemoteAddress *net.UDPAddr
	Package *netpkg.Package
}

func (s *UdpServer) Listen() {
	log.Println("Creating UDP server with address", s.address)

	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		log.Fatal("Cannot resolve UDP address:", err)
	}
	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("Cannot start UDP server:", err)
	}

	defer ln.Close()

	for {
		buffer := make([]byte, 2048) // TODO
		n, raddr, err := ln.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("Cannot read from remote:", err)
		}

		pkgBin := bytes.Trim(buffer[0:n], "\n")
		pkg, err := netpkg.Unserialize(pkgBin)
		if err != nil {
			log.Fatal("Cannot parse package:", err)
		}

		s.Incoming <- UdpPackage{RemoteAddress: raddr, Package: pkg}
	}
}

func NewUdpServer(addr string) *UdpServer {
	return &UdpServer{address: addr, Incoming: make(chan UdpPackage)}
}
