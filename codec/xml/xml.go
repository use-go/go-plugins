// Package xml provides a xml codec
package xml

import (
	"encoding/xml"
	"io"

	"github.com/micro/go-micro/codec"
)

//Codec for xml
type Codec struct {
	Conn    io.ReadWriteCloser
	Encoder *xml.Encoder
	Decoder *xml.Decoder
}

//ReadHeader return nil ,because no need of head
func (c *Codec) ReadHeader(m *codec.Message, t codec.MessageType) error {
	return nil
}

//ReadBody Get Data to handle
func (c *Codec) ReadBody(b interface{}) error {
	if b == nil {
		return nil
	}
	return c.Decoder.Decode(b)
}

func (c *Codec) Write(m *codec.Message, b interface{}) error {
	if b == nil {
		return nil
	}
	return c.Encoder.Encode(b)
}

//Close stream
func (c *Codec) Close() error {
	return c.Conn.Close()
}

func (c *Codec) String() string {
	return "json"
}

//NewCodec return xml codec
func NewCodec(c io.ReadWriteCloser) codec.Codec {
	return &Codec{
		Conn:    c,
		Decoder: xml.NewDecoder(c),
		Encoder: xml.NewEncoder(c),
	}
}
