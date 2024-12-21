package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Nava890/jprqOwn.git/server/config"
	"github.com/Nava890/jprqOwn.git/server/events"
	"github.com/Nava890/jprqOwn.git/server/server"
	"github.com/Nava890/jprqOwn.git/server/tunnels"
)

type Jprq struct {
	config          config.Config
	eventServer     server.TCPServer
	publicServer    server.TCPServer
	publicServerTLS server.TCPServer
	httpTunnels     map[string]*tunnels.HttpTunnel
	userTunnels     map[string]map[string]tunnels.Tunnel
	subDomainMap    map[string]string
	activeTunnels   map[string]tunnels.Tunnel
}

func (j *Jprq) Init(conf config.Config) error {
	j.config = conf
	j.httpTunnels = make(map[string]*tunnels.HttpTunnel)
	j.userTunnels = make(map[string]map[string]tunnels.Tunnel)

	if err := j.eventServer.Init(conf.EventServerPort, "jprq_event_server"); err != nil {
		return err
	}
	if err := j.publicServer.Init(conf.PublicServerPort, "jprq_public_server"); err != nil {
		return err
	}
	err := j.publicServerTLS.InitTLS(conf.PublicServerTLSPort, "jprq_public_server_tls", conf.TLSCertFile, conf.TLSKeyFile)
	return err
}

func (j *Jprq) Start() {
	go j.eventServer.Start(j.serveEventConn)
	go j.publicServer.Start(j.servePublicConn)
	go j.publicServerTLS.Start(j.servePublicConn)
}
func (j *Jprq) Stop() error {
	if err := j.eventServer.Stop(); err != nil {
		return err
	}
	if err := j.publicServer.Stop(); err != nil {
		return err
	}
	err := j.publicServerTLS.Stop()
	return err
}
func (j *Jprq) servePublicConn(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	host, buffer, err := parseHost(conn)
	if err != nil || host == "" {
		writeResponse(conn, 400, "Bad Request", "Bad Request")
		return nil
	}
	if tunnelHost, ok := j.subDomainMap[host]; ok && tunnelHost != "" {
		host = tunnelHost
	}
	host = strings.ToLower(host)
	t, found := j.httpTunnels[host]
	if !found {
		writeResponse(conn, 404, "Not Found", "tunnel not found. create one at jprq.io")
		return fmt.Errorf("unknown host requested %s", host)
	}
	return t.PublicConnectionHandler(conn, buffer)
}
func (j *Jprq) serveEventConn(conn net.Conn) error {
	defer conn.Close()

	var event events.Event[events.TunnelRequested]
	if err := event.Read(conn); err != nil {
		return err
	}

	request := event.Data
	if request.Protocol != events.HTTP && request.Protocol != events.TCP {
		return events.WriteError(conn, "invalid protocol %s", request.Protocol)
	}

	if request.Subdomain == "" {
		request.Subdomain = "sample"
	}
	if err := validate(request.Subdomain); err != nil {
		return events.WriteError(conn, "invalid subdomain %s: %s", request.Subdomain, err.Error())
	}
	hostname := fmt.Sprintf("%s.%s", request.Subdomain, j.config.DomainName)
	if _, ok := j.httpTunnels[hostname]; ok {
		return events.WriteError(conn, "subdomain is busy: %s, try another one", request.Subdomain)
	}

	var t tunnels.Tunnel
	var maxConsLimit = j.config.MaxConsPerTunnel

	switch request.Protocol {
	case events.HTTP:
		tn, err := tunnels.NewHttp(hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError(conn, "failed to create http tunnel", err.Error())
		}
		j.subDomainMap[hostname] = hostname
		j.httpTunnels[hostname] = tn
		defer delete(j.subDomainMap, hostname)
		defer delete(j.httpTunnels, hostname)
		t = tn
	}

	if len(j.userTunnels[hostname]) == 0 {
		j.userTunnels[hostname] = make(map[string]tunnels.Tunnel)
	}
	tunnelId := fmt.Sprintf("%s:%d", t.Hostname(), t.PublicServerPort())
	j.userTunnels[hostname][tunnelId] = t
	defer delete(j.userTunnels[hostname], tunnelId)

	t.Open()
	defer t.Close()
	opened := events.Event[events.TunnelOpenened]{
		Data: &events.TunnelOpenened{
			Hostname:      t.Hostname(),
			Protocol:      t.Protocol(),
			PublicServer:  t.PublicServerPort(),
			PrivateServer: t.PrivateServerPort(),
		},
	}
	if err := opened.Write(conn); err != nil {
		return err
	}

	buffer := make([]byte, 8) // wait until connection is closed
	for {
		_ = conn.SetReadDeadline(time.Now().Add(time.Minute))
		if _, err := conn.Read(buffer); err == io.EOF {
			break
		}
	}
	return nil
}
