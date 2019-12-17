package ares

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	// "manager" or "node"
	Port    string `json:"port"`
	Monitor string `json:"monitor"`

	Debug      string `json:"debug"`
	ZipkinAddr string `json:"zipkin"`
}

func LoadConfig(filename string) (Configuration, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Configuration{}, err
	}

	var c Configuration
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Configuration{}, err
	}

	return c, nil
}
