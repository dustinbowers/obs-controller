package main

import (
	"log"
	"obs-controller/controller"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func KeepAlive() {
	for {
		time.Sleep(30 * time.Second)
	}
}

func main() {
	log.Printf("Welcome!\n")
	ctl, err := controller.NewController(
		"localhost:4455",
		"7hZUoT9MDL3jXiTf",
		"29739507")
	if err != nil {
		panic(err)
	}

	//go func() {
	//	err := ctl.Run()
	//	if err != nil {
	//		fmt.Errorf("controller crashed: %v", err)
	//	}
	//}()

	err = ctl.Run()
	if err != nil {
		log.Printf("OBS Controller stopped.\n")
	}

	// Keep the controller running for now
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}
