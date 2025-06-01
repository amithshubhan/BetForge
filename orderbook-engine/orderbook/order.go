package orderbook

type Order struct {
    ID       string `json:"id"`
    MatchID  string `json:"match_id"`
    TeamID   string `json:"team_id"`
    UserID   string `json:"user_id"`
    Side     string `json:"side"` // "bid" or "ask"
    Price    float64 `json:"price"`
    Quantity float64 `json:"quantity"`
}
