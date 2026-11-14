package main

import (
	"fmt"
	"strconv"
	"strings"
)

// parseMemInfo extracts MemTotal and MemAvailable (in KiB) from the
// contents of Linux /proc/meminfo. Lines look like:
//
//	MemTotal:       16384000 kB
//	MemAvailable:    8192000 kB
//
// ok is false when either value is missing or the total is not
// positive.
func parseMemInfo(contents string) (totalKiB, availableKiB int64, ok bool) {
	var total, available int64
	var haveTotal, haveAvailable bool
	for _, line := range strings.Split(contents, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			total, haveTotal = v, true
		case "MemAvailable:":
			available, haveAvailable = v, true
		}
	}
	if !haveTotal || !haveAvailable || total <= 0 {
		return 0, 0, false
	}
	return total, available, true
}

// memoryStatus maps the available-memory percentage to a status:
// under 10% available fails, under 20% warns.
func memoryStatus(availablePct float64) string {
	switch {
	case availablePct < 10:
		return StatusFail
	case availablePct < 20:
		return StatusWarn
	default:
		return StatusOK
	}
}

// parseLoadAvg extracts the 1, 5 and 15 minute load averages from the
// contents of Linux /proc/loadavg, whose full form is:
//
//	0.52 0.31 0.20 1/234 5678
func parseLoadAvg(contents string) (load1, load5, load15 float64, err error) {
	fields := strings.Fields(contents)
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("expected at least 3 fields, got %d", len(fields))
	}
	vals := make([]float64, 3)
	for i := 0; i < 3; i++ {
		vals[i], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("field %d: %w", i, err)
		}
	}
	return vals[0], vals[1], vals[2], nil
}

// loadStatus maps the 1-minute load per CPU to a status: above 2.0
// per CPU fails, above 1.0 per CPU warns.
func loadStatus(loadPerCPU float64) string {
	switch {
	case loadPerCPU > 2.0:
		return StatusFail
	case loadPerCPU > 1.0:
		return StatusWarn
	default:
		return StatusOK
	}
}

// diskStatus maps the free-space percentage to a status: under 5%
// free fails, under 15% free warns.
func diskStatus(freePct float64) string {
	switch {
	case freePct < 5:
		return StatusFail
	case freePct < 15:
		return StatusWarn
	default:
		return StatusOK
	}
}

// humanBytes formats a byte count with a binary unit, for example
// "242.12 GiB".
func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// formatOffset formats a UTC offset in seconds as +hh:mm or -hh:mm.
func formatOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	return fmt.Sprintf("%s%02d:%02d", sign, seconds/3600, (seconds%3600)/60)
}
