package main

import (
	"context"
	"fmt"
	"log"
	"obs-controller/controller"
)

// App struct
type App struct {
	ctx           context.Context
	ObsController *controller.ObsController
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	log.Printf("Welcome Streamer!\n")

	// Load configs from file
	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Failed to load config.toml: %s", err)
	}

	// Fetch twitch user ID from username
	twitchUserId, err := GetTwitchUserID(config.TwitchUsername)
	if err != nil {
		log.Fatalf("Failed to get twitch user id: %s\n", err)
	}

	// Create the ObsController that holds the OBS and Web proxy websocket connections
	a.ObsController, err = controller.NewController(
		fmt.Sprintf("%s:%s", config.ObsHost, config.ObsPort),
		config.ObsPassword,
		twitchUserId)
	if err != nil {
		log.Fatalf("Failed to create OBS controller: %s", err)
	}
	defer a.ObsController.Cleanup()

	// Start the main listen-parse-update event loop
	go func() {
		err = a.ObsController.Run()
		if err != nil {
			log.Printf("OBS Controller Error: %s\n", err)
		}
	}()

}

// domReady is called after the front-end dom has been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
	// 在这里添加你的操作
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
	// Perform your teardown here
	// 在此处做一些资源释放的操作
}
