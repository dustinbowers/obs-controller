package main

import (
	"log"
	"obs-controller/controller"
)

func main() {
	log.Printf("Welcome!\n")
	ctl, err := controller.NewController(
		"localhost:4455",
		"7hZUoT9MDL3jXiTf",
		"29739507")
	if err != nil {
		panic(err)
	}

	err = ctl.Run()
	if err != nil {
		log.Printf("OBS Controller stopped.\nError: %s\n", err)
	}

	// Keep the controller running for now
	//quitChannel := make(chan os.Signal, 1)
	//signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	//<-quitChannel
}
