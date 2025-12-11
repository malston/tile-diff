// ABOUTME: Disk space checking and management utilities.
// ABOUTME: Ensures sufficient disk space before downloads with configurable buffer.
package pivnet

import (
	"fmt"
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
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
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
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
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
