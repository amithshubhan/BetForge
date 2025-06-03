package main

import (
	"context"
	"log"
	"time"

	"github.com/amithshubhan/Bet_Now/orderbookpb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	orderbookAddr = "localhost:50051" // Orderbook service gRPC endpoint
)

func main() {
	conn, err := grpc.Dial(orderbookAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to orderbook service: %v", err)
	}
	defer conn.Close()

	client := orderbookpb.NewOrderbookServiceClient(conn)

	ticker := time.NewTicker(60 * time.Second) // Every 3 hours
	defer ticker.Stop()

	log.Println("Match Service started. Creating matches every 1 minute...")

	for {
		select {
		case <-ticker.C:
			go createAndRegisterMatch(client)
		}
	}
}

func createAndRegisterMatch(client orderbookpb.OrderbookServiceClient) {
	matchID := uuid.New().String()
	teamA := "CSK"
	teamB := "MI"

	log.Printf("Creating new match: %s vs %s (ID: %s)", teamA, teamB, matchID)

	req := &orderbookpb.MatchRequest{
		MatchId: matchID,
		TeamA:   teamA,
		TeamB:   teamB,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RegisterMatch(ctx, req)
	if err != nil {
		log.Printf("failed to register match: %v", err)
		return
	}

	log.Printf("Match registered successfully: %s", resp.Status)
}
// 770b8b49-027b-46ec-b427-d45b80e0a137