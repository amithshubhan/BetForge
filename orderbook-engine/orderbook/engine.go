package orderbook

import (
	"fmt"
	"log"
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

// --- Enhanced Place Order with Sports Betting Logic ---

func PlaceOrder(order Order) {
	book := getOrderBook(order.MatchID, order.TeamID)
	opposingTeamID := getOpposingTeamID(order.MatchID, order.TeamID)
	opposingBook := getOrderBook(order.MatchID, opposingTeamID)

	book.mu.Lock()
	opposingBook.mu.Lock()
	defer book.mu.Unlock()
	defer opposingBook.mu.Unlock()

	if order.Side == "bid" {
		// When someone wants to buy Team A shares (BACK Team A)
		remainingQty := order.Quantity
		
		// STRATEGY 1: Match with asks (sell orders) in the same team's orderbook
		// This represents someone who wants to SELL their Team A position
		remainingQty = matchWithSameTeamAsks(order, book, remainingQty)
		
		// STRATEGY 2: Cross-team matching with opposing team's bids
		// Buying Team A is equivalent to selling Team B
		// So we can match with people buying Team B at compatible odds
		if remainingQty > 0 {
			remainingQty = matchWithOpposingTeamBids(order, opposingBook, remainingQty, opposingTeamID)
		}

		// If we still have quantity to fill, add to orderbook
		if remainingQty > 0 {
			order.Quantity = remainingQty
			book.Bids.Push(order)
			log.Printf("Partial fill - Added remaining bid to orderbook: %.2f units at %.2f", 
				remainingQty, order.Price)
		}

	} else if order.Side == "ask" {
		// When someone wants to sell Team A shares (LAY Team A)
		remainingQty := order.Quantity
		
		// STRATEGY 1: Match with bids (buy orders) in the same team's orderbook
		// This represents someone who wants to BUY Team A position
		remainingQty = matchWithSameTeamBids(order, book, remainingQty)
		
		// STRATEGY 2: Cross-team matching with opposing team's asks
		// Selling Team A is equivalent to buying Team B
		// So we can match with people selling Team B at compatible odds
		if remainingQty > 0 {
			remainingQty = matchWithOpposingTeamAsks(order, opposingBook, remainingQty, opposingTeamID)
		}

		// If we still have quantity to fill, add to orderbook
		if remainingQty > 0 {
			order.Quantity = remainingQty
			book.Asks.Push(order)
			log.Printf("Partial fill - Added remaining ask to orderbook: %.2f units at %.2f", 
				remainingQty, order.Price)
		}
	}

	// Update market prices after matching
	updateMatchPrices(order.MatchID)
	
	// Print the updated orderbook state
	PrintOrderBook(order.MatchID, order.TeamID, book, opposingTeamID, opposingBook)
	
	// Publish the event for other services
	PublishMatchEvent(order)
}

// --- Matching Functions ---

func matchWithSameTeamAsks(bidOrder Order, book *OrderBook, remainingQty float64) float64 {
	for remainingQty > 0 && book.Asks.Len() > 0 {
		bestAsk := book.Asks.Peek()
		
		// Check if the ask price is less than or equal to bid price
		if bestAsk.Price > bidOrder.Price {
			break // No matching asks at acceptable price
		}

		// Calculate how much we can trade
		matchQty := min(remainingQty, bestAsk.Quantity)
		tradePrice := bestAsk.Price // Use ask price for the trade
		tradeValue := tradePrice * matchQty

		// Execute the trade
		log.Printf("SAME-TEAM Match: Bid for %s - %.2f units at Price %.2f (Total: ₹%.2f)", 
			bidOrder.TeamID, matchQty, tradePrice, tradeValue)

		// TODO: Transfer money and shares
		executeTrade(bidOrder.UserID, bestAsk.UserID, matchQty, tradePrice, bidOrder.TeamID, "SAME_TEAM_BID_ASK")

		// Update quantities
		bestAsk.Quantity -= matchQty
		remainingQty -= matchQty

		// Remove the ask if fully filled
		if bestAsk.Quantity <= 0 {
			book.Asks.Pop()
		}
	}
	return remainingQty
}

func matchWithSameTeamBids(askOrder Order, book *OrderBook, remainingQty float64) float64 {
	for remainingQty > 0 && book.Bids.Len() > 0 {
		bestBid := book.Bids.Peek()
		
		// Check if the bid price is greater than or equal to ask price
		if bestBid.Price < askOrder.Price {
			break // No matching bids at acceptable price
		}

		// Calculate how much we can trade
		matchQty := min(remainingQty, bestBid.Quantity)
		tradePrice := bestBid.Price // Use bid price for the trade
		tradeValue := tradePrice * matchQty

		// Execute the trade
		log.Printf("SAME-TEAM Match: Ask for %s - %.2f units at Price %.2f (Total: ₹%.2f)", 
			askOrder.TeamID, matchQty, tradePrice, tradeValue)

		// TODO: Transfer money and shares
		executeTrade(bestBid.UserID, askOrder.UserID, matchQty, tradePrice, askOrder.TeamID, "SAME_TEAM_ASK_BID")

		// Update quantities
		bestBid.Quantity -= matchQty
		remainingQty -= matchQty

		// Remove the bid if fully filled
		if bestBid.Quantity <= 0 {
			book.Bids.Pop()
		}
	}
	return remainingQty
}

func matchWithOpposingTeamBids(bidOrder Order, opposingBook *OrderBook, remainingQty float64, opposingTeamID string) float64 {
	// Buying Team A can match with buying Team B if the combined odds make sense
	// This creates natural price discovery between teams
	
	for remainingQty > 0 && opposingBook.Bids.Len() > 0 {
		opposingBid := opposingBook.Bids.Peek()
		
		// Check if cross-team trade is profitable
		// The sum of odds should be close to the total probability (accounting for margin)
		if !areOddsCompatibleForCrossTrade(bidOrder.Price, opposingBid.Price) {
			break
		}

		matchQty := min(remainingQty, opposingBid.Quantity)
		// Use average price or more sophisticated pricing model
		tradePrice := calculateCrossTradePrice(bidOrder.Price, opposingBid.Price)
		tradeValue := tradePrice * matchQty

		log.Printf("CROSS-TEAM Match: %s Bid %.2f vs %s Bid %.2f - %.2f units at %.2f (Total: ₹%.2f)", 
			bidOrder.TeamID, bidOrder.Price, opposingTeamID, opposingBid.Price, 
			matchQty, tradePrice, tradeValue)

		// Execute cross-team trade
		executeCrossTrade(bidOrder.UserID, opposingBid.UserID, matchQty, tradePrice, 
			bidOrder.TeamID, opposingTeamID, "CROSS_TEAM_BID_BID")

		opposingBid.Quantity -= matchQty
		remainingQty -= matchQty

		if opposingBid.Quantity <= 0 {
			opposingBook.Bids.Pop()
		}
	}
	return remainingQty
}

func matchWithOpposingTeamAsks(askOrder Order, opposingBook *OrderBook, remainingQty float64, opposingTeamID string) float64 {
	// Selling Team A can match with selling Team B under certain conditions
	
	for remainingQty > 0 && opposingBook.Asks.Len() > 0 {
		opposingAsk := opposingBook.Asks.Peek()
		
		if !areOddsCompatibleForCrossTrade(askOrder.Price, opposingAsk.Price) {
			break
		}

		matchQty := min(remainingQty, opposingAsk.Quantity)
		tradePrice := calculateCrossTradePrice(askOrder.Price, opposingAsk.Price)
		tradeValue := tradePrice * matchQty

		log.Printf("CROSS-TEAM Match: %s Ask %.2f vs %s Ask %.2f - %.2f units at %.2f (Total: ₹%.2f)", 
			askOrder.TeamID, askOrder.Price, opposingTeamID, opposingAsk.Price, 
			matchQty, tradePrice, tradeValue)

		executeCrossTrade(askOrder.UserID, opposingAsk.UserID, matchQty, tradePrice, 
			askOrder.TeamID, opposingTeamID, "CROSS_TEAM_ASK_ASK")

		opposingAsk.Quantity -= matchQty
		remainingQty -= matchQty

		if opposingAsk.Quantity <= 0 {
			opposingBook.Asks.Pop()
		}
	}
	return remainingQty
}

// --- Trading and Price Logic ---

func executeTrade(buyerID, sellerID string, quantity, price float64, teamID, tradeType string) {
	tradeValue := price * quantity
	log.Printf("TRADE EXECUTED [%s]: Buyer: %s, Seller: %s, Team: %s, Qty: %.2f, Price: %.2f, Value: ₹%.2f", 
		tradeType, buyerID, sellerID, teamID, quantity, price, tradeValue)
	
	// TODO: Implement actual transfers
	// transferMoney(buyerID, sellerID, tradeValue)
	// transferShares(sellerID, buyerID, quantity, teamID)
}

func executeCrossTrade(user1ID, user2ID string, quantity, price float64, team1ID, team2ID, tradeType string) {
	tradeValue := price * quantity
	log.Printf("CROSS-TRADE EXECUTED [%s]: User1: %s (%s), User2: %s (%s), Qty: %.2f, Price: %.2f, Value: ₹%.2f", 
		tradeType, user1ID, team1ID, user2ID, team2ID, quantity, price, tradeValue)
	
	// TODO: Implement cross-team position transfers
	// This is more complex as it involves offsetting positions
}

func areOddsCompatibleForCrossTrade(odds1, odds2 float64) bool {
	// Check if the combined implied probabilities make sense for a cross-trade
	impliedProb1 := 1.0 / odds1
	impliedProb2 := 1.0 / odds2
	totalProb := impliedProb1 + impliedProb2
	
	// Allow trades where total probability is between 95% and 105% (5% margin)
	return totalProb >= 0.95 && totalProb <= 1.05
}

func calculateCrossTradePrice(odds1, odds2 float64) float64 {
	// Simple average - you can implement more sophisticated pricing
	return (odds1 + odds2) / 2
}

func updateMatchPrices(matchID string) {
	mu.RLock()
	defer mu.RUnlock()
	
	teams := MatchTeams[matchID]
	teamABook := MatchBooks[matchID][teams[0]]
	teamBBook := MatchBooks[matchID][teams[1]]
	
	if teamABook == nil || teamBBook == nil {
		return
	}
	
	// Calculate current market prices based on best bids/asks
	teamAPrice := calculateMarketPrice(teamABook)
	teamBPrice := calculateMarketPrice(teamBBook)
	
	log.Printf("Market Update - Match %s: %s=%.2f, %s=%.2f", 
		matchID, teams[0], teamAPrice, teams[1], teamBPrice)
}

func calculateMarketPrice(book *OrderBook) float64 {
	book.mu.Lock()
	defer book.mu.Unlock()
	
	var midPrice float64 = 2.0 // Default odds
	
	if book.Bids.Len() > 0 && book.Asks.Len() > 0 {
		bestBid := book.Bids.Peek().Price
		bestAsk := book.Asks.Peek().Price
		midPrice = (bestBid + bestAsk) / 2
	} else if book.Bids.Len() > 0 {
		midPrice = book.Bids.Peek().Price
	} else if book.Asks.Len() > 0 {
		midPrice = book.Asks.Peek().Price
	}
	
	return midPrice
}

// --- Helper Functions ---

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func PrintOrderBook(matchID, teamID string, book *OrderBook, opposingTeamID string, opposingBook *OrderBook) {
	fmt.Printf("\n=== Match %s Order Books ===\n", matchID)
	
	fmt.Printf("Team %s:\n", teamID)
	fmt.Println("  Bids (Buy Orders):")
	if book.Bids.Len() > 0 {
		for _, bid := range book.Bids.Items() {
			fmt.Printf("    Price: %.2f, Quantity: %.2f, User: %s\n", bid.Price, bid.Quantity, bid.UserID)
		}
	} else {
		fmt.Println("    No bids")
	}
	
	fmt.Println("  Asks (Sell Orders):")
	if book.Asks.Len() > 0 {
		for _, ask := range book.Asks.Items() {
			fmt.Printf("    Price: %.2f, Quantity: %.2f, User: %s\n", ask.Price, ask.Quantity, ask.UserID)
		}
	} else {
		fmt.Println("    No asks")
	}

	fmt.Printf("\nTeam %s:\n", opposingTeamID)
	fmt.Println("  Bids (Buy Orders):")
	if opposingBook.Bids.Len() > 0 {
		for _, bid := range opposingBook.Bids.Items() {
			fmt.Printf("    Price: %.2f, Quantity: %.2f, User: %s\n", bid.Price, bid.Quantity, bid.UserID)
		}
	} else {
		fmt.Println("    No bids")
	}
	
	fmt.Println("  Asks (Sell Orders):")
	if opposingBook.Asks.Len() > 0 {
		for _, ask := range opposingBook.Asks.Items() {
			fmt.Printf("    Price: %.2f, Quantity: %.2f, User: %s\n", ask.Price, ask.Quantity, ask.UserID)
		}
	} else {
		fmt.Println("    No asks")
	}
	
	fmt.Println("=====================================")
}