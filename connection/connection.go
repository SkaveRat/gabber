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
	iqChannel 		   := make(chan IqStanza)
	answerChannel      := make(chan []byte)

	go this.handleAnswerConnection(answerChannel); //outgoing stream
	go this.handleIncoming(authRequestChannel, streamStartChannel, iqChannel); //incoming stream

	for {
		select {
		case _ = <-authRequestChannel:
			fmt.Println("auth incoming")
			this.isAuthed = true
			//TODO check credentials
			success,_ := xml.Marshal(objects.SaslSuccess{})
			answerChannel <- success
			answerChannel <- []byte("</stream:stream>")
		case _ = <-streamStartChannel:
			fmt.Println("Incoming Stream")
			answerChannel <- getStreamBegin()
			if(!this.isAuthed) {
				answerChannel <- getAuthRequest()
			}else{
				answerChannel <- getStreamFeatures()
			}
		case iqToken := <-iqChannel:
			value,_ := xml.Marshal(iqToken)
			answerChannel <- value
		}
	}
}

func (this *Connection) handleAnswerConnection(answerChan chan []byte) {
	for answer := range answerChan {
		fmt.Fprint(this.conn, string(answer))
	}
}

func (this *Connection) handleIncoming(authRequestChannel chan bool, incomingStreamChannel chan bool, iqChannel chan IqStanza) {
	connection := util.Tee{this.conn, os.Stdout}
	decoder := xml.NewDecoder(connection);
	decoder.Strict = false;

	for {
		token, _ := decoder.Token()
		switch tokenType := token.(type) {
		case xml.StartElement:
			var elmt xml.StartElement = xml.StartElement(tokenType)
			name := elmt.Name.Local
			switch name {
			case "stream":
				incomingStreamChannel<-true
			case "auth":
				authRequestChannel<-true
			case "iq":
				iq := IqStanza{}

				for _,attr := range elmt.Attr {
					switch attr.Name.Local{
					case "id":
						iq.Id = attr.Value
					case "type":
						iq.Type = attr.Value
					}
				}

				bind := Bind{}
				decoder.Decode(&bind)
				iq.Bind = bind

				iqResponse := getBindResponse(iq.Id)
				iqChannel <- iqResponse
			}
		case xml.ProcInst:
			//XML header
		}
	}

}

func getBindResponse(id string) IqStanza {

	bind := Bind{}
	bind.Jid = Jid{Value:"xabber@localhost/foobar"}

	iq := IqStanza{
		Id: id,
		Type: "result",
		Bind: bind,
	}

	return iq
}

type IqStanza struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	Bind Bind  `xml:""`
}

type Bind struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource Resource `xml:resource,chardata,omitempty`
	Jid Jid `xml:resource,chardata`
}

type Jid struct {
	XMLName xml.Name `xml:"jid"`
	Value string `xml:",chardata"`
}

type Resource struct {
	XMLName xml.Name `xml:"resource"`
	Value string `xml:",chardata"`
}

func getBindConfirm() []byte {
	var features objects.StreamFeatures = objects.StreamFeatures{}
	featuresBytes,_ := xml.Marshal(features);

	return featuresBytes
}

func getStreamFeatures() []byte {
	var features objects.StreamFeatures = objects.StreamFeatures{}
	featuresBytes,_ := xml.Marshal(features);

	return featuresBytes
}

func getAuthRequest() []byte {
	var plainMechanism objects.SaslMechanisms = objects.SaslMechanisms{}
	plainMechanism.Mechanism[0] = "PLAIN"
	features := objects.SaslFeatures{}
	features.Mechanisms = plainMechanism
	featuresBytes, _ := xml.Marshal(features)

	return featuresBytes
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
