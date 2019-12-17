package ares

import (
	"encoding/json"
	"io/ioutil"
)

type ManagerConf struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type AppConf struct {
	Name string `json:"name"`
	Run  string `json:"run"`
	Dir  string `json:"dir"`
}

type Configuration struct {
	Port    string `json:"port"`
	Monitor string `json:"monitor"`
	Name    string `json:"name"`

	Manager *ManagerConf `json:"manager"`
	Apps    []*AppConf   `json:"apps"`

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
