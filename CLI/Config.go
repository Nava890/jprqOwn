package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

var localConfig = ".jprq-config"
var remoteConfig = ""

type Config struct {
	Remote struct {
		Domain string `json:"domain"`
		Events string `json:"events"`
	}
}

func (c *Config) Load() error {
	response := `{
		"domain": "nava.in.net",
		"events": "event.nava.in.net:4321"
	}`

	if err := json.NewDecoder(strings.NewReader(response)).Decode(&c.Remote); err != nil {
		return fmt.Errorf("error decoding config file: %s", err)
	}
	return nil

}
