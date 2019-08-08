package xmlrpc

import (
	"bytes"
	"fmt"
	"io"

	"encoding/xml"

	"github.com/micro/go-micro/codec"
)

type xmlCodec struct {
	buf *bytes.Buffer
	mt  codec.MessageType
	rwc io.ReadWriteCloser

	c *clientCodec
	s *serverCodec
}

func (x *xmlCodec) Close() error {
	x.buf.Reset()
	return x.rwc.Close()
}

func (x *xmlCodec) String() string {
	return "xml-rpc"
}

func (x *xmlCodec) Write(m *codec.Message, body interface{}) error {
	switch m.Type {
	case codec.Request:
		return x.c.Write(m, body)
	case codec.Response:
		return x.s.Write(m, body)
	case codec.Event:
		data, err := xml.Marshal(body)
		if err != nil {
			return err
		}
		_, err = x.rwc.Write(data)
		return err
	default:
		return fmt.Errorf("Unrecognised message type: %v", m.Type)
	}
}

func (x *xmlCodec) ReadHeader(m *codec.Message, mt codec.MessageType) error {
	x.buf.Reset()
	x.mt = mt

	switch mt {
	case codec.Request:
		return x.s.ReadHeader(m)
	case codec.Response:
		return x.c.ReadHeader(m)
	case codec.Event:
		io.Copy(x.buf, x.rwc)
	default:
		return fmt.Errorf("Unrecognised message type: %v", mt)
	}
	return nil
}

func (x *xmlCodec) ReadBody(body interface{}) error {
	switch x.mt {
	case codec.Request:
		return x.s.ReadBody(body)
	case codec.Response:
		return x.c.ReadBody(body)
	case codec.Event:
		if body != nil {
			return xml.Unmarshal(x.buf.Bytes(), body)
		}
	default:
		return fmt.Errorf("Unrecognised message type: %v", x.mt)
	}
	return nil
}

//NewCodec return a xml codec
func NewCodec(rwc io.ReadWriteCloser) codec.Codec {
	return &xmlCodec{
		buf: bytes.NewBuffer(nil),
		rwc: rwc,
		c:   newClientCodec(rwc),
		s:   newServerCodec(rwc),
	}
}
