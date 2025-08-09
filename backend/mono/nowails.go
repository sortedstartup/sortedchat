//go:build !wails

package main

import (
	"log"
	"net/http"
)

// Wails is a no-op function when wails build tag is not present
func Wails(mux *http.ServeMux) {
	log.Println("server-only build, Wails is not enabled")
}

// When we dont use wails, we want to wait for the server to error
// If wails is being used it will block the exit of the app unless running
// we only need it in case of no-wails,
// if this wait code it used with wails also it cause `wails dev` to hang
func WaitForServerError(serverErr chan error) {
	err := <-serverErr
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
