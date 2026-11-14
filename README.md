# Local Issues Check

A small tool from RavGB that checks a local machine for common
issues and prints a short, plain report.

It runs fully on the machine you run it on and sends nothing
anywhere. It is written in Go using only the standard library, so
the one download is the tool itself: no runtime to install.

Contact: hello@ravgb.com
Site: https://ravgb.com

## Build

You need Go installed. Then:

```bash
go build -o local-issues-check .
```

That produces one binary, `local-issues-check`, which you can copy
anywhere and run.

## Run

```bash
./local-issues-check
```

Check a different folder's disk space and writability:

```bash
./local-issues-check --path /some/folder
```

Save the report to a file as well as printing it:

```bash
./local-issues-check --output report.txt
```

Print the report as JSON instead of text (handy for other tools).
If `--output` is set too, the file gets the same JSON:

```bash
./local-issues-check --json
./local-issues-check --json --output report.json
```

## Flags

| Flag | Default | What it does |
| --- | --- | --- |
| `--path` | `.` (current folder) | Folder to test for disk space and writability |
| `--output` | (none) | Also save the report to this file |
| `--json` | off | Print the report as JSON instead of text |

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | No check failed (warnings are allowed) |
| 1 | At least one check failed |
| 2 | Usage or setup error, for example the `--path` folder does not exist, or the report file could not be written |

The exit code makes the tool easy to call from other scripts.

## What it checks

The checks run at the same time, and the report always lists them
in this order. Each one comes back OK, WARN, FAIL, or SKIP (SKIP
means the figure is not available on this system; it never counts
as a failure).

- **System** — operating system, architecture, Go version,
  hostname, and CPU count. Report only.
- **Disk space** — total, used, and free space for the `--path`
  folder. FAIL when under 5% is free, WARN under 15%.
- **Memory** — total and available memory from `/proc/meminfo`
  (Linux). FAIL when under 10% is available, WARN under 20%. On
  other systems this reports SKIP.
- **Load** — the 1-minute load average from `/proc/loadavg`
  (Linux), compared with the CPU count. FAIL above 2.0 per CPU,
  WARN above 1.0 per CPU. On other systems this reports SKIP.
- **Folder writable** and **Temp folder writable** — proves each
  folder is writable by creating a small file there and deleting
  it again. FAIL if either cannot be written to.
- **Local name lookup** — the machine's own hostname must resolve.
  FAIL when it does not; a name that does not resolve breaks many
  local services.
- **Loopback** — opens a listening socket on 127.0.0.1, connects
  to it, and exchanges a few bytes both ways, all within 5
  seconds. FAIL on any error or timeout.
- **Clock** — local time, UTC time, and the zone offset. Report
  only.

The report ends with a summary counting the checks, the failures,
and the warnings, and names the failing checks when there are any.

## Example output

A real run on a Linux machine:

```text
Local issues check
RavGB - hello@ravgb.com
Started: 2026-11-14 00:32:16
Folder tested: /home/hatch/workspace/ravgb-coding/local-issues-check
------------------------------------------------------------
[OK  ] System: linux (amd64), Go go1.27.4, host htch-runtime, 4 CPUs
[OK  ] Disk space: /home/hatch/workspace/ravgb-coding/local-issues-check is 2.8% used, 242.13 GiB free of 249.98 GiB
[WARN] Memory: 11.9% of memory is available (941.07 MiB of 7.75 GiB) - low
[OK  ] Load: 1 minute load average 3.60 on 4 CPU(s), 0.90 per CPU
[OK  ] Folder writable: /home/hatch/workspace/ravgb-coding/local-issues-check is writable
[OK  ] Temp folder writable: /tmp is writable
[OK  ] Local name lookup: htch-runtime resolves to 127.0.1.1
[OK  ] Loopback: a local socket round trip works
[OK  ] Clock: local time 2026-11-14 00:32:16 UTC, UTC 2026-11-14 00:32:16, offset +00:00
------------------------------------------------------------
Summary: 9 checks, 0 failed, 1 warning(s), 0 skipped
No failing checks.
```

## License

MIT License. See the LICENSE file in this repository.

Copyright (c) 2026 RavGB Technologies Inc.
