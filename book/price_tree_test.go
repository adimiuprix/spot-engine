package book

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestNewPriceTree tests creation of bid and ask trees
func TestNewPriceTree(t *testing.T) {
	// Test bid tree (descending)
	bidTree := NewPriceTree(true)
	if bidTree == nil {
		t.Fatal("Expected non-nil bid tree")
	}
	if !bidTree.descending {
		t.Error("Expected descending=true for bid tree")
	}

	// Test ask tree (ascending)
	askTree := NewPriceTree(false)
	if askTree == nil {
		t.Fatal("Expected non-nil ask tree")
	}
	if askTree.descending {
		t.Error("Expected descending=false for ask tree")
	}
}

// TestPriceTree_AddAndGet tests adding and retrieving price levels
func TestPriceTree_AddAndGet(t *testing.T) {
	tree := NewPriceTree(false)

	// Add price level
	price := decimal.NewFromInt(100)
	level := NewPriceLevel(price)
	tree.Add(level)

	// Get the price level
	retrieved := tree.Get(price)
	if retrieved == nil {
		t.Fatal("Expected to retrieve price level")
	}
	if !retrieved.Price.Equal(price) {
		t.Errorf("Expected price %v, got %v", price, retrieved.Price)
	}

	// Get non-existent price
	nonExistent := tree.Get(decimal.NewFromInt(999))
	if nonExistent != nil {
		t.Error("Expected nil for non-existent price")
	}
}

// TestPriceTree_Remove tests removing price levels
func TestPriceTree_Remove(t *testing.T) {
	tree := NewPriceTree(false)

	price := decimal.NewFromInt(100)
	level := NewPriceLevel(price)
	tree.Add(level)

	// Verify it's there
	if tree.Get(price) == nil {
		t.Fatal("Expected price level to exist before removal")
	}

	// Remove it
	tree.Remove(price)

	// Verify it's gone
	if tree.Get(price) != nil {
		t.Error("Expected price level to be removed")
	}
}

// TestPriceTree_Best tests getting the best price
func TestPriceTree_Best(t *testing.T) {
	tests := []struct {
		name       string
		descending bool
		prices     []int64
		wantBest   int64
	}{
		{
			name:       "Ask tree - lowest price",
			descending: false,
			prices:     []int64{105, 100, 110, 102},
			wantBest:   100,
		},
		{
			name:       "Bid tree - highest price",
			descending: true,
			prices:     []int64{95, 100, 90, 98},
			wantBest:   100,
		},
		{
			name:       "Single price",
			descending: false,
			prices:     []int64{50},
			wantBest:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPriceTree(tt.descending)

			// Add all prices
			for _, p := range tt.prices {
				price := decimal.NewFromInt(p)
				level := NewPriceLevel(price)
				tree.Add(level)
			}

			// Get best
			best := tree.Best()
			if best == nil {
				t.Fatal("Expected best price to be non-nil")
			}

			wantPrice := decimal.NewFromInt(tt.wantBest)
			if !best.Price.Equal(wantPrice) {
				t.Errorf("Expected best price %v, got %v", wantPrice, best.Price)
			}
		})
	}
}

// TestPriceTree_BestEmpty tests best price on empty tree
func TestPriceTree_BestEmpty(t *testing.T) {
	tree := NewPriceTree(false)
	best := tree.Best()
	if best != nil {
		t.Error("Expected nil best price for empty tree")
	}
}

// TestPriceTree_Len tests tree length
func TestPriceTree_Len(t *testing.T) {
	tree := NewPriceTree(false)

	if tree.Len() != 0 {
		t.Errorf("Expected length 0, got %d", tree.Len())
	}

	// Add 3 price levels
	for i := int64(1); i <= 3; i++ {
		price := decimal.NewFromInt(i * 100)
		level := NewPriceLevel(price)
		tree.Add(level)
	}

	if tree.Len() != 3 {
		t.Errorf("Expected length 3, got %d", tree.Len())
	}

	// Remove 1
	tree.Remove(decimal.NewFromInt(200))

	if tree.Len() != 2 {
		t.Errorf("Expected length 2 after removal, got %d", tree.Len())
	}
}

// TestPriceTree_Clear tests clearing the tree
func TestPriceTree_Clear(t *testing.T) {
	tree := NewPriceTree(false)

	// Add some levels
	for i := int64(1); i <= 5; i++ {
		price := decimal.NewFromInt(i * 100)
		level := NewPriceLevel(price)
		tree.Add(level)
	}

	if tree.Len() != 5 {
		t.Fatalf("Expected length 5, got %d", tree.Len())
	}

	// Clear
	tree.Clear()

	if tree.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", tree.Len())
	}

	// Best should be nil
	if tree.Best() != nil {
		t.Error("Expected nil best after clear")
	}
}

