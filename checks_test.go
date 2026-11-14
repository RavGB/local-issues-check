package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckWritableTempDir(t *testing.T) {
	dir := t.TempDir()
	res := checkWritable(dir, "Folder writable")
	if res.Status != StatusOK {
		t.Fatalf("status = %s, want OK (detail: %s)", res.Status, res.Detail)
	}
	// The check must clean up after itself.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("writable check left files behind: %v", entries)
	}
}

func TestCheckWritableMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	res := checkWritable(missing, "Folder writable")
	if res.Status != StatusFail {
		t.Fatalf("status = %s, want FAIL (detail: %s)", res.Status, res.Detail)
	}
}

func TestCheckLoopback(t *testing.T) {
	res := checkLoopback()
	if res.Status != StatusOK {
		t.Fatalf("status = %s, want OK (detail: %s)", res.Status, res.Detail)
	}
}

func TestCheckClock(t *testing.T) {
	res := checkClock()
	if res.Status != StatusOK {
		t.Fatalf("status = %s, want OK", res.Status)
	}
	if !strings.Contains(res.Detail, "UTC") || !strings.Contains(res.Detail, "offset") {
		t.Fatalf("clock detail missing pieces: %s", res.Detail)
	}
}

func TestCheckSystem(t *testing.T) {
	res := checkSystem()
	if res.Status != StatusOK {
		t.Fatalf("status = %s, want OK (detail: %s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "host") {
		t.Fatalf("system detail missing hostname: %s", res.Detail)
	}
}
