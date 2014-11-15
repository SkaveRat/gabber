package connection

import (
	"net"
	"fmt"
	"encoding/xml"
	"os"
	"github.com/skaverat/gabber/objects"
	"bytes"
	"github.com/skaverat/gabber/util"
)

type Connection struct {
	conn net.Conn
	isAuthed bool
}

func Run(connChan chan net.Conn) {
	c := Connection{isAuthed: false}
	c.Create(connChan)
}

func (this *Connection) Create(connChan chan net.Conn) {
	this.conn = <-connChan;

	authRequestChannel := make(chan bool)
	streamStartChannel := make(chan bool)
	go handleConnection(authRequestChannel, streamStartChannel, this.conn);

	for {
		select {
		case _ = <-authRequestChannel:
			fmt.Println("auth incoming")
			this.isAuthed = true
			success,_ := xml.Marshal(objects.SaslSuccess{})
			answer(this.conn, success)
			answer(this.conn, []byte("</stream:stream>"))
		case _ = <-streamStartChannel:
			fmt.Println("Incoming Stream")
			answer(this.conn, getStreamBegin())
			if(!this.isAuthed) {
				sendAuthRequest(this.conn)
			}else{
				fmt.Println("authed")
			}
		}
	}
}

func handleConnection(authRequestChannel chan bool, incomingStreamChannel chan bool, conn net.Conn) {
	connection := util.Tee{conn, os.Stdout}
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
