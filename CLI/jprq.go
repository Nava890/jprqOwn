package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Nava890/jprqOwn.git/events"
)

type jprqClient struct {
	config       Config
	protocol     string
	subdomain    string
	localServer  string
	remoteServer string
	publicServer string
}

func (j *jprqClient) Start(port int) {
	eventCon, err := net.Dial("tcp", j.config.Remote.Events)

	if err != nil {
		log.Fatalf("failed to connect to the event server %s\n", err)
	}
	defer eventCon.Close()

	request := events.Event[events.TunnelRequested]{
		Data: &events.TunnelRequested{
			Protocol:  j.protocol,
			Subdomain: j.subdomain,
		},
	}
	if err := request.Write(eventCon); err != nil {
		log.Fatalf("failed to send request %s", err)
	}

	var t events.Event[events.TunnelOpenened]
	if err := t.Read(eventCon); err != nil {
		log.Fatalf("failed to reccieve tunnel info: %s", err)
	}
	if t.Data.ErrorMessage != "" {
		log.Fatalf(t.Data.ErrorMessage)
	}

	j.localServer = fmt.Sprintf("localhost:%d", port)
	j.remoteServer = fmt.Sprintf("%s:%d", j.config.Remote.Domain, t.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", t.Data.Hostname, t.Data.PublicServer)

	if j.protocol == "http" {
		j.publicServer = fmt.Sprintf("https://%s", t.Data.Hostname)
	}

	fmt.Printf("Status: \t Online \n")
	fmt.Printf("Protocol: \t %s \n", strings.ToUpper(j.protocol))
	fmt.Printf("Forwarded: \t %s -> %s \n", strings.TrimSuffix(j.publicServer, ":80"), j.localServer)

	var event events.Event[events.ConnectionRecieved]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("failed to recieve the data: %s", err)
		}
		go j.handleEvent(*event.Data)
	}
}
func (j *jprqClient) handleEvent(event events.ConnectionRecieved) {
	localCon, err := net.Dial("tcp", j.localServer)
	if err != nil {
		log.Printf("failed to connect to the local server: %s", err)
		return
	}
	defer localCon.Close()

	remoteCon, err := net.Dial("tcp", j.remoteServer)
	if err != nil {
		log.Printf("failed to connect to the remote server: %s", err)
		return
	}
	defer remoteCon.Close()

	buffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(buffer, event.ClientPort)
	remoteCon.Write(buffer)

	go tunnel.Bind(localCon, remoteCon, nil)
	tunnel.Bind(remoteCon, localCon, nil)
	return

}
