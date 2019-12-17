package main

import (
	"fmt"
	"os"
	"time"
)

func main() {

	config, err := LoadConfig(`config.json`)
	if err != nil {
		fmt.Printf("Failed to load config file; %v", err)
		os.Exit(1)
	}

	fmt.Printf("Running for %d sec", config.Duration)
	time.Sleep(time.Second * time.Duration(config.Duration))
	fmt.Printf("Exit as %d", config.Result)
	os.Exit(config.Result)
}