// TestPriceTree_Ascend tests iteration order
func TestPriceTree_Ascend(t *testing.T) {
	tests := []struct {
		name       string
		descending bool
		prices     []int64
		wantOrder  []int64
	}{
		{
			name:       "Ask tree - ascending order",
			descending: false,
			prices:     []int64{105, 100, 110, 102},
			wantOrder:  []int64{100, 102, 105, 110},
		},
		{
			name:       "Bid tree - descending order",
			descending: true,
			prices:     []int64{95, 100, 90, 98},
			wantOrder:  []int64{100, 98, 95, 90},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPriceTree(tt.descending)

			// Add all prices
			for _, p := range tt.prices {
				price := decimal.NewFromInt(p)
				level := NewPriceLevel(price)
				tree.Add(level)
			}

			// Collect prices in iteration order
			var gotOrder []int64
			tree.Ascend(func(level *PriceLevel) bool {
				gotOrder = append(gotOrder, level.Price.IntPart())
				return true // Continue
			})

			// Verify order
			if len(gotOrder) != len(tt.wantOrder) {
				t.Fatalf("Expected %d prices, got %d", len(tt.wantOrder), len(gotOrder))
			}

			for i, want := range tt.wantOrder {
				if gotOrder[i] != want {
					t.Errorf("Position %d: expected price %d, got %d", i, want, gotOrder[i])
				}
			}
		})
	}
}

// TestPriceTree_AscendEarlyStop tests stopping iteration early
func TestPriceTree_AscendEarlyStop(t *testing.T) {
	tree := NewPriceTree(false)

	// Add 5 prices
	for i := int64(1); i <= 5; i++ {
		price := decimal.NewFromInt(i * 100)
		level := NewPriceLevel(price)
		tree.Add(level)
	}

	// Iterate but stop after 3
	count := 0
	tree.Ascend(func(level *PriceLevel) bool {
		count++
		return count < 3 // Stop after 3
	})

	if count != 3 {
		t.Errorf("Expected to visit 3 levels, visited %d", count)
	}
}

// TestPriceTree_ReplaceOrInsert tests updating existing price level
func TestPriceTree_ReplaceOrInsert(t *testing.T) {
	tree := NewPriceTree(false)
	price := decimal.NewFromInt(100)

	// Add first level
	level1 := NewPriceLevel(price)
	level1.Volume = decimal.NewFromInt(10)
	tree.Add(level1)

	// Add second level with same price (should replace)
	level2 := NewPriceLevel(price)
	level2.Volume = decimal.NewFromInt(20)
	tree.Add(level2)

	// Should still have only 1 level
	if tree.Len() != 1 {
		t.Errorf("Expected 1 level after replace, got %d", tree.Len())
	}

	// Should have the new volume
	retrieved := tree.Get(price)
	if !retrieved.Volume.Equal(decimal.NewFromInt(20)) {
		t.Errorf("Expected volume 20, got %v", retrieved.Volume)
	}
}

// TestPriceTree_AscendGreaterOrEqual tests range iteration
func TestPriceTree_AscendGreaterOrEqual(t *testing.T) {
	tree := NewPriceTree(false)

	// Add prices: 100, 200, 300, 400, 500
	for i := int64(1); i <= 5; i++ {
		price := decimal.NewFromInt(i * 100)
		level := NewPriceLevel(price)
		tree.Add(level)
	}

	// Iterate from 250 and above
	var collected []int64
	tree.AscendGreaterOrEqual(decimal.NewFromInt(250), func(level *PriceLevel) bool {
		collected = append(collected, level.Price.IntPart())
		return true
	})

	// Should get: 300, 400, 500
	expected := []int64{300, 400, 500}
	if len(collected) != len(expected) {
		t.Fatalf("Expected %d prices, got %d", len(expected), len(collected))
	}

	for i, want := range expected {
		if collected[i] != want {
			t.Errorf("Position %d: expected %d, got %d", i, want, collected[i])
		}
	}
}

// TestPriceTree_DescendLessOrEqual tests descending range iteration
func TestPriceTree_DescendLessOrEqual(t *testing.T) {
	tree := NewPriceTree(false) // Ascending tree for asks

	// Add prices: 100, 200, 300, 400, 500
	for i := int64(1); i <= 5; i++ {
		price := decimal.NewFromInt(i * 100)
		level := NewPriceLevel(price)
		tree.Add(level)
	}

	// Iterate from 350 and below (in descending order)
	var collected []int64
	tree.DescendLessOrEqual(decimal.NewFromInt(350), func(level *PriceLevel) bool {
		collected = append(collected, level.Price.IntPart())
		return true
	})

	// Should get: 300, 200, 100 (descending order)
	expected := []int64{300, 200, 100}
	if len(collected) != len(expected) {
		t.Fatalf("Expected %d prices, got %d", len(expected), len(collected))
	}

	for i, want := range expected {
		if collected[i] != want {
			t.Errorf("Position %d: expected %d, got %d", i, want, collected[i])
		}
	}
}
