package main

import (
	"context"
	"log"

	"github.com/amithshubhan/Bet_Now/orderbook-engine/orderbook"

	"github.com/amithshubhan/Bet_Now/orderbookpb"
)

type orderbookServer struct {
	orderbookpb.UnimplementedOrderbookServiceServer
}

func (s *orderbookServer) RegisterMatch(ctx context.Context, req *orderbookpb.MatchRequest) (*orderbookpb.RegisterMatchResponse, error) {
	log.Printf("Registering match: %s (%s vs %s)", req.MatchId, req.TeamA, req.TeamB)
	orderbook.RegisterMatch(req.MatchId, req.TeamA, req.TeamB)
	return &orderbookpb.RegisterMatchResponse{
		Status: "Match registered successfully",
	}, nil
}
