package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Time-In-Force (TIF) Example ===\n")

	// Create engine
	config := engine.Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
	}
	eng := engine.New(config)
	eng.Start()
	defer eng.Stop()

	// Monitor events
	go monitorEvents(eng.Events())

	// Wait for engine to start
	time.Sleep(100 * time.Millisecond)

	// ========================================
	// Setup: Place some resting orders in the book
	// ========================================
	fmt.Println("--- Setup: Placing resting orders ---")

	// Sell orders at different price levels
	sellOrders := []*order.Order{
		createOrder("SELL-1", 101, 1, "BTCUSD", order.Sell, order.GTC, 50100, 1.0),
		createOrder("SELL-2", 102, 2, "BTCUSD", order.Sell, order.GTC, 50200, 1.5),
		createOrder("SELL-3", 103, 3, "BTCUSD", order.Sell, order.GTC, 50300, 2.0),
	}

	// Buy orders at different price levels
	buyOrders := []*order.Order{
		createOrder("BUY-1", 104, 4, "BTCUSD", order.Buy, order.GTC, 49900, 1.0),
		createOrder("BUY-2", 105, 5, "BTCUSD", order.Buy, order.GTC, 49800, 1.5),
		createOrder("BUY-3", 106, 6, "BTCUSD", order.Buy, order.GTC, 49700, 2.0),
	}

	for _, o := range sellOrders {
		eng.SubmitOrder(o)
	}
	for _, o := range buyOrders {
		eng.SubmitOrder(o)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n--- Orderbook State ---")
	printOrderBook(eng)

	// ========================================
	// Test 1: GTC (Good-Til-Cancel)
	// ========================================
	fmt.Println("\n\n=== Test 1: GTC (Good-Til-Cancel) ===")
	fmt.Println("Buy 0.8 BTC at 50100 (will match 0.8, then fully filled)")

	gtcOrder := createOrder("GTC-BUY-1", 201, 101, "BTCUSD", order.Buy, order.GTC, 50100, 0.8)
	eng.SubmitOrder(gtcOrder)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Trade 0.8 at 50100, order fully filled")

	// ========================================
	// Test 2: IOC (Immediate-Or-Cancel) - Partial Fill
	// ========================================
	fmt.Println("\n\n=== Test 2: IOC (Immediate-Or-Cancel) - Partial Fill ===")
	fmt.Println("Buy 2.0 BTC at 50200 (will match 0.2 + 1.5, cancel remaining 0.3)")

	iocOrder := createOrder("IOC-BUY-1", 202, 102, "BTCUSD", order.Buy, order.IOC, 50200, 2.0)
	eng.SubmitOrder(iocOrder)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Trade 0.2 at 50100, Trade 1.5 at 50200, Cancel remaining 0.3")

	// ========================================
	// Test 3: IOC - No Liquidity
	// ========================================
	fmt.Println("\n\n=== Test 3: IOC - No Liquidity (Price Mismatch) ===")
	fmt.Println("Buy 1.0 BTC at 50000 (price doesn't cross)")

	iocNoMatch := createOrder("IOC-BUY-2", 203, 103, "BTCUSD", order.Buy, order.IOC, 50000, 1.0)
	eng.SubmitOrder(iocNoMatch)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: REJECT - invalid_price (price doesn't cross)")

	// ========================================
	// Test 4: FOK (Fill-Or-Kill) - Success
	// ========================================
	fmt.Println("\n\n=== Test 4: FOK (Fill-Or-Kill) - Full Fill Success ===")
	fmt.Println("Buy 2.0 BTC at 50300 (enough liquidity available)")

	fokSuccess := createOrder("FOK-BUY-1", 204, 104, "BTCUSD", order.Buy, order.FOK, 50300, 2.0)
	eng.SubmitOrder(fokSuccess)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Trade 2.0 at 50300")

	// ========================================
	// Test 5: FOK - Reject (Insufficient Size)
	// ========================================
	fmt.Println("\n\n=== Test 5: FOK - Reject Insufficient Liquidity ===")
	fmt.Println("Buy 5.0 BTC at 60000 (not enough liquidity)")

	fokReject := createOrder("FOK-BUY-2", 205, 105, "BTCUSD", order.Buy, order.FOK, 60000, 5.0)
	eng.SubmitOrder(fokReject)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: REJECT - insufficient_size")

	// ========================================
	// Test 6: PostOnly - Success (Won't Match)
	// ========================================
	fmt.Println("\n\n=== Test 6: PostOnly - Success (No Immediate Match) ===")
	fmt.Println("Sell 1.0 BTC at 50500 (won't match, will rest in book)")

	postOnlySuccess := createOrder("PO-SELL-1", 206, 106, "BTCUSD", order.Sell, order.PostOnly, 50500, 1.0)
	eng.SubmitOrder(postOnlySuccess)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Order added to book at 50500")

	// ========================================
	// Test 7: PostOnly - Reject (Would Match)
	// ========================================
	fmt.Println("\n\n=== Test 7: PostOnly - Reject (Would Match Immediately) ===")
	fmt.Println("Sell 1.0 BTC at 49900 (would match with best bid)")

	postOnlyReject := createOrder("PO-SELL-2", 207, 107, "BTCUSD", order.Sell, order.PostOnly, 49900, 1.0)
	eng.SubmitOrder(postOnlyReject)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: REJECT - post_only_would_match")

	// Wait for all events to be processed
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n\n--- Final Orderbook State ---")
	printOrderBook(eng)

	fmt.Println("\n=== TIF Example Complete ===")
}

// Helper function to create orders easily
func createOrder(orderID string, id uint64, userID uint64, symbol string, side order.Side, tif order.TimeInForce, price float64, qty float64) *order.Order {
	o := order.NewOrder(
		id,
		orderID,
		fmt.Sprintf("cmd-%d", id),
		userID,
		symbol,
		side,
		order.Limit,
		tif,
		decimal.NewFromFloat(price),
		decimal.NewFromFloat(qty),
		time.Now().UnixNano(),
	)
	return &o
}

func monitorEvents(events <-chan *event.OrderBookLog) {
	for log := range events {
		switch log.LogType {
		case event.LogTypeTrade:
			fmt.Printf("  [TRADE] %s: %s BTC @ %s (taker: %s, maker: %s)\n",
				log.LogType,
				log.TradeQuantity.String(),
				log.TradePrice.String(),
				log.TakerOrderID,
				log.MakerOrderID,
			)

		case event.LogTypeCancel:
			fmt.Printf("  [CANCEL] Order %s: %s BTC @ %s\n",
				log.OrderID,
				log.RemainingSize.String(),
				log.Price.String(),
			)

		case event.LogTypeReject:
			fmt.Printf("  [REJECT] Order %s: %s - %s\n",
				log.OrderID,
				log.RejectReason,
				log.RejectDetail,
			)

		case event.LogTypeFill:
			status := "partial"
			if log.EventType == event.EventTypeFill {
				status = "full"
			}
			fmt.Printf("  [FILL] %s: Order %s filled %s BTC, remaining %s (%s)\n",
				log.Side,
				log.OrderID,
				log.FilledSize.String(),
				log.RemainingSize.String(),
				status,
			)
		}
	}
}

func printOrderBook(eng *engine.Engine) {
	bk := eng.GetOrderBook()

	fmt.Println("Asks (Sell Orders):")
	// For asks, we want to show from lowest to highest (natural Ascend order)
	bk.AskTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
		}
		return true
	})

	fmt.Println("Bids (Buy Orders):")
	// For bids, Ascend gives us highest to lowest (because tree is descending)
	bk.BidTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
		}
		return true
	})
}
