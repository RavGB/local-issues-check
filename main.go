// local-issues-check checks a local machine for common issues and
// prints a short, plain report: system information, disk space,
// available memory, load, writable folders, own-name resolution, a
// loopback round trip, and the clock.
//
// It runs fully on the machine you run it on, uses only the Go
// standard library, and sends nothing anywhere.
//
// Exit codes: 0 = no failing checks (warnings are allowed),
// 1 = at least one check failed, 2 = usage or setup error.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Status values for a single check.
const (
	StatusOK   = "OK"
	StatusWarn = "WARN"
	StatusFail = "FAIL"
	StatusSkip = "SKIP"
)

// Result is the outcome of one check.
type Result struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// Report is the full set of check results plus a summary.
type Report struct {
	Tool    string   `json:"tool"`
	Started string   `json:"started"`
	Path    string   `json:"path"`
	Checks  []Result `json:"checks"`
	Summary Summary  `json:"summary"`
}

// Summary counts results by status.
type Summary struct {
	Total    int `json:"total"`
	Failed   int `json:"failed"`
	Warnings int `json:"warnings"`
	Skipped  int `json:"skipped"`
}

func summarize(checks []Result) Summary {
	s := Summary{Total: len(checks)}
	for _, c := range checks {
		switch c.Status {
		case StatusFail:
			s.Failed++
		case StatusWarn:
			s.Warnings++
		case StatusSkip:
			s.Skipped++
		}
	}
	return s
}

// runChecks runs every check concurrently. Each goroutine writes only
// to its own numbered slot, so the report order is always the same:
// system, disk, memory, load, folder writable, temp writable, name
// lookup, loopback, clock.
func runChecks(path string) []Result {
	results := make([]Result, 9)
	var wg sync.WaitGroup
	run := func(slot int, fn func() Result) {
		defer wg.Done()
		results[slot] = fn()
	}
	wg.Add(9)
	go run(0, checkSystem)
	go run(1, func() Result { return checkDisk(path) })
	go run(2, checkMemory)
	go run(3, checkLoad)
	go run(4, func() Result { return checkWritable(path, "Folder writable") })
	go run(5, func() Result { return checkWritable(os.TempDir(), "Temp folder writable") })
	go run(6, checkHostnameLookup)
	go run(7, checkLoopback)
	go run(8, checkClock)
	wg.Wait()
	return results
}

// renderText formats the report the way a person reads it.
func renderText(r Report) string {
	out := "Local issues check\n"
	out += "RavGB - hello@ravgb.com\n"
	out += fmt.Sprintf("Started: %s\n", r.Started)
	out += fmt.Sprintf("Folder tested: %s\n", r.Path)
	out += "------------------------------------------------------------\n"
	for _, c := range r.Checks {
		out += fmt.Sprintf("[%-4s] %s: %s\n", c.Status, c.Name, c.Detail)
	}
	out += "------------------------------------------------------------\n"
	out += fmt.Sprintf("Summary: %d checks, %d failed, %d warning(s), %d skipped\n",
		r.Summary.Total, r.Summary.Failed, r.Summary.Warnings, r.Summary.Skipped)
	if r.Summary.Failed > 0 {
		out += "Failing checks:\n"
		for _, c := range r.Checks {
			if c.Status == StatusFail {
				out += fmt.Sprintf("  - %s: %s\n", c.Name, c.Detail)
			}
		}
	} else {
		out += "No failing checks.\n"
	}
	return out
}

// renderJSON formats the report as indented JSON.
func renderJSON(r Report) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func run(pathFlag, outputFlag string, jsonFlag bool) int {
	abs, err := filepath.Abs(pathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve folder %q: %v\n", pathFlag, err)
		return 2
	}
	info, err := os.Stat(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read folder %q: %v\n", abs, err)
		return 2
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%q is not a folder\n", abs)
		return 2
	}

	rep := Report{
		Tool:    "local-issues-check",
		Started: time.Now().Format("2006-01-02 15:04:05"),
		Path:    abs,
		Checks:  runChecks(abs),
	}
	rep.Summary = summarize(rep.Checks)

	var text string
	if jsonFlag {
		text, err = renderJSON(rep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not render the report: %v\n", err)
			return 2
		}
	} else {
		text = renderText(rep)
	}
	fmt.Print(text)

	if outputFlag != "" {
		if err := os.WriteFile(outputFlag, []byte(text), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Could not save the report: %v\n", err)
			return 2
		}
		fmt.Printf("Report saved to %s\n", outputFlag)
	}

	if rep.Summary.Failed > 0 {
		return 1
	}
	return 0
}

func main() {
	path := flag.String("path", ".", "folder to test for disk space and writability")
	output := flag.String("output", "", "also save the report to this file")
	asJSON := flag.Bool("json", false, "print the report as JSON instead of text")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Unexpected arguments: %v\n", flag.Args())
		os.Exit(2)
	}
	os.Exit(run(*path, *output, *asJSON))
}
