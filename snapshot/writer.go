package snapshot

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	SchemaVersion   = 1
	EngineVersion   = "v0.8.0"
	TempDirPrefix   = "snapshot-tmp-"
	SnapshotBinFile = "snapshot.bin"
	MetadataFile    = "metadata.json"
)

// Writer handles snapshot writing
type Writer struct {
	targetDir string
}

// NewWriter creates a new snapshot writer
func NewWriter(targetDir string) *Writer {
	return &Writer{
		targetDir: targetDir,
	}
}

// WriteSnapshot writes a complete snapshot atomically
func (w *Writer) WriteSnapshot(snapshots []*OrderBookSnapshot, globalSeqID uint64) error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp(filepath.Dir(w.targetDir), TempDirPrefix)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up on error

	// Write snapshot.bin with footer
	snapshotPath := filepath.Join(tempDir, SnapshotBinFile)
	_, err = w.writeSnapshotBin(snapshotPath, snapshots)
	if err != nil {
		return fmt.Errorf("failed to write snapshot.bin: %w", err)
	}

	// Calculate checksum of snapshot.bin
	checksum, err := calculateFileCRC32(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Write metadata.json
	metadata := &SnapshotMetadata{
		SchemaVersion:      SchemaVersion,
		Timestamp:          time.Now().UnixNano(),
		GlobalLastCmdSeqID: globalSeqID,
		EngineVersion:      EngineVersion,
		SnapshotChecksum:   checksum,
	}

	metadataPath := filepath.Join(tempDir, MetadataFile)
	if err := w.writeMetadata(metadataPath, metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Atomic rename: move temp dir to target dir
	if err := os.RemoveAll(w.targetDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old snapshot: %w", err)
	}

	if err := os.Rename(tempDir, w.targetDir); err != nil {
		return fmt.Errorf("failed to rename temp dir: %w", err)
	}

	return nil
}

// writeSnapshotBin writes snapshot data and footer to snapshot.bin
func (w *Writer) writeSnapshotBin(path string, snapshots []*OrderBookSnapshot) ([]MarketSegment, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var segments []MarketSegment
	currentOffset := int64(0)

	// Write each market snapshot
	for _, snap := range snapshots {
		data, err := json.Marshal(snap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal snapshot for %s: %w", snap.MarketID, err)
		}

		// Write data
		n, err := f.Write(data)
		if err != nil {
			return nil, fmt.Errorf("failed to write snapshot for %s: %w", snap.MarketID, err)
		}

		// Record segment
		segments = append(segments, MarketSegment{
			MarketID: snap.MarketID,
			Offset:   currentOffset,
			Length:   int64(n),
			Checksum: calculateDataCRC32(data),
		})

		currentOffset += int64(n)
	}

	// Write footer
	footer := SnapshotFileFooter{
		Markets: segments,
	}

	footerData, err := json.Marshal(footer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal footer: %w", err)
	}

	// Write footer JSON
	if _, err := f.Write(footerData); err != nil {
		return nil, fmt.Errorf("failed to write footer: %w", err)
	}

	// Write footer length (4 bytes) at the end
	footerLen := uint32(len(footerData))
	if err := binary.Write(f, binary.LittleEndian, footerLen); err != nil {
		return nil, fmt.Errorf("failed to write footer length: %w", err)
	}

	return segments, nil
}

// writeMetadata writes metadata.json
func (w *Writer) writeMetadata(path string, metadata *SnapshotMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
