package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// TODO: Create a new App instance
	app := NewApp()

	// TODO: Configure Wails application
	err := wails.Run(&options.App{
		Title:  "Lyrics Overlay",
		Width:  800,
		Height: 120,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// TODO: Set transparent background
		// Alpha value 0 makes the window background fully transparent
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},

		// TODO: Set startup callback
		OnStartup: app.startup,

		// TODO: Bind Go methods to frontend
		Bind: []interface{}{
			app,
		},

		// TODO: Enable frameless window (no title bar)
		Frameless: true,

		// TODO: Platform-specific options
		Linux: &linux.Options{
			// Enable window transparency on Linux
			WindowIsTranslucent: true,
		},

		// TODO: Add Windows options ?
		// Windows: &windows.Options{
		//     WindowIsTranslucent: true,
		// },
	})
	if err != nil {
		panic(err)
	}
}
