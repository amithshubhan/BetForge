package routes

import (
	"fmt"
	"net/http"
)

// RegisterRoutes sets up all the routes for the application.
func RegisterRoutes(router *http.ServeMux) {
    router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Welcome to the Sports Betting App!")
    })

    // Add more routes here as needed
    router.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "About the Sports Betting App")
    })
}