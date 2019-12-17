package main

import (
	ares "local/ares/manager/core"
)

func main() {
	app := ares.NewApp()
	app.Init()
	app.Run()

	app.Close()
}
