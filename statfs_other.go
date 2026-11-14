//go:build !linux && !darwin

package main

import "errors"

// On platforms without a statfs implementation here, the disk check
// reports itself unavailable instead of failing.
func statfsFor(path string) (fsStats, error) {
	return fsStats{}, errors.New("disk figures are not available on this platform")
}
