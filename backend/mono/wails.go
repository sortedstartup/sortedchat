package main

import (
	"context"
	"embed"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend-build-wails/dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type MuxHandler struct {
	mux *http.ServeMux
}

func (h *MuxHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/hack")

	h.mux.ServeHTTP(res, req)
}

func Wails(mux *http.ServeMux) {

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "SortedChat",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: &MuxHandler{mux: mux},
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
