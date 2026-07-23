package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Auto-Snapshot Example ===")
	fmt.Println("Demonstrates periodic snapshots for disaster recovery\n")

	// Setup snapshot directory
	snapshotBaseDir := "./snapshots"
	os.MkdirAll(snapshotBaseDir, 0755)

	// Setup engine
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	ctx := context.Background()

	// Create market
	fmt.Println("📊 Creating market BTC/USDT...")
	createMarket(ctx, eng)

	// Start auto-snapshot background task
	snapshotInterval := 5 * time.Second
	fmt.Printf("🔄 Starting auto-snapshot (every %s)\n", snapshotInterval)
	fmt.Println("   Keeping last 5 snapshots\n")

	go autoSnapshot(eng, snapshotBaseDir, snapshotInterval, 5)

	// Simulate trading activity
	fmt.Println("📈 Simulating trading activity...")
	fmt.Println("   (Watch snapshots being created periodically)\n")

	go simulateTrading(ctx, eng)

	// Run for 30 seconds
	time.Sleep(30 * time.Second)

	fmt.Println("\n✅ Demo complete!")
	fmt.Printf("📁 Check snapshots in: %s/\n", snapshotBaseDir)

	// Show latest snapshot
	fmt.Println("\n📊 Latest Snapshot Stats:")
	showLatestSnapshot(eng, snapshotBaseDir)
}

// autoSnapshot periodically saves snapshots to disk
func autoSnapshot(eng *engine.MatchingEngine, baseDir string, interval time.Duration, keepCount int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	snapshotCount := 0

	for range ticker.C {
		snapshotCount++
		timestamp := time.Now().Format("20060102-150405")
		targetDir := filepath.Join(baseDir, timestamp)

		// Take snapshot
		filename, err := eng.TakeSnapshotToFile(targetDir)
		if err != nil {
			fmt.Printf("[%s] ❌ Snapshot failed: %v\n", time.Now().Format("15:04:05"), err)
			continue
		}

		fmt.Printf("[%s] ✅ Snapshot #%d saved: %s (dir: %s)\n",
			time.Now().Format("15:04:05"), snapshotCount, filename, timestamp)

		// Cleanup old snapshots
		cleanupOldSnapshots(baseDir, keepCount)
	}
}

// cleanupOldSnapshots removes old snapshots, keeping only the latest N
func cleanupOldSnapshots(baseDir string, keepCount int) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}

	// Filter directories only
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	// Sort by name (timestamp format sorts correctly)
	sort.Strings(dirs)

	// Remove old snapshots
	if len(dirs) > keepCount {
		toRemove := len(dirs) - keepCount
		for i := 0; i < toRemove; i++ {
			dirPath := filepath.Join(baseDir, dirs[i])
			os.RemoveAll(dirPath)
			fmt.Printf("   🗑️  Removed old snapshot: %s\n", dirs[i])
		}
	}
}

// simulateTrading places orders periodically
func simulateTrading(ctx context.Context, eng *engine.MatchingEngine) {
	orderCount := 0

	for {
		time.Sleep(2 * time.Second)

		orderCount++
		side := "buy"
		if orderCount%2 == 0 {
			side = "sell"
		}

		price := int64(50000 + (orderCount%10)*100)
		size := 0.1 * float64(orderCount%5+1)

		req := &protocol.PlaceOrderRequest{
			BaseCommand: protocol.BaseCommand{
				CommandID: fmt.Sprintf("cmd-auto-%d", orderCount),
				UserID:    100 + uint64(orderCount%10),
				MarketID:  "BTC/USDT",
				Timestamp: time.Now().UnixNano(),
			},
			OrderID:   fmt.Sprintf("order-auto-%d", orderCount),
			Side:      side,
			OrderType: "limit",
			Price:     decimal.NewFromInt(price),
			Size:      decimal.NewFromFloat(size),
		}

		future, _ := eng.PlaceOrderAsync(ctx, req)
		result, _ := future.Wait(ctx)

		if result != nil && result.Accepted {
			fmt.Printf("[%s] 📝 Order #%d: %s %.1f BTC @ $%d\n",
				time.Now().Format("15:04:05"), orderCount, side, size, price)
		}
	}
}

// createMarket creates the trading market
func createMarket(ctx context.Context, eng *engine.MatchingEngine) {
	req := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-create",
			UserID:    1,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.00000001),
	}

	future, _ := eng.CreateMarket(ctx, req)
	future.Wait(ctx)
	fmt.Println("   ✅ Market created")
}

// showLatestSnapshot displays stats from the latest snapshot
func showLatestSnapshot(eng *engine.MatchingEngine, baseDir string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil || len(entries) == 0 {
		fmt.Println("   No snapshots found")
		return
	}

	// Find latest directory
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	if len(dirs) == 0 {
		fmt.Println("   No snapshots found")
		return
	}

	latestDir := filepath.Join(baseDir, dirs[len(dirs)-1])

	// Create new engine to read snapshot
	tempPublisher := event.NewChannelPublisher(1000)
	tempEng := engine.NewMatchingEngine(tempPublisher)
	metadata, err := tempEng.RestoreFromFile(latestDir)
	if err != nil {
		fmt.Printf("   ❌ Failed to read snapshot: %v\n", err)
		return
	}

	// Get stats
	stats, err := tempEng.GetStats("BTC/USDT")
	if err != nil {
		fmt.Printf("   ❌ Failed to read snapshot: %v\n", err)
		return
	}

	// Get stats
	stats, statsErr := tempEng.GetStats("BTC/USDT")
	if statsErr != nil {
		fmt.Printf("   ❌ Failed to get stats: %v\n", statsErr)
		return
	}

	fmt.Printf("   📁 Directory: %s\n", dirs[len(dirs)-1])
	fmt.Printf("   🔢 Sequence ID: %d\n", metadata.GlobalLastCmdSeqID)
	fmt.Printf("   🕐 Timestamp: %s\n", time.Unix(0, metadata.Timestamp).Format("2006-01-02 15:04:05"))
	fmt.Printf("   📊 Market BTC/USDT:\n")
	fmt.Printf("      - Bid Levels: %d\n", stats.BidCount)
	fmt.Printf("      - Ask Levels: %d\n", stats.AskCount)
	if stats.BestBid.GreaterThan(decimal.Zero) {
		fmt.Printf("      - Best Bid: $%s\n", stats.BestBid)
	}
	if stats.BestAsk.GreaterThan(decimal.Zero) {
		fmt.Printf("      - Best Ask: $%s\n", stats.BestAsk)
	}
}
