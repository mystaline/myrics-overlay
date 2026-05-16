package main

import (
	"context"
	_ "embed"
	"embed"

	"github.com/mystaline/myrics-overlay/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed configs/config.yaml.example
var defaultConfig []byte

//go:embed all:frontend
var assets embed.FS

func main() {
	cfg := config.LoadOrDefault("configs/config.yaml", defaultConfig)

	app := NewApp(cfg)

	if err := wails.Run(&options.App{
		Title:  "Myrics",
		Width:  800,
		Height: 100,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DisableResize:    true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
		OnShutdown:       func(ctx context.Context) { deleteLyricsFile() },
		Bind:             []any{app},
		Frameless:        true,
		Windows: &windows.Options{
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
			DisableWindowIcon:    true,
		},
	}); err != nil {
		panic(err)
	}
}
