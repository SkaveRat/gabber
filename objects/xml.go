package objects

import "encoding/xml"

type Stream struct {
	XMLName       xml.Name `xml:"jabber:client stream:stream"`
	XMLNameAttr   string `xml:"xmlns:stream,attr"`
	Id            string `xml:"id,attr"`
	From          string `xml:"from,attr"`
	Version       string `xml:"version,attr"`
}

type IncomingStream struct {
	XMLName       xml.Name `xml:"jabber:client stream:stream"`
	XMLNameAttr   string `xml:"xmlns:stream,attr"`
	To            string `xml:"to,attr"`
	Version       string `xml:"version,attr"`
}


type SaslAuthRequest struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string `xml:"mechanism,attr"`
	Text      string `xml:",chardata"`
}

type SaslSuccess struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

type SaslFailure struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
}

type SaslMechanisms struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism [1]string `xml:"mechanism"`
}

//auth features
type SaslFeatures struct {
	XMLName    xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl stream:features"`
	Mechanisms SaslMechanisms `xml:",omitempty"`
}

//session features
type StreamFeatures struct {
	XMLName    xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl stream:features"`
	Bind       FeatureBind
	Session    FeatureSession
}


type FeatureSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	Optional Optional
}

type FeatureBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Required Required
}

type Iq struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	From string `xml:"from,attr"`
	To string `xml:"to,attr"`
	Type string `xml:"type,attr"`
}

type Required struct {}
type Optional struct {}


type IncomingStanza struct {
	XMLName xml.Name `xml:""`
	InnerXml []byte `xml:",innerxml"`
	Id string `xml:"id,attr,omitmissing"`
}


type IqStanzaItemsQuery struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	ItemsQuery ItemsQuery `xml:""`
}


type IqStanzaRosterQuery struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string  `xml:"type,attr"`
	From string  `xml:"from,attr,omitempty"`
	To string  `xml:"to,attr,omitempty"`
	RosterQuery RosterQuery `xml:""`
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

type RosterItem struct {
	XMLName xml.Name `xml:"item"`
	Jid string `xml:"jid,attr"`
}

type ItemsQuery struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#items query"`
}

type InfoQuery struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#info query"`
}

type RosterQuery struct {
	XMLName xml.Name `xml:"jabber:iq:roster query"`
	RosterItems []RosterItem
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
