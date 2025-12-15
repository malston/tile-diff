//go:build windows

// ABOUTME: Disk space checking and management utilities for Windows.
// ABOUTME: Ensures sufficient disk space before downloads with configurable buffer.
package pivnet

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
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

	availableBytes, err := d.getAvailableSpaceWindows(checkPath)
	if err != nil {
		return false, err
	}

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

	availableBytes, err := d.getAvailableSpaceWindows(checkPath)
	if err != nil {
		return 0, err
	}

	return int64(availableBytes), nil
}

// getAvailableSpaceWindows uses Windows API to get disk space
func (d *DiskManager) getAvailableSpaceWindows(path string) (uint64, error) {
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to convert path: %w", err)
	}

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		(*uint64)(unsafe.Pointer(&freeBytesAvailable)),
		(*uint64)(unsafe.Pointer(&totalBytes)),
		(*uint64)(unsafe.Pointer(&totalFreeBytes)),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to check disk space: %w", err)
	}

	return freeBytesAvailable, nil
}

// calculateRequired calculates total space needed (file + buffer)
func (d *DiskManager) calculateRequired(fileSize int64) int64 {
	bufferBytes := d.minFreeSpaceGB * 1024 * 1024 * 1024
	return fileSize + bufferBytes
}
