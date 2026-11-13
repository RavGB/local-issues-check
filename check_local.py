#!/usr/bin/env python3
"""Check a local machine for common issues and print a short report.

RavGB
Contact: hello@ravgb.com

Runs fully on the local machine using only the Python standard library.
Nothing is sent anywhere. Results are printed to the screen and can
also be saved to a text file.

Usage:
    python3 check_local.py
    python3 check_local.py --output report.txt
    python3 check_local.py --path /some/folder

Exit code:
    0  no failing checks (warnings are allowed)
    1  at least one check failed
    2  the script itself could not run a check setup step
"""

import argparse
import os
import platform
import shutil
import socket
import sys
import tempfile
import time
from datetime import datetime, timezone

OK = "OK"
WARN = "WARN"
FAIL = "FAIL"
SKIP = "SKIP"

results = []


def record(status, name, detail):
    results.append((status, name, detail))
    print(f"[{status:4}] {name}: {detail}")


def check_system():
    name = "System"
    try:
        detail = (
            f"{platform.system()} {platform.release()} "
            f"({platform.machine()}), Python {platform.python_version()}, "
            f"host {socket.gethostname()}"
        )
        record(OK, name, detail)
    except Exception as exc:  # a failed read is itself a finding
        record(FAIL, name, f"could not read system information: {exc}")


def check_disk(path):
    name = "Disk space"
    try:
        usage = shutil.disk_usage(path)
        used_pct = round(usage.used / usage.total * 100, 1)
        free_gb = round(usage.free / (1024 ** 3), 2)
        detail = f"{path} is {used_pct}% used, {free_gb} GB free"
        if used_pct >= 95:
            record(FAIL, name, detail + " - almost full, free space now")
        elif used_pct >= 85:
            record(WARN, name, detail + " - getting full")
        else:
            record(OK, name, detail)
    except Exception as exc:
        record(FAIL, name, f"could not read disk usage for {path}: {exc}")


def check_memory():
    name = "Memory"
    meminfo = "/proc/meminfo"
    if not os.path.exists(meminfo):
        record(SKIP, name, "detailed memory figures are not available on this system")
        return
    try:
        values = {}
        with open(meminfo, "r", encoding="utf-8") as handle:
            for line in handle:
                parts = line.split()
                if len(parts) >= 2 and parts[0] in ("MemTotal:", "MemAvailable:"):
                    values[parts[0]] = int(parts[1])
        total = values.get("MemTotal:", 0)
        available = values.get("MemAvailable:", 0)
        if total <= 0:
            record(SKIP, name, "memory figures could not be read")
            return
        free_pct = round(available / total * 100, 1)
        detail = f"{free_pct}% of memory is available"
        if free_pct < 10:
            record(FAIL, name, detail + " - very low, close programs or add memory")
        elif free_pct < 20:
            record(WARN, name, detail + " - low")
        else:
            record(OK, name, detail)
    except Exception as exc:
        record(FAIL, name, f"could not read memory information: {exc}")


def check_load():
    name = "Load"
    try:
        load1, _load5, _load15 = os.getloadavg()
        cores = os.cpu_count() or 1
        detail = f"1 minute load average {load1:.2f} on {cores} core(s)"
        if load1 > cores * 2:
            record(FAIL, name, detail + " - machine is heavily overloaded")
        elif load1 > cores:
            record(WARN, name, detail + " - busy")
        else:
            record(OK, name, detail)
    except (AttributeError, OSError):
        record(SKIP, name, "load average is not available on this system")


def check_writable(folder, name):
    try:
        fd, test_path = tempfile.mkstemp(prefix=".write-test-", dir=folder)
        os.close(fd)
        os.remove(test_path)
        record(OK, name, f"{folder} is writable")
    except Exception as exc:
        record(FAIL, name, f"{folder} is not writable: {exc}")


def check_hostname_lookup():
    name = "Local name lookup"
    try:
        hostname = socket.gethostname()
        address = socket.gethostbyname(hostname)
        record(OK, name, f"{hostname} resolves to {address}")
    except Exception as exc:
        record(FAIL, name, f"this machine's own name does not resolve: {exc}")


def check_loopback():
    name = "Loopback"
    server = None
    try:
        server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server.bind(("127.0.0.1", 0))
        port = server.getsockname()[1]
        server.listen(1)
        client = socket.create_connection(("127.0.0.1", port), timeout=5)
        conn, _addr = server.accept()
        client.sendall(b"ping")
        data = conn.recv(4)
        client.close()
        conn.close()
        if data == b"ping":
            record(OK, name, "a local socket round trip works")
        else:
            record(FAIL, name, "a local socket round trip returned unexpected data")
    except Exception as exc:
        record(FAIL, name, f"local socket test failed: {exc}")
    finally:
        if server is not None:
            server.close()


def check_clock():
    name = "Clock"
    try:
        now_local = datetime.now().astimezone()
        now_utc = datetime.now(timezone.utc)
        offset = now_local.utcoffset()
        record(
            OK,
            name,
            f"local time {now_local.strftime('%Y-%m-%d %H:%M:%S %Z')}, "
            f"UTC {now_utc.strftime('%Y-%m-%d %H:%M:%S')}, offset {offset}",
        )
    except Exception as exc:
        record(FAIL, name, f"could not read the clock: {exc}")


def main():
    parser = argparse.ArgumentParser(
        description="Check a local machine for common issues."
    )
    parser.add_argument(
        "--output",
        help="also save the report to this text file",
        default=None,
    )
    parser.add_argument(
        "--path",
        help="folder to test for disk space (default: current folder)",
        default=".",
    )
    args = parser.parse_args()
    path = os.path.abspath(args.path)

    lines = []

    def emit(text=""):
        print(text)
        lines.append(text)

    emit("Local issues check")
    emit("RavGB - hello@ravgb.com")
    emit(f"Started: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    emit(f"Folder tested: {path}")
    emit("-" * 60)

    # The record() helper prints as it goes; collect its lines too by
    # re-printing a summary at the end for the saved report.
    check_system()
    check_disk(path)
    check_memory()
    check_load()
    check_writable(path, "Folder writable")
    check_writable(tempfile.gettempdir(), "Temp folder writable")
    check_hostname_lookup()
    check_loopback()
    check_clock()

    for status, check_name, detail in results:
        lines.append(f"[{status:4}] {check_name}: {detail}")

    failures = [r for r in results if r[0] == FAIL]
    warnings = [r for r in results if r[0] == WARN]

    emit("-" * 60)
    emit(
        f"Summary: {len(results)} checks, "
        f"{len(failures)} failed, {len(warnings)} warning(s)"
    )
    if failures:
        emit("Failing checks:")
        for _status, check_name, detail in failures:
            emit(f"  - {check_name}: {detail}")
    else:
        emit("No failing checks.")

    if args.output:
        try:
            with open(args.output, "w", encoding="utf-8") as handle:
                handle.write("\n".join(lines) + "\n")
            print(f"Report saved to {args.output}")
        except Exception as exc:
            print(f"Could not save the report: {exc}")
            return 2

    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
