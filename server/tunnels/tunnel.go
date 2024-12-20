package tunnels

import (
	"io"
	"net"
	"sync"

	"github.com/Nava890/jprqOwn.git/server/server"
)

type Tunnel interface {
	Open()
	Close()
	Hostname() string
	Protocol() string
	PublicServerPort() string
	PrivateServerPort() string
}

type tunnel struct {
	hostName      string
	maxConLimit   int
	eventWriter   io.Writer
	eventWriterMx sync.Mutex
	privateServer server.TCPServer
	publicCons    map[uint16]net.Conn
	initialBuffer map[uint16][]byte
}

func newTunnel(hostname string, eventWrite io.Writer, maxConLimit int) tunnel {
	return tunnel{
		hostName:      hostname,
		maxConLimit:   maxConLimit,
		eventWriter:   eventWrite,
		publicCons:    make(map[uint16]net.Conn),
		initialBuffer: make(map[uint16][]byte),
	}
}
func (t *tunnel) Close() {
	t.privateServer.Stop()
	for port, con := range t.publicCons {
		con.Close()
		delete(t.publicCons, port)
		delete(t.initialBuffer, port)
	}
}

func (t *tunnel) Hostname() string {
	return t.hostName
}

func (t *tunnel) PrivateServerPort() uint16 {
	return t.privateServer.Port()
}
