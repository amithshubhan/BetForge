package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/amithshubhan/Bet_Now/internal/routes"
)

func main() {
    // Creating a new ServerMux
    router := http.NewServeMux()

    // Register routes
	routes.RegisterRoutes(router)

    // Start the server on port 8080
    port := ":8080"
    server := &http.Server{
        Addr:    port,
        Handler: router,
    }

    fmt.Printf("Server is listening on port %s\n", port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}