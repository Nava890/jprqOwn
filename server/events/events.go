package events

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	TCP  string = "tcp"
	HTTP string = "http"
)

type EventType interface {
	TunnelRequested | TunnelOpenened | ConnectionRecieved
}

type Event[Type EventType] struct {
	Data *Type
}

type TunnelRequested struct {
	Protocol  string
	Subdomain string
}

type TunnelOpenened struct {
	Hostname      string
	Protocol      string
	PublicServer  string
	PrivateServer string
	ErrorMessage  string
}

type ConnectionRecieved struct {
	ClientIp    net.IP
	ClientPort  uint16
	RateLimited bool
}

func WriteError(eventWriter io.Writer, message string, args ...string) error {
	event := Event[TunnelOpenened]{
		Data: &TunnelOpenened{
			ErrorMessage: fmt.Sprintf(message, args),
		},
	}
	event.Write(eventWriter)
	return errors.New(event.Data.ErrorMessage)
}

func (e *Event[EventType]) Write(conn io.Writer) error {
	data, err := e.encode()
	if err != nil {
		return err
	}
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(data)))
	if _, err := conn.Write(length); err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err

}
func (e *Event[EventType]) encode() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(e.Data); err != nil {
		return nil, err
	}
	data := buffer.Bytes()
	return data, nil

}

func (e *Event[EventType]) Read(conn io.Reader) error {
	buffer := make([]byte, 2)
	if _,err := conn.Read(buffer),err != nil{
		return err
	}
	length := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, length)
	if _,err := conn.Read(buffer),err != nil{
		return err
	}
	err:=e.Decode(buffer)
	return err
}

func (e *Event[EventType]) Decode(data []byte) error {
		buffer:=bytes.NewBuffer(data)
		dec:=gob.NewDecoder(buffer)
		err:=dec.Decode(&data)
		return err
}
