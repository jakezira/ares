package main

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Duration int `json:"duration"`
	Result   int `json:"result"`
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
