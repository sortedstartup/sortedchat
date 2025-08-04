package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"strings"

	pb "sortedstartup/chatservice/proto"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx    context.Context
	client pb.SortedChatClient
}

// NewApp creates a new App application struct
func NewApp(client pb.SortedChatClient) *App {
	return &App{
		client: client,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) DoChat(message string) string {
	clientStream, err := a.client.Chat(a.ctx, &pb.ChatRequest{
		Text:   message,
		ChatId: "d5eda2ad-855a-4cd6-a561-6340b1587682",
		Model:  "gpt-4.1",
	})
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	fmt.Println("Client Stream Created", clientStream)

	var responseBuilder strings.Builder

	for {
		response, err := clientStream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Stream ended successfully")
				break
			}
			return fmt.Sprintf("Stream error: %v", err)
		}

		responseBuilder.WriteString(response.Text)

		fmt.Printf("Received: %s\n", response.Text)
	}

	collectedResponse := responseBuilder.String()
	fmt.Printf("Complete response: %s\n", collectedResponse)

	return collectedResponse
}

func (a *App) CreateNewChat() string {
	res, err := a.client.CreateChat(a.ctx, &pb.CreateChatRequest{
		Name: "New Chat",
	})
	if err != nil {
		return fmt.Sprintf("Error creating chat: %v", err)
	}
	fmt.Printf("New chat created with ID: %s\n", res.ChatId)
	return res.ChatId
}

// whenever you need to generate go bindings -> main.go comment the channel listner

func Wails(client pb.SortedChatClient) {
	// Create an instance of the app structure
	// app := NewApp(apiserver)
	app := NewApp(client)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "wails-poc",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
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
