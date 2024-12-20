package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Flags struct {
	subdomain string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatalf("no command specified")
	}
	if len(os.Args) < 3 {
		log.Fatalf("not the required amount of args")
	}
	protocol, port := "", 0
	command, arg := os.Args[1], os.Args[2]
	flags := parseFlags(os.Args[3:])
	switch command {
	case "tcp", "http":
		protocol = command
		port, _ = strconv.Atoi(arg)
	default:
		log.Fatal("unknown command: %s, jprq --help", command)
	}
	if port < 0 {
		log.Fatal("port number cannot be less than zero")
	}

	var conf Config
	if err := conf.Load(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("jprq %s \t press Ctrl+C to quit\n\n", "own")
	defer log.Println("jprq tunnel closed")

	client := jprqClient{
		config:    conf,
		protocol:  protocol,
		subdomain: flags.subdomain,
	}

	go client.Start(port)
}
func parseFlags(args []string) Flags {
	var flags Flags
	for i, arg := range args {
		switch arg {
		case "-s", "-subdomain", "--subdomain":
			flags.subdomain = args[i+1]
		}
	}
	return flags
}
