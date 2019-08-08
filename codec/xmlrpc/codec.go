package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/micro/go-micro/codec"
)

type clientCodec struct {
	rwc io.ReadWriteCloser
}

type serverCodec struct {
	rwc io.ReadWriteCloser
}

type request struct {
	ServiceMethod string
	Seq           string
}

type response struct {
	ServiceMethod string
	Seq           string
	Error         string
}

func (c *clientCodec) Write(m *codec.Message, body interface{}) error {
	if err := marshalToStream(c.rwc, &request{
		ServiceMethod: m.Endpoint,
		Seq:           m.Id,
	}); err != nil {
		return err
	}
	if err := marshalToStream(c.rwc, body); err != nil {
		return err
	}
	return nil
}

func (c *clientCodec) ReadHeader(m *codec.Message) error {
	r := &response{}
	if err := unmarshalFromStream(c.rwc, r); err != nil {
		return err
	}
	m.Id = r.Seq
	m.Endpoint = r.ServiceMethod
	m.Error = r.Error
	return nil
}

func (c *clientCodec) ReadBody(body interface{}) error {
	if body == nil {
		return nil
	}
	return unmarshalFromStream(c.rwc, body)
}

func (c *clientCodec) Close() error {
	return c.rwc.Close()
}

func (s *serverCodec) ReadHeader(m *codec.Message) error {
	r := &request{}
	if err := unmarshalFromStream(s.rwc, r); err != nil {
		return err
	}
	m.Id = r.Seq
	m.Endpoint = r.ServiceMethod
	return nil
}

func (s *serverCodec) ReadBody(body interface{}) error {
	if body == nil {
		return nil
	}
	return unmarshalFromStream(s.rwc, body)
}

func unmarshalFromStream(rwc io.ReadWriteCloser, val interface{}) (err error) {

	lenbuf := make([]byte, 4)
	var n int
	n, err = io.ReadFull(rwc, lenbuf)
	if err != nil {
		return err
	}
	if n != 4 {
		return io.ErrUnexpectedEOF
	}
	length := n + 4
	b := make([]byte, length)
	n, err = io.ReadFull(rwc, b[4:])
	if err != nil {
		if err == io.EOF {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	if n != int(length-4) {
		return io.ErrUnexpectedEOF
	}

	newbytes := bytes.NewBuffer(b).Bytes()
	err = xml.Unmarshal(newbytes, val)
	if err != nil {
		return err
	}
	return err
}

func marshalToStream(rwc io.ReadWriteCloser, val interface{}) (err error) {
	buf, err := xml.Marshal(val)
	if err != nil {
		return err
	}
	_, err = rwc.Write(buf)

	return err
}

func (s *serverCodec) Write(m *codec.Message, body interface{}) error {

	res := &response{
		ServiceMethod: m.Endpoint,
		Seq:           m.Id,
		Error:         m.Error,
	}

	if err := marshalToStream(s.rwc, res); err != nil {
		return err
	}
	if err := marshalToStream(s.rwc, body); err != nil {
		return err
	}
	return nil
}

func (s *serverCodec) Close() error {
	return s.rwc.Close()
}

func newClientCodec(rwc io.ReadWriteCloser) *clientCodec {
	return &clientCodec{rwc}
}

func newServerCodec(rwc io.ReadWriteCloser) *serverCodec {
	return &serverCodec{rwc}
}
