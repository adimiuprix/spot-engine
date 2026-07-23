package snapshot

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Reader handles snapshot reading
type Reader struct {
	snapshotDir string
}

// NewReader creates a new snapshot reader
func NewReader(snapshotDir string) *Reader {
	return &Reader{
		snapshotDir: snapshotDir,
	}
}

// ReadSnapshot reads and validates a complete snapshot
func (r *Reader) ReadSnapshot() (*SnapshotMetadata, []*OrderBookSnapshot, error) {
	// Read and validate metadata
	metadataPath := filepath.Join(r.snapshotDir, MetadataFile)
	metadata, err := r.readMetadata(metadataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Validate snapshot.bin checksum
	snapshotPath := filepath.Join(r.snapshotDir, SnapshotBinFile)
	actualChecksum, err := calculateFileCRC32(snapshotPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != metadata.SnapshotChecksum {
		return nil, nil, fmt.Errorf("checksum mismatch: expected %d, got %d", 
			metadata.SnapshotChecksum, actualChecksum)
	}

	// Read snapshot.bin with footer
	snapshots, err := r.readSnapshotBin(snapshotPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read snapshot.bin: %w", err)
	}

	return metadata, snapshots, nil
}

// readMetadata reads and parses metadata.json
func (r *Reader) readMetadata(path string) (*SnapshotMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata SnapshotMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// readSnapshotBin reads snapshot.bin and returns all market snapshots
func (r *Reader) readSnapshotBin(path string) ([]*OrderBookSnapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	// Read footer length (last 4 bytes)
	if fileSize < 4 {
		return nil, fmt.Errorf("file too small: %d bytes", fileSize)
	}

	if _, err := f.Seek(fileSize-4, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to footer length: %w", err)
	}

	var footerLen uint32
	if err := binary.Read(f, binary.LittleEndian, &footerLen); err != nil {
		return nil, fmt.Errorf("failed to read footer length: %w", err)
	}

	// Validate footer bounds
	footerStart := fileSize - int64(footerLen) - 4
	if footerStart < 0 || footerStart > fileSize {
		return nil, fmt.Errorf("invalid footer bounds: start=%d, size=%d", footerStart, fileSize)
	}

	// Read footer
	if _, err := f.Seek(footerStart, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to footer: %w", err)
	}

	footerData := make([]byte, footerLen)
	if _, err := f.Read(footerData); err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}

	var footer SnapshotFileFooter
	if err := json.Unmarshal(footerData, &footer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal footer: %w", err)
	}

	// Read each market segment
	var snapshots []*OrderBookSnapshot
	for _, segment := range footer.Markets {
		// Validate segment bounds
		if segment.Offset < 0 || segment.Offset+segment.Length > footerStart {
			return nil, fmt.Errorf("invalid segment bounds for %s", segment.MarketID)
		}

		// Read segment data
		if _, err := f.Seek(segment.Offset, 0); err != nil {
			return nil, fmt.Errorf("failed to seek to segment %s: %w", segment.MarketID, err)
		}

		segmentData := make([]byte, segment.Length)
		if _, err := f.Read(segmentData); err != nil {
			return nil, fmt.Errorf("failed to read segment %s: %w", segment.MarketID, err)
		}

		// Verify segment checksum
		actualChecksum := calculateDataCRC32(segmentData)
		if actualChecksum != segment.Checksum {
			return nil, fmt.Errorf("checksum mismatch for %s: expected %d, got %d",
				segment.MarketID, segment.Checksum, actualChecksum)
		}

		// Unmarshal snapshot
		var snap OrderBookSnapshot
		if err := json.Unmarshal(segmentData, &snap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot for %s: %w", segment.MarketID, err)
		}

		snapshots = append(snapshots, &snap)
	}

	return snapshots, nil
}
