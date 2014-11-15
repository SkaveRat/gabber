package main

import (
	"net"
	"io"
	"os"
	"encoding/xml"
	"fmt"
	"bytes"
	"github.com/skaverat/gabber/objects"
)


func main() {

	listener, _ := net.Listen("tcp", "0.0.0.0:5222")

	for {
		conn, _ := listener.Accept()

		connChan := make(chan net.Conn)
		go handleIncoming(connChan)
		connChan<-conn
	}
}

func handleIncoming(connChan chan net.Conn) {
	conn := <-connChan

	authRequestChannel := make(chan bool)
	streamStartChannel := make(chan bool)
	go handleConnection(authRequestChannel, streamStartChannel, conn);

	for {
		select {
		case _ = <-authRequestChannel:
			fmt.Println("auth incoming")
//			success,_ := xml.Marshal(saslSuccess{})
//			answer(conn, success)
//			answer(conn, []byte("</stream:stream>"))
		case _ = <-streamStartChannel:
			fmt.Println("Incoming Stream")
			answer(conn, getStreamBegin())
			sendAuthRequest(conn)
		}
	}
}

func handleConnection(authRequestChannel chan bool, incomingStreamChannel chan bool, conn net.Conn) {
	connection := tee{conn, os.Stdout}
	decoder := xml.NewDecoder(connection);
	decoder.Strict = false;

	for {
		token, _ := decoder.RawToken()
		switch tokenType := token.(type) {
		case xml.StartElement:
			elmt := xml.StartElement(tokenType)
			name := elmt.Name.Local
			switch name {
			case "stream":
			incomingStreamChannel<-true
			case "auth":
			authRequestChannel<-true
			}
		case xml.ProcInst:
			fmt.Println("xml header")
		}
	}

}

func sendAuthRequest(conn net.Conn) {
		var plainMechanism objects.SaslMechanisms = objects.SaslMechanisms{}
		plainMechanism.Mechanism[0] = "PLAIN"
		features := objects.SaslFeatures{}
		features.Mechanisms = plainMechanism
		featuresBytes, _ := xml.Marshal(features)

		answer(conn, featuresBytes)
}

func getStreamBegin() []byte {
	stream := objects.Stream{}
	stream.XMLNameAttr = "http://etherx.jabber.org/streams"
	stream.Id = "kjsandkjbfhjbdsjfbsdf"
	stream.From = "localhost"
	stream.Version = "1.0"
	streamBytes, _ := xml.Marshal(stream)

	streamArray := bytes.Split(streamBytes, []byte("</"))
	return streamArray[0]
}


func answer(conn net.Conn, data []byte) {
	fmt.Fprint(conn, string(data))
}

type tee struct {
	r io.Reader
	w io.Writer
}

func (t tee) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		t.w.Write(p[0:n])
		t.w.Write([]byte("\n"))
	}
	return
}
