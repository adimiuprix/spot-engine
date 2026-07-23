package snapshot

import (
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// OrderBookSnapshot contains the full state of a single OrderBook
type OrderBookSnapshot struct {
	MarketID     string                  `json:"market_id"`
	SeqID        uint64                  `json:"seq_id"`          // Current sequence ID
	LastCmdSeqID uint64                  `json:"last_cmd_seq_id"` // Last processed command sequence
	TradeID      uint64                  `json:"trade_id"`        // Current trade sequence
	Bids         []*order.Order          `json:"bids"`            // Ordered list of bids
	Asks         []*order.Order          `json:"asks"`            // Ordered list of asks
	State        protocol.OrderBookState `json:"state"`           // Market state
	MinLotSize   decimal.Decimal         `json:"min_lot_size"`    // Min trade unit
}

// SnapshotMetadata holds the global metadata for a snapshot
type SnapshotMetadata struct {
	SchemaVersion      int    `json:"schema_version"`         // Version for backward compatibility
	Timestamp          int64  `json:"timestamp"`              // Unix nano
	GlobalLastCmdSeqID uint64 `json:"global_last_cmd_seq_id"` // Global max sequence
	EngineVersion      string `json:"engine_version"`         // Engine version
	SnapshotChecksum   uint32 `json:"snapshot_checksum"`      // CRC32 of snapshot.bin
}

// SnapshotFileFooter is stored at the end of snapshot.bin
type SnapshotFileFooter struct {
	Markets []MarketSegment `json:"markets"` // Index of market segments
}

// MarketSegment contains metadata for a market's data in snapshot.bin
type MarketSegment struct {
	MarketID string `json:"market_id"`
	Offset   int64  `json:"offset"`   // Start offset in file
	Length   int64  `json:"length"`   // Length in bytes
	Checksum uint32 `json:"checksum"` // CRC32 of segment
}

// calculateFileCRC32 calculates the CRC32 checksum of a file
func calculateFileCRC32(filePath string) (uint32, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, f); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}

// calculateDataCRC32 calculates CRC32 of byte data
func calculateDataCRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
