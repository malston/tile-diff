//go:build unix

// ABOUTME: Disk space checking and management utilities.
// ABOUTME: Ensures sufficient disk space before downloads with configurable buffer.
package pivnet

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// DiskManager handles disk space checks
type DiskManager struct {
	minFreeSpaceGB int64
}

// NewDiskManager creates a new disk space manager
func NewDiskManager(minFreeSpaceGB int64) *DiskManager {
	return &DiskManager{
		minFreeSpaceGB: minFreeSpaceGB,
	}
}

// HasEnoughSpace checks if there's enough disk space for a download
func (d *DiskManager) HasEnoughSpace(path string, fileSize int64) (bool, error) {
	// If path doesn't exist, check parent directory
	checkPath := path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		checkPath = filepath.Dir(path)
		// Keep going up until we find an existing directory
		for {
			if _, err := os.Stat(checkPath); err == nil {
				break
			}
			parent := filepath.Dir(checkPath)
			if parent == checkPath {
				// Reached root without finding existing dir
				return false, fmt.Errorf("no existing directory found in path hierarchy")
			}
			checkPath = parent
		}
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(checkPath, &stat)
	if err != nil {
		return false, fmt.Errorf("failed to check disk space: %w", err)
	}

	// Available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// Required space: file size + minimum buffer
	requiredBytes := d.calculateRequired(fileSize)

	return availableBytes >= uint64(requiredBytes), nil
}

// GetAvailableSpace returns available space in bytes
func (d *DiskManager) GetAvailableSpace(path string) (int64, error) {
	// If path doesn't exist, check parent directory
	checkPath := path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		checkPath = filepath.Dir(path)
		for {
			if _, err := os.Stat(checkPath); err == nil {
				break
			}
			parent := filepath.Dir(checkPath)
			if parent == checkPath {
				return 0, fmt.Errorf("no existing directory found in path hierarchy")
			}
			checkPath = parent
		}
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(checkPath, &stat)
	if err != nil {
		return 0, fmt.Errorf("failed to check disk space: %w", err)
	}

	return int64(stat.Bavail * uint64(stat.Bsize)), nil
}

// calculateRequired calculates total space needed (file + buffer)
func (d *DiskManager) calculateRequired(fileSize int64) int64 {
	bufferBytes := d.minFreeSpaceGB * 1024 * 1024 * 1024
	return fileSize + bufferBytes
}
