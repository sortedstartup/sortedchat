//go:build wails

package main

import (
	"bytes"
	"context"
	"embed"
	"io"
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
	handler http.Handler
}

func (h *MuxHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/hack")

	// Buffer request body for grpc-web to avoid webkit reader conflicts
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	h.handler.ServeHTTP(res, req)
}

func Wails(handler http.Handler) {

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "SortedChat",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: &MuxHandler{handler: handler},
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

// NOOP since wails will block exit of the app unless running
// if we still wait for server error, `wail dev` gets stuck in generating bindings
func WaitForServerError(serverErr chan error) {

}
