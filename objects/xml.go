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

type Required struct {}
type Optional struct {}
