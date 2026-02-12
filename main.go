package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// TODO: Create a new App instance
	app := NewApp()

	// TODO: Configure Wails application
	err := wails.Run(&options.App{
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DisableResize: true,
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

		Windows: &windows.Options{
			WindowIsTranslucent: true,
		},
	})
	if err != nil {
		panic(err)
	}
}
