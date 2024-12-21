package tunnels

import (
	"io"
	"net"
)

const DefaultPort = 80

type HttpTunnel struct {
	tunnel
}

func NewHttp(hostName string, eventWriter io.Writer, maxConLimit int) (*HttpTunnel, error) {
	t := &HttpTunnel{
		tunnel: newTunnel(hostName, eventWriter, maxConLimit),
	}
	if err := t.privateServer.Init(0, "http-tunnel-privateServer"); err != nil {
		return t, err
	}
	return t, nil
}
func (t *HttpTunnel) Protocol() string {
	return "http"
}

func (t *HttpTunnel) PublicServerPort() uint16 {
	return DefaultPort
}

func (t *HttpTunnel) Open() {
	go t.privateServer.Start(t.privateConnectionHandler)
}

func (t *HttpTunnel) PublicConnectionHandler(publicCon net.Conn, initialBuffer []byte) error {
	port := uint16(publicCon.RemoteAddr().(*net.TCPAddr).Port)
	t.initialBuffer[port] = initialBuffer
	return t.publicConnectionHandler(publicCon)
}
