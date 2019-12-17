package main

import (
	ares "local/ares/node/core"
)

func main() {
	app := ares.NewApp()
	app.Init()
	app.Run()

	app.Close()
}
