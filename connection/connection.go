package connection

import (
	"code.google.com/p/go.crypto/pbkdf2"
	"net"
	"fmt"
	"encoding/xml"
	"os"
	"github.com/skaverat/gabber/objects"
	"bytes"
	"github.com/skaverat/gabber/util"
	"crypto/sha1"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Connection struct {
	conn net.Conn
	db *sql.DB
	isAuthed bool
}

func Run(connChan chan net.Conn, dbChan chan *sql.DB) {
	c := Connection{isAuthed: false}
	c.Create(connChan, dbChan)
}

func (c *Connection) Create(connChan chan net.Conn, dbChan chan *sql.DB) {
	c.conn = <-connChan;
	c.db = <-dbChan

	authRequestChannel := make(chan objects.AuthCredentials)
	streamStartChannel := make(chan bool)
	connCloseChannel := make(chan bool)
	iqChannel 		   := make(chan objects.IncomingStanza)
	answerChannel      := make(chan []byte)

	go c.handleAnswerConnection(answerChannel); //outgoing stream
	go c.handleIncoming(authRequestChannel, streamStartChannel, iqChannel, connCloseChannel); //incoming stream

	Mainloop:
	for {
		select {
		case request := <-authRequestChannel:
			c.isAuthed = true
			if(c.validLogin(request)) {
				success,_ := xml.Marshal(objects.SaslSuccess{})
				answerChannel <- success
			}else{
				fail,_ := xml.Marshal(objects.SaslFailure{})
				answerChannel <- fail
			}
			answerChannel <- []byte("</stream:stream>")
		case _ = <-streamStartChannel:
			answerChannel <- getStreamBegin()
			if(!c.isAuthed) {
				answerChannel <- getAuthRequest()
			}else{
				answerChannel <- getStreamFeatures()
			}
		case stanza := <-iqChannel:
			bind       := objects.Bind{}
			session    := objects.Session{}
			infoquery  := objects.InfoQuery{}
			itemsquery := objects.ItemsQuery{}
			rosterquery := objects.RosterQuery{}
			var response []byte
			if nil == xml.Unmarshal(stanza.InnerXml, &bind) {
				response = getBindResponse(stanza.Id)
			}else if nil == xml.Unmarshal(stanza.InnerXml, &session) {
				response = getSessionResponse(stanza.Id)
			}else if nil == xml.Unmarshal(stanza.InnerXml, &infoquery) {
				response = getInfoQueryResponse(stanza.Id)
			}else if nil == xml.Unmarshal(stanza.InnerXml, &itemsquery) {
				response = getItemsQueryResponse(stanza.Id)
			}else if nil == xml.Unmarshal(stanza.InnerXml, &rosterquery) {
				response = getRosterQueryResponse(stanza.Id)
			}
			answerChannel <- response
		case _ = <-connCloseChannel:
			answerChannel <- []byte("</stream:stream>")
			break Mainloop
		}
	}
	c.conn.Close()
}

func (this *Connection) handleAnswerConnection(answerChan chan []byte) {
	for answer := range answerChan {
		fmt.Fprint(this.conn, string(answer))
	}
}

func (this *Connection) handleIncoming(authRequestChannel chan objects.AuthCredentials, incomingStreamChannel chan bool, iqChannel chan objects.IncomingStanza, connCloseChannel chan bool) {
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
				var stanza objects.IncomingStanza = objects.IncomingStanza{}
				err := decoder.Decode(&stanza)
				if (err != nil) {
					connCloseChannel <- true
				}

				switch stanza.XMLName.Local {
				case "auth":
					isWaitingForStream = true
					authrequest := objects.AuthCredentials{stanza.InnerXml}
					authRequestChannel <- authrequest
					break Stanzaloop //TODO don't switch loops when not authed
				case "iq":
					iqChannel<-stanza
				}
			}
		}
	}
}


func (c *Connection) validLogin(details objects.AuthCredentials) bool {
	var (
		userpass []byte
		usersalt []byte
	)

	tx,_ := c.db.Begin()
	defer tx.Commit()

	stmt := tx.QueryRow("SELECT password, salt FROM users WHERE username = ?", details.Username());
	stmt.Scan(&userpass, &usersalt)

	hash := pbkdf2.Key(details.Password(), usersalt, 65536, sha1.Size, sha1.New)
	return fmt.Sprintf("%s", userpass) == fmt.Sprintf("%x", hash)
}

func getInfoQueryResponse(id string) []byte {
	iq := objects.IqStanzaInfoQuery{
		Id: id,
		Type: "result",
		From: "localhost",
		InfoQuery: objects.InfoQuery{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getRosterQueryResponse(id string) []byte {
	roster := objects.RosterQuery{}
	roster.RosterItems = append(roster.RosterItems, objects.RosterItem{Jid: "foobar@localhost"})
	iq := objects.IqStanzaRosterQuery{
		Id: id,
		Type: "result",
		To: "xabber@localhost/foobar",
		RosterQuery: roster,
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getItemsQueryResponse(id string) []byte {
	iq := objects.IqStanzaItemsQuery{
		Id: id,
		Type: "result",
		From: "localhost",
		ItemsQuery: objects.ItemsQuery{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getBindResponse(id string) []byte {
	bind := objects.Bind{}
	bind.Jid = objects.Jid{Value:"xabber@localhost/foobar"}

	iq := objects.IqStanzaBind{
		Id: id,
		Type: "result",
		Bind: bind,
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
}

func getSessionResponse(id string) []byte {

	iq := objects.IqStanzaSession{
		Id: id,
		Type: "result",
		Session: objects.Session{},
	}
	iqBytes,_ := xml.Marshal(iq)
	return iqBytes
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
