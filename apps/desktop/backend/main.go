package main

import (
	"embed"

	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "api-client",
		Description: "A local-first, Git-native API development environment",
		Services: []application.Service{
			application.NewService(&AppService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "API Client",
		Width:            1280,
		Height:           800,
		MinWidth:         940,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(15, 17, 21),
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}