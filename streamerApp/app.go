package main

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"obs-controller/controller"
	"obs-controller/controller/types"
	"os"
	"time"
)

// App struct
type App struct {
	ctx           context.Context
	ObsController *controller.ObsController
}

// NewApp creates a new App application struct
func NewApp() *App {
	newApp := &App{}

	// Report logging up to the frontend
	multiWriter := io.MultiWriter(os.Stdout, newApp)
	log.SetOutput(multiWriter)

	return newApp
}

func (a *App) Write(p []byte) (n int, err error) {
	logLine := string(p)
	// Emit the log line as an event to the Wails frontend
	runtime.EventsEmit(a.ctx, "log_event", logLine)
	return len(p), nil
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	log.Printf("Welcome Streamer!\n")

	// Create the ObsController that holds the OBS and Web proxy websocket connections
	newController, err := controller.NewController()
	if err != nil {
		log.Fatalf("Failed to create OBS controller: %s", err)
	}
	a.ObsController = newController

	// Start the main listen-parse-update event loop
	//go func() {
	//	err = a.ObsController.Start()
	//	if err != nil {
	//		log.Printf("OBS Controller Error: %s\n", err)
	//	}
	//}()

	// Start sending connection updates to the frontend
	go func() {
		updateTicker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-updateTicker.C:
				runtime.EventsEmit(a.ctx, "connection_status", a.ObsController.ConnectionStatus)
			}
		}
	}()
}

// domReady is called after the front-end dom has been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue,
// false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.ObsController != nil {
		a.ObsController.Cleanup()
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

func (a *App) GetUserConfig() types.Config {
	return *a.ObsController.UserConfig
}

func (a *App) Connect() error {
	err := a.ObsController.Start()
	if err != nil {
		log.Printf("Failed to start obs controller: %s", err)
		a.ObsController.ConnectionStatus = "Error"
		return err
	}
	return nil
}

func (a *App) Disconnect() error {
	log.Printf("Attempting disconnect...")
	return a.ObsController.Stop()
}
