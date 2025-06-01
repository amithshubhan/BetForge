package main

import (
	"log"
	"net/http"
	"orderbook-engine/handlers"
)

func main() {
    http.HandleFunc("POST /place-order", handlers.PlaceOrderHandler)
	http.HandleFunc("POST /register-match", handlers.RegisterMatchHandler)
    log.Println("Orderbook Engine running on :8081")
    log.Fatal(http.ListenAndServe(":8081", nil))
}
