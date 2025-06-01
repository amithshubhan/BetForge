package orderbook

import (
	"fmt"
	"sync"
)


type OrderBook struct {
	Bids *Heap[Order] // Max-heap for Bids
	Asks *Heap[Order] // Min-heap for Asks
	mu   sync.Mutex
}

// --- Storage for Match Data ---

var (
	MatchBooks = make(map[string]map[string]*OrderBook) // matchID → (teamID → OrderBook)
	MatchTeams = make(map[string][2]string)             // matchID → [teamA, teamB]
	mu         sync.RWMutex
)

// --- Match Registration ---

func RegisterMatch(matchID, teamA, teamB string) {
	mu.Lock()
	defer mu.Unlock()
	MatchTeams[matchID] = [2]string{teamA, teamB}
	MatchBooks[matchID] = make(map[string]*OrderBook)
}

// --- OrderBook Retrieval ---

func getOrderBook(matchID, teamID string) *OrderBook {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := MatchBooks[matchID]; !exists {
		MatchBooks[matchID] = make(map[string]*OrderBook)
	}
	if _, exists := MatchBooks[matchID][teamID]; !exists {
		MatchBooks[matchID][teamID] = &OrderBook{
			Bids: New[Order](func(a, b Order) bool { return a.Price > b.Price }),
			Asks: New[Order](func(a, b Order) bool { return a.Price < b.Price }),
		}
	}
	return MatchBooks[matchID][teamID]
}

// --- Opposing Team Lookup ---

func getOpposingTeamID(matchID, teamID string) string {
	mu.RLock()
	defer mu.RUnlock()
	teams := MatchTeams[matchID]
	if teams[0] == teamID {
		return teams[1]
	}
	return teams[0]
}

// --- Place Order and Matching Logic ---

func PlaceOrder(order Order) {
	book := getOrderBook(order.MatchID, order.TeamID)
	opposingTeamID := getOpposingTeamID(order.MatchID, order.TeamID)
	opposingBook := getOrderBook(order.MatchID, opposingTeamID)

	book.mu.Lock()
	opposingBook.mu.Lock()
	defer book.mu.Unlock()
	defer opposingBook.mu.Unlock()

	mirror := func(p float64) float64 { return 1.0 - p }

	if order.Side == "bid" {
		remainingQty := order.Quantity
		maxPrice := order.Price

		for remainingQty > 0 && opposingBook.Asks.Len() > 0 {
			bestAsk := opposingBook.Asks.Peek()
			if bestAsk.Price > mirror(maxPrice) {
				break
			}
			matchQty := min(remainingQty, bestAsk.Quantity)
			tradePrice := mirror(bestAsk.Price)
			tradeValue := tradePrice * matchQty

			fmt.Printf("Matched: BUY %.2f units at Price %.2f (₹%.2f)\n", matchQty, tradePrice, tradeValue)

			bestAsk.Quantity -= matchQty
			remainingQty -= matchQty
			if bestAsk.Quantity <= 0 {
				opposingBook.Asks.Pop()
			}
		}

		if remainingQty > 0 {
			order.Quantity = remainingQty
			book.Bids.Push(order)
		}

	} else if order.Side == "ask" {
		remainingQty := order.Quantity
		minPrice := order.Price

		for remainingQty > 0 && opposingBook.Bids.Len() > 0 {
			bestBid := opposingBook.Bids.Peek()
			if bestBid.Price < mirror(minPrice) {
				break
			}
			matchQty := min(remainingQty, bestBid.Quantity)
			tradePrice := mirror(bestBid.Price)
			tradeValue := tradePrice * matchQty

			fmt.Printf("Matched: SELL %.2f units at Price %.2f (₹%.2f)\n", matchQty, tradePrice, tradeValue)

			bestBid.Quantity -= matchQty
			remainingQty -= matchQty
			if bestBid.Quantity <= 0 {
				opposingBook.Bids.Pop()
			}
		}

		if remainingQty > 0 {
			order.Quantity = remainingQty
			book.Asks.Push(order)
		}
	}

	PrintOrderBook(order.MatchID, order.TeamID, book, opposingTeamID, opposingBook)
	fmt.Printf("Order placed: %+v\n", order)
	PublishMatchEvent(order)
}

// --- Helper Functions ---

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func PrintOrderBook(matchID, teamID string, book *OrderBook, opposingTeamID string, opposingBook *OrderBook) {
	fmt.Printf("\nMatch %s - Order Book for Team %s:\n", matchID, teamID)
	fmt.Println("Bids:")
	for _, bid := range book.Bids.Items() {
		fmt.Printf("Price: %.2f, Quantity: %.2f\n", bid.Price, bid.Quantity)
	}
	fmt.Println("Asks:")
	for _, ask := range book.Asks.Items() {
		fmt.Printf("Price: %.2f, Quantity: %.2f\n", ask.Price, ask.Quantity)
	}

	fmt.Printf("\nMatch %s - Order Book for Opposing Team %s:\n", matchID, opposingTeamID)
	fmt.Println("Bids:")
	for _, bid := range opposingBook.Bids.Items() {
		fmt.Printf("Price: %.2f, Quantity: %.2f\n", bid.Price, bid.Quantity)
	}
	fmt.Println("Asks:")
	for _, ask := range opposingBook.Asks.Items() {
		fmt.Printf("Price: %.2f, Quantity: %.2f\n", ask.Price, ask.Quantity)
	}
}

