package main

import (
	"net"
	"log"
	"bufio"
	"strconv"

	//"github.com/emersion/go-kdeconnect/netpkg"
)

const port = 1715
const protocolVersion = 5

func handleConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}
}

func listen() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func main() {
	listen()
}
