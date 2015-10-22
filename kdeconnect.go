package main

import (
	"net"
	"log"
	"strconv"
	"bufio"

	"github.com/emersion/go-kdeconnect/netpkg"
	"github.com/emersion/go-kdeconnect/crypto"
)

const port = 1714
const protocolVersion = 5

func connect(addr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	go (func () {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			pkg, err := netpkg.Unserialize(scanner.Bytes())
			if err != nil {
				log.Fatal(err)
			}

			if pkg.Type == netpkg.PairType {
				pair := pkg.Body.(*netpkg.Pair)
				pubkey, _ := crypto.UnmarshalPublicKey([]byte(pair.PublicKey))
				log.Println("Received public key:", pubkey)

				privkey, _ := crypto.GenerateKey()
				privkeyBin, _ := crypto.MarshalPublicKey(&privkey.PublicKey)
				packet := &netpkg.Package{
					Type: netpkg.PairType,
					Body: &netpkg.Pair{
						PublicKey: string(privkeyBin),
						Pair: true,
					},
				}
				conn.Write(packet.Serialize())
			} else {
				log.Println("Unknown message:", pkg.Body)
			}
		}
	})()

	packet := &netpkg.Package{
		Type: netpkg.IdentityType,
		Body: &netpkg.Identity{
			DeviceId: "go",
			DeviceName: "go",
			ProtocolVersion: protocolVersion,
			DeviceType: "desktop",
		},
	}
	conn.Write(packet.Serialize())
}

func listen() {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err.Error())
	}
	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err.Error())
	}

	for {
		buffer := make([]byte, 2048) // TODO
		n, raddr, err := ln.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		pkg, err := netpkg.Unserialize(buffer[0:n])
		if err != nil {
			log.Fatal(err)
		}

		if pkg.Type == netpkg.IdentityType {
			identity := pkg.Body.(*netpkg.Identity)
			log.Println("New device:", identity)

			go connect(&net.TCPAddr{
				IP: raddr.IP,
				Port: identity.TcpPort,
				Zone: raddr.Zone,
			})
		} else {
			log.Println(pkg)
		}
	}
}

func broadcast() {
	con, err := net.Dial("udp", "255.255.255.255:"+strconv.Itoa(port))
	if err != nil {
		panic(err.Error())
	}

	packet := &netpkg.Package{
		Type: "kdeconnect.identity",
		Body: &netpkg.Identity{
			DeviceId: "go",
			DeviceName: "go",
			ProtocolVersion: protocolVersion,
			DeviceType: "desktop",
		},
	}

	_, err = con.Write(packet.Serialize())
	if err != nil {
		log.Println(err)
	}
}

func main() {
	//go broadcast()
	listen()
}
