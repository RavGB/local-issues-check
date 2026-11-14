package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleReport() Report {
	rep := Report{
		Tool:    "local-issues-check",
		Started: "2026-11-13 12:00:00",
		Path:    "/tmp",
		Checks: []Result{
			{"System", StatusOK, "linux (amd64), host testbox"},
			{"Disk space", StatusWarn, "/tmp is 88.0% used - getting full"},
			{"Memory", StatusFail, "5.0% of memory is available - very low"},
			{"Load", StatusSkip, "load average is not available on this system"},
		},
	}
	rep.Summary = summarize(rep.Checks)
	return rep
}

func TestRenderJSONValid(t *testing.T) {
	out, err := renderJSON(sampleReport())
	if err != nil {
		t.Fatalf("renderJSON error: %v", err)
	}
	var decoded Report
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("rendered JSON does not parse: %v\n%s", err, out)
	}
	if decoded.Summary.Failed != 1 || decoded.Summary.Warnings != 1 || decoded.Summary.Skipped != 1 {
		t.Fatalf("decoded summary = %+v, want 1 failed, 1 warning, 1 skipped", decoded.Summary)
	}
	if len(decoded.Checks) != 4 || decoded.Checks[1].Name != "Disk space" {
		t.Fatalf("decoded checks wrong: %+v", decoded.Checks)
	}
}

func TestRenderText(t *testing.T) {
	out := renderText(sampleReport())
	for _, want := range []string{
		"Local issues check",
		"[WARN] Disk space",
		"[FAIL] Memory",
		"Summary: 4 checks, 1 failed, 1 warning(s), 1 skipped",
		"Failing checks:",
		"- Memory:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text report missing %q\n---\n%s", want, out)
		}
	}
}

func TestRunChecksOrderAndCount(t *testing.T) {
	checks := runChecks(t.TempDir())
	wantNames := []string{"System", "Disk space", "Memory", "Load", "Folder writable",
		"Temp folder writable", "Local name lookup", "Loopback", "Clock"}
	if len(checks) != len(wantNames) {
		t.Fatalf("got %d checks, want %d", len(checks), len(wantNames))
	}
	for i, name := range wantNames {
		if checks[i].Name != name {
			t.Errorf("check %d = %q, want %q", i, checks[i].Name, name)
		}
		if checks[i].Status == "" {
			t.Errorf("check %q has empty status", name)
		}
	}
}
