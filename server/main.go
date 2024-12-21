package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Nava890/jprqOwn.git/server/config"
)

func main() {
	var (
		conf config.Config
		jprq Jprq
	)
	err := conf.Load()
	if err != nil {
		log.Fatalf("failed to load config %s", err)
	}
	err = jprq.Init(conf)
	if err != nil {
		log.Fatalf("failed to init client %s", err)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	jprq.Start()
	defer jprq.Stop()
	<-signalChan
}
