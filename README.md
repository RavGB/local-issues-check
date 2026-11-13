# Local Issues Check

A small script from RavGB that checks a local machine for common
issues and prints a short, plain report.

It runs fully on the machine you run it on. Nothing is sent anywhere,
and it uses only the Python standard library, so there is nothing to
install beyond Python 3.

Contact: hello@ravgb.com
Site: https://ravgb.com

## What it checks

- System information and Python version
- Disk space on the folder you point it at
- Available memory (on Linux)
- Load average compared with the number of cores
- That the folder and the temp folder are writable
- That the machine's own name resolves
- That a local socket round trip (loopback) works
- The local clock and its UTC offset

Each check comes back OK, WARN, FAIL, or SKIP, and the script ends
with a summary. The exit code is 0 when nothing failed, and 1 when at
least one check failed, so it can also be used in other scripts.

## How to run it

```bash
python3 check_local.py
```

Save the report to a file as well:

```bash
python3 check_local.py --output report.txt
```

Test a different folder's disk space and writability:

```bash
python3 check_local.py --path /some/folder
```

## License

MIT License. See the LICENSE file in this repository.

Copyright (c) 2026 RavGB Technologies Inc.

## Example output

```text
Local issues check
RavGB - hello@ravgb.com
----------------------------------------------------------------
[OK  ] System: Linux 6.8.0 (x86_64), Python 3.12.3, host my-laptop
[OK  ] Disk space: /home is 54.0% used, 121.4 GB free
[OK  ] Memory: 48.2% of memory is available
[OK  ] Load: 1 minute load average 0.31 on 8 core(s)
[OK  ] Folder writable: /home is writable
[OK  ] Temp folder writable: /tmp is writable
[OK  ] Local name lookup: my-laptop resolves to 127.0.1.1
[OK  ] Loopback: a local socket round trip works
[OK  ] Clock: local time 2026-11-13 10:00:00 CST, offset -1 day, 18:00:00
----------------------------------------------------------------
Summary: 9 checks, 0 failed, 0 warning(s)
No failing checks.
```
