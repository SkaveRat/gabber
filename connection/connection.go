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
	connCloseChannel := make(chan bool)
	iqChannel 		   := make(chan []byte)
	answerChannel      := make(chan []byte)

	go this.handleAnswerConnection(answerChannel); //outgoing stream
	go this.handleIncoming(authRequestChannel, streamStartChannel, iqChannel, connCloseChannel); //incoming stream

	Mainloop:
	for {
		select {
		case _ = <-authRequestChannel:
			this.isAuthed = true
			//TODO check credentials
			success,_ := xml.Marshal(objects.SaslSuccess{})
			answerChannel <- success
			answerChannel <- []byte("</stream:stream>")
		case _ = <-streamStartChannel:
			answerChannel <- getStreamBegin()
			if(!this.isAuthed) {
				answerChannel <- getAuthRequest()
			}else{
				answerChannel <- getStreamFeatures()
			}
		case iqData := <-iqChannel:
			answerChannel <- iqData
		case _ = <-connCloseChannel:
			answerChannel <- []byte("</stream:stream>")
			break Mainloop
		}
	}
	this.conn.Close()
}

func (this *Connection) handleAnswerConnection(answerChan chan []byte) {
	for answer := range answerChan {
		fmt.Fprint(this.conn, string(answer))
	}
}

func (this *Connection) handleIncoming(authRequestChannel chan bool, incomingStreamChannel chan bool, iqChannel chan []byte, connCloseChannel chan bool) {
	connection := util.Tee{this.conn, os.Stdout}
	decoder := xml.NewDecoder(connection);
	decoder.Strict = false;

	var isWaitingForStream bool = true

	for { //TODO: case condition when switching between stream- and stanza loop?
		if(isWaitingForStream) {
			Streamloop:
			for {
				token, _ := decoder.Token()
				switch tokenType := token.(type) {
				case xml.StartElement:
					var elmt xml.StartElement = xml.StartElement(tokenType)
					name := elmt.Name.Local
					switch name {
					case "stream":
						incomingStreamChannel<-true
						isWaitingForStream = false
						break Streamloop
					}
				case xml.ProcInst:
					//XML header
				}
			}
		}else {
			Stanzaloop:
			for {
				var stanza IncomingStanza = IncomingStanza{}
				err := decoder.Decode(&stanza)
				if (err != nil) {
					connCloseChannel <- true
				}

				switch stanza.XMLName.Local {
				case "auth":
					isWaitingForStream = true
					authRequestChannel <- true
					break Stanzaloop
				case "iq":
					bind := Bind{}
					session := Session{}
					infoquery := InfoQuery{}
					itemsquery := ItemsQuery{}
					if nil == xml.Unmarshal(stanza.InnerXml, &bind) {
						iqChannel<-getBindResponse(stanza.Id)
					}else if nil == xml.Unmarshal(stanza.InnerXml, &session) {
						iqChannel<-getSessionResponse(stanza.Id)
					}else if nil == xml.Unmarshal(stanza.InnerXml, &infoquery) {
						iqChannel<-getInfoQueryResponse(stanza.Id)
					}else if nil == xml.Unmarshal(stanza.InnerXml, &itemsquery) {
						iqChannel<-getItemsQueryResponse(stanza.Id)
					}
				}
			}
		}
	}
}

type IncomingStanza struct {
	XMLName xml.Name `xml:""`
	InnerXml []byte `xml:",innerxml"`
	Id string `xml:"id,attr,omitmissing"`
}

func getInfoQueryResponse(id string) []byte {
	iq := IqStanzaInfoQuery{
		Id: id,
		Type: "result",
		From: "localhost",
		InfoQuery: InfoQuery{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getItemsQueryResponse(id string) []byte {
	iq := IqStanzaItemsQuery{
		Id: id,
		Type: "result",
		From: "localhost",
		ItemsQuery: ItemsQuery{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getBindResponse(id string) []byte {
	bind := Bind{}
	bind.Jid = Jid{Value:"xabber@localhost/foobar"}

	iq := IqStanzaBind{
		Id: id,
		Type: "result",
		Bind: bind,
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getSessionResponse(id string) []byte {

	iq := IqStanzaSession{
		Id: id,
		Type: "result",
		Session: Session{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

type IqStanzaItemsQuery struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	ItemsQuery ItemsQuery `xml:""`
}

type IqStanzaInfoQuery struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	InfoQuery InfoQuery `xml:""`
}

type IqStanzaBind struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	Bind Bind  `xml:""`
}

type IqStanzaSession struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	Session Session  `xml:""`
}

type ItemsQuery struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#items query"`
}

type InfoQuery struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#info query"`
}


type Session struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
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
