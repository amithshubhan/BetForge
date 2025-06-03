package main

import (
	"log"
	"net"
	"net/http"

	"github.com/amithshubhan/Bet_Now/orderbook-engine/handlers"
	"github.com/amithshubhan/Bet_Now/orderbookpb"
	"google.golang.org/grpc"
)



func startGRPCServer(grpcServer *grpc.Server, listener net.Listener) {
	log.Printf("Starting gRPC server on %s", listener.Addr().String())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startHTTPServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/place-order", handlers.PlaceOrderHandler)
	
	log.Printf("Starting HTTP server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

func main() {
	// Initialize gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderbookpb.RegisterOrderbookServiceServer(grpcServer, &orderbookServer{})

	// Start servers in goroutines
	go startGRPCServer(grpcServer, listener)
	go startHTTPServer(":8081")

	// Keep the main goroutine alive
	select {}
}
