package main

import (
	"context"
	"log"
	"net/http"

	"github.com/open-event-systems/go-utils/gracefulshutdown"
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	gs := gracefulshutdown.GracefulHTTPShutdown{
		Server:  server,
		Context: context.Background(),
	}

	err := gs.ListenAndServe()
	log.Printf("exited: %v", err)
}
