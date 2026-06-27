package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Fast Switch",
		Width:             980,
		Height:            620,
		MinWidth:          860,
		MinHeight:         560,
		DisableResize:     false,
		Frameless:         true,
		StartHidden:       true,
		HideWindowOnClose: true,
		AlwaysOnTop:       true,
		BackgroundColour:  &options.RGBA{R: 246, G: 241, B: 233, A: 0},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               true,
				FullSizeContent:            true,
			},
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
