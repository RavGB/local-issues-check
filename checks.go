package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"time"
)

// loopbackTimeout bounds the whole loopback round trip: listening,
// dialing, and the byte exchange must all finish within it.
const loopbackTimeout = 5 * time.Second

func checkSystem() Result {
	host, err := os.Hostname()
	if err != nil {
		return Result{"System", StatusFail, fmt.Sprintf("could not read system information: %v", err)}
	}
	return Result{"System", StatusOK, fmt.Sprintf("%s (%s), Go %s, host %s, %d CPUs",
		runtime.GOOS, runtime.GOARCH, runtime.Version(), host, runtime.NumCPU())}
}

func checkDisk(path string) Result {
	fs, err := statfsFor(path)
	if err != nil {
		return Result{"Disk space", StatusSkip, fmt.Sprintf("disk figures are not available for %s: %v", path, err)}
	}
	if fs.Total == 0 {
		return Result{"Disk space", StatusSkip, fmt.Sprintf("disk figures could not be read for %s", path)}
	}
	usedPct := float64(fs.Used) / float64(fs.Total) * 100
	freePct := float64(fs.Free) / float64(fs.Total) * 100
	detail := fmt.Sprintf("%s is %.1f%% used, %s free of %s",
		path, usedPct, humanBytes(fs.Free), humanBytes(fs.Total))
	switch diskStatus(freePct) {
	case StatusFail:
		return Result{"Disk space", StatusFail, detail + " - almost full, free space now"}
	case StatusWarn:
		return Result{"Disk space", StatusWarn, detail + " - getting full"}
	default:
		return Result{"Disk space", StatusOK, detail}
	}
}

func checkMemory() Result {
	if runtime.GOOS != "linux" {
		return Result{"Memory", StatusSkip, "detailed memory figures are not available on this system"}
	}
	raw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return Result{"Memory", StatusSkip, "detailed memory figures are not available on this system"}
	}
	total, available, ok := parseMemInfo(string(raw))
	if !ok {
		return Result{"Memory", StatusSkip, "memory figures could not be read"}
	}
	pct := float64(available) / float64(total) * 100
	detail := fmt.Sprintf("%.1f%% of memory is available (%s of %s)",
		pct, humanBytes(uint64(available)*1024), humanBytes(uint64(total)*1024))
	switch memoryStatus(pct) {
	case StatusFail:
		return Result{"Memory", StatusFail, detail + " - very low, close programs or add memory"}
	case StatusWarn:
		return Result{"Memory", StatusWarn, detail + " - low"}
	default:
		return Result{"Memory", StatusOK, detail}
	}
}

func checkLoad() Result {
	if runtime.GOOS != "linux" {
		return Result{"Load", StatusSkip, "load average is not available on this system"}
	}
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return Result{"Load", StatusSkip, "load average is not available on this system"}
	}
	load1, _, _, err := parseLoadAvg(string(raw))
	if err != nil {
		return Result{"Load", StatusFail, fmt.Sprintf("could not read load average: %v", err)}
	}
	cpus := runtime.NumCPU()
	if cpus < 1 {
		cpus = 1
	}
	perCPU := load1 / float64(cpus)
	detail := fmt.Sprintf("1 minute load average %.2f on %d CPU(s), %.2f per CPU", load1, cpus, perCPU)
	switch loadStatus(perCPU) {
	case StatusFail:
		return Result{"Load", StatusFail, detail + " - machine is heavily overloaded"}
	case StatusWarn:
		return Result{"Load", StatusWarn, detail + " - busy"}
	default:
		return Result{"Load", StatusOK, detail}
	}
}

// checkWritable proves a folder is writable by creating a small temp
// file in it and deleting it again.
func checkWritable(folder, name string) Result {
	f, err := os.CreateTemp(folder, ".write-test-*")
	if err != nil {
		return Result{name, StatusFail, fmt.Sprintf("%s is not writable: %v", folder, err)}
	}
	testPath := f.Name()
	_ = f.Close()
	_ = os.Remove(testPath)
	return Result{name, StatusOK, fmt.Sprintf("%s is writable", folder)}
}

func checkHostnameLookup() Result {
	host, err := os.Hostname()
	if err != nil {
		return Result{"Local name lookup", StatusFail, fmt.Sprintf("could not read this machine's name: %v", err)}
	}
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return Result{"Local name lookup", StatusFail, fmt.Sprintf("this machine's own name (%s) does not resolve: %v", host, err)}
	}
	return Result{"Local name lookup", StatusOK, fmt.Sprintf("%s resolves to %s", host, addrs[0])}
}

// checkLoopback listens on a random loopback port, accepts the
// connection in a goroutine, and exchanges a few bytes both ways.
func checkLoopback() Result {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return Result{"Loopback", StatusFail, fmt.Sprintf("local socket test failed: %v", err)}
	}
	defer ln.Close()

	errCh := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(loopbackTimeout))
		buf := make([]byte, 4)
		if _, err := io.ReadFull(conn, buf); err != nil {
			errCh <- err
			return
		}
		_, err = conn.Write(buf)
		errCh <- err
	}()

	dialer := net.Dialer{Timeout: loopbackTimeout}
	conn, err := dialer.Dial("tcp", ln.Addr().String())
	if err != nil {
		return Result{"Loopback", StatusFail, fmt.Sprintf("local socket test failed: %v", err)}
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(loopbackTimeout))
	if _, err := conn.Write([]byte("ping")); err != nil {
		return Result{"Loopback", StatusFail, fmt.Sprintf("local socket test failed: %v", err)}
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return Result{"Loopback", StatusFail, fmt.Sprintf("local socket test failed: %v", err)}
	}
	if string(buf) != "ping" {
		return Result{"Loopback", StatusFail, "a local socket round trip returned unexpected data"}
	}
	if err := <-errCh; err != nil {
		return Result{"Loopback", StatusFail, fmt.Sprintf("local socket test failed: %v", err)}
	}
	return Result{"Loopback", StatusOK, "a local socket round trip works"}
}

func checkClock() Result {
	now := time.Now()
	_, offset := now.Zone()
	detail := fmt.Sprintf("local time %s, UTC %s, offset %s",
		now.Format("2006-01-02 15:04:05 MST"),
		now.UTC().Format("2006-01-02 15:04:05"),
		formatOffset(offset))
	return Result{"Clock", StatusOK, detail}
}
