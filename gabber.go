package main

import (
	"net"
	"github.com/skaverat/gabber/connection"
)


func main() {

	listener, _ := net.Listen("tcp", "0.0.0.0:5222")

	for {
		conn, _ := listener.Accept()

		connChan := make(chan net.Conn)
		go connection.Run(connChan)
		connChan<-conn
	}
}
