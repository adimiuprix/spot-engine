package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

func runManagementExample() {
	fmt.Println("=== Management Commands Example ===\n")

	// Create event publisher
	publisher := event.NewChannelPublisher(10000)

	// Create matching engine
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	// Start event listener
	go func() {
		for log := range publisher.Channel() {
			switch log.LogType {
			case event.LogTypeAdmin:
				fmt.Printf("[ADMIN] %s: Market %s | %s -> %s\n",
					log.EventType,
					log.MarketID,
					log.OldState,
					log.NewState,
				)
				if len(log.ConfigChanges) > 0 {
					fmt.Printf("  Config changes: %+v\n", log.ConfigChanges)
				}
				if log.AdminReason != "" {
					fmt.Printf("  Reason: %s\n", log.AdminReason)
				}
			case event.LogTypeReject:
				fmt.Printf("[REJECT] Market %s: %s - %s\n",
					log.MarketID,
					log.RejectReason,
					log.RejectDetail,
				)
			}
		}
	}()

	ctx := context.Background()
	now := time.Now().UnixNano()

	// 1. Create BTC-USDT market
	fmt.Println("1. Creating BTC-USDT market...")
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdCreateMarket,
			CommandID: "cmd-create-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now,
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}

	future, err := eng.CreateMarket(ctx, createReq)
	if err != nil {
		fmt.Printf("Error creating market: %v\n", err)
		return
	}

	success, err := future.Wait(ctx)
	if err != nil {
		fmt.Printf("Create market failed: %v\n", err)
	} else if success {
		fmt.Println("✓ Market created successfully\n")
	}

	time.Sleep(100 * time.Millisecond)

	// 2. Get market stats
	fmt.Println("2. Getting market stats...")
	stats, err := eng.GetStats("BTC-USDT")
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
	} else {
		fmt.Printf("   State: %s\n", stats.State)
		fmt.Printf("   MinLotSize: %s\n", stats.MinLotSize.String())
		fmt.Printf("   Bid Levels: %d\n", stats.BidCount)
		fmt.Printf("   Ask Levels: %d\n\n", stats.AskCount)
	}

	// 3. Suspend market
	fmt.Println("3. Suspending market...")
	suspendReq := &protocol.SuspendMarketRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdSuspendMarket,
			CommandID: "cmd-suspend-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now + 1,
		},
		Reason: "Maintenance",
	}

	suspendFuture, err := eng.SuspendMarket(ctx, suspendReq)
	if err != nil {
		fmt.Printf("Error suspending market: %v\n", err)
	}

	success, err = suspendFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("Suspend market failed: %v\n", err)
	} else if success {
		fmt.Println("✓ Market suspended\n")
	}

	time.Sleep(100 * time.Millisecond)

	// 4. Check state after suspension
	stats, _ = eng.GetStats("BTC-USDT")
	fmt.Printf("   Current state: %s\n\n", stats.State)

	// 5. Resume market
	fmt.Println("4. Resuming market...")
	resumeReq := &protocol.ResumeMarketRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdResumeMarket,
			CommandID: "cmd-resume-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now + 2,
		},
	}

	resumeFuture, err := eng.ResumeMarket(ctx, resumeReq)
	if err != nil {
		fmt.Printf("Error resuming market: %v\n", err)
	}

	success, err = resumeFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("Resume market failed: %v\n", err)
	} else if success {
		fmt.Println("✓ Market resumed\n")
	}

	time.Sleep(100 * time.Millisecond)

	// 6. Update config
	fmt.Println("5. Updating market config...")
	updateReq := &protocol.UpdateConfigRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdUpdateConfig,
			CommandID: "cmd-update-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now + 3,
		},
		MinLotSize: decimal.NewFromFloat(0.01),
	}

	updateFuture, err := eng.UpdateConfig(ctx, updateReq)
	if err != nil {
		fmt.Printf("Error updating config: %v\n", err)
	}

	success, err = updateFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("Update config failed: %v\n", err)
	} else if success {
		fmt.Println("✓ Config updated\n")
	}

	time.Sleep(100 * time.Millisecond)

	// 7. Try to create duplicate market (should fail)
	fmt.Println("6. Trying to create duplicate market...")
	duplicateReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdCreateMarket,
			CommandID: "cmd-create-2",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now + 4,
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}

	dupFuture, err := eng.CreateMarket(ctx, duplicateReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		_, err = dupFuture.Wait(ctx)
		if err != nil {
			fmt.Println("✓ Duplicate creation rejected as expected\n")
		}
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Println("=== Management Example Complete ===")
}
