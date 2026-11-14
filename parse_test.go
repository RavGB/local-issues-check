package main

import "testing"

func TestParseMemInfo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTotal int64
		wantAvail int64
		wantOK    bool
	}{
		{
			name:      "normal meminfo",
			input:     "MemTotal:       16384000 kB\nMemFree:         1024000 kB\nMemAvailable:    8192000 kB\nBuffers:          512000 kB\n",
			wantTotal: 16384000, wantAvail: 8192000, wantOK: true,
		},
		{
			name:   "missing MemAvailable",
			input:  "MemTotal:       16384000 kB\nMemFree:         1024000 kB\n",
			wantOK: false,
		},
		{
			name:   "missing MemTotal",
			input:  "MemAvailable:    8192000 kB\n",
			wantOK: false,
		},
		{
			name:   "zero total",
			input:  "MemTotal:                0 kB\nMemAvailable:          0 kB\n",
			wantOK: false,
		},
		{
			name:   "garbage lines are ignored",
			input:  "nonsense\nMemTotal: notanumber kB\nMemAvailable:    8192000 kB\n",
			wantOK: false,
		},
		{name: "empty input", input: "", wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, avail, ok := parseMemInfo(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && (total != tc.wantTotal || avail != tc.wantAvail) {
				t.Fatalf("got (%d, %d), want (%d, %d)", total, avail, tc.wantTotal, tc.wantAvail)
			}
		})
	}
}

func TestMemoryStatus(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, StatusFail},
		{9.9, StatusFail},
		{10, StatusWarn},
		{19.9, StatusWarn},
		{20, StatusOK},
		{75, StatusOK},
	}
	for _, tc := range tests {
		if got := memoryStatus(tc.pct); got != tc.want {
			t.Errorf("memoryStatus(%.1f) = %s, want %s", tc.pct, got, tc.want)
		}
	}
}

func TestParseLoadAvg(t *testing.T) {
	l1, l5, l15, err := parseLoadAvg("0.52 0.31 0.20 1/234 5678\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l1 != 0.52 || l5 != 0.31 || l15 != 0.20 {
		t.Fatalf("got (%.2f, %.2f, %.2f)", l1, l5, l15)
	}
	for _, bad := range []string{"", "0.52 0.31", "a b c 1/2 3"} {
		if _, _, _, err := parseLoadAvg(bad); err == nil {
			t.Errorf("parseLoadAvg(%q): expected error, got none", bad)
		}
	}
}

func TestLoadStatus(t *testing.T) {
	tests := []struct {
		perCPU float64
		want   string
	}{
		{0.5, StatusOK},
		{1.0, StatusOK},
		{1.5, StatusWarn},
		{2.0, StatusWarn},
		{2.5, StatusFail},
	}
	for _, tc := range tests {
		if got := loadStatus(tc.perCPU); got != tc.want {
			t.Errorf("loadStatus(%.2f) = %s, want %s", tc.perCPU, got, tc.want)
		}
	}
}

func TestDiskStatus(t *testing.T) {
	tests := []struct {
		freePct float64
		want    string
	}{
		{0, StatusFail},
		{4.9, StatusFail},
		{5, StatusWarn},
		{14.9, StatusWarn},
		{15, StatusOK},
		{90, StatusOK},
	}
	for _, tc := range tests {
		if got := diskStatus(tc.freePct); got != tc.want {
			t.Errorf("diskStatus(%.1f) = %s, want %s", tc.freePct, got, tc.want)
		}
	}
}

func TestFormatOffset(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "+00:00"},
		{-6 * 3600, "-06:00"},
		{5*3600 + 30*60, "+05:30"},
		{-4*3600 - 30*60, "-04:30"},
	}
	for _, tc := range tests {
		if got := formatOffset(tc.seconds); got != tc.want {
			t.Errorf("formatOffset(%d) = %s, want %s", tc.seconds, got, tc.want)
		}
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KiB"},
		{1536, "1.50 KiB"},
		{1024 * 1024, "1.00 MiB"},
		{268435456000, "250.00 GiB"},
	}
	for _, tc := range tests {
		if got := humanBytes(tc.n); got != tc.want {
			t.Errorf("humanBytes(%d) = %s, want %s", tc.n, got, tc.want)
		}
	}
}

func TestSummarize(t *testing.T) {
	checks := []Result{
		{"A", StatusOK, ""},
		{"B", StatusWarn, ""},
		{"C", StatusFail, ""},
		{"D", StatusFail, ""},
		{"E", StatusSkip, ""},
	}
	s := summarize(checks)
	if s.Total != 5 || s.Failed != 2 || s.Warnings != 1 || s.Skipped != 1 {
		t.Fatalf("summarize = %+v, want total 5, failed 2, warnings 1, skipped 1", s)
	}
}
