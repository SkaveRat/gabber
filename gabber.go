package main

import (
	"net"
	"io"
	"os"
	"encoding/xml"
	"fmt"
	"bytes"
)

type stream struct {
	XMLName       xml.Name `xml:"jabber:client stream:stream"`
	XMLNameAttr   string `xml:"xmlns:stream,attr"`
	Id            string `xml:"id,attr"`
	From          string `xml:"from,attr"`
	Version       string `xml:"version,attr"`
}

type incomingStream struct {
	XMLName       xml.Name `xml:"jabber:client stream:stream"`
	XMLNameAttr   string `xml:"xmlns:stream,attr"`
	To            string `xml:"to,attr"`
	Version       string `xml:"version,attr"`
}


type saslAuthRequest struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string `xml:"mechanism,attr"`
	Text      string `xml:",chardata"`
}

type saslSuccess struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}


type saslMechanisms struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism [1]string `xml:"mechanism"`
}

//auth features
type saslFeatures struct {
	XMLName    xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl stream:features"`
	Mechanisms saslMechanisms `xml:",omitempty"`
}

//session features
type streamFeatures struct {
	XMLName    xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl stream:features"`
	Bind       featureBind
	Session    featureSession
}


type featureSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	Optional optional
}

type featureBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Required required
}

type required struct {}
type optional struct {}


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
		var plainMechanism saslMechanisms = saslMechanisms{}
		plainMechanism.Mechanism[0] = "PLAIN"
		features := saslFeatures{}
		features.Mechanisms = plainMechanism
		featuresBytes, _ := xml.Marshal(features)

		answer(conn, featuresBytes)
}

func getStreamBegin() []byte {
	stream := stream{}
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
