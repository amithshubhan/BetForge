package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/amithshubhan/Bet_Now/orderbook-engine/orderbook"
)

func PlaceOrderHandler(w http.ResponseWriter, r *http.Request) {
    var order orderbook.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "invalid input", http.StatusBadRequest)
        return
    }
	
    go orderbook.PlaceOrder(order) // goroutine or channel-driven
    w.WriteHeader(http.StatusAccepted)
}

type Match struct {
	MatchID string `json:"match_id"`
	TeamA   string `json:"team_a"`
	TeamB   string `json:"team_b"`
}

func RegisterMatchHandler(w http.ResponseWriter, r *http.Request) {
	var match Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	
	go orderbook.RegisterMatch(match.MatchID, match.TeamA, match.TeamB) // goroutine or channel-driven
	w.WriteHeader(http.StatusAccepted)
}
