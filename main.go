package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create service instances
	qbitService := &QBitService{}
	matcherService := &MatcherService{}

	// Create the application
	app := application.New(application.Options{
		Name:        "qBittorrent File Matcher",
		Description: "Match torrents with files on disk",
		Services: []application.Service{
			application.NewService(qbitService),
			application.NewService(matcherService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create the main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "qBittorrent File Matcher",
		Width:  1200,
		Height: 800,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(9, 9, 11), // Dark background
		URL:              "/",
	})

	// Run the application
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
