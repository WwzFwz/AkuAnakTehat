"""Run pinned local k6 scenarios and record actual client-api TCP connections.

Requires Python 3 and Docker Compose. Credentials remain in ignored env files.
The three-second mock override is temporary; only PVMBG is recreated/restored.
"""
import argparse
import json
import math
import os
from pathlib import Path
import re
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / "docs/evidence/integration/load"
NATIVE_K6 = os.environ.get("K6_BINARY")
K6_OVERRIDE_VARS = (
    "VUS",
    "DURATION",
    "SLEEP",
    "OUTAGE_SECONDS",
    "RECOVERY_SECONDS",
    "OUTAGE_VUS",
    "RECOVERY_VUS",
    "K6_HTTP_TIMEOUT",
)
DEFAULT_DURATIONS = {"seismic-only": "60s", "sustained": "90s"}
DEFAULT_RUNNER_TIMEOUT_SECONDS = 240.0
RUNNER_TIMEOUT_MARGIN_SECONDS = 60.0
DURATION_PART = re.compile(r"(?P<value>\d+(?:\.\d+)?)(?P<unit>ms|s|m|h)")


def docker(*args, **kwargs):
    return subprocess.run(["docker", "compose", *args], cwd=ROOT, check=True, **kwargs)


def scenario_overrides(duration=None):
    overrides = {key: os.environ[key] for key in K6_OVERRIDE_VARS if key in os.environ}
    if duration is not None:
        overrides["DURATION"] = duration
    return overrides


def apply_overrides(environment, overrides):
    result = environment.copy()
    result.update(overrides)
    return result


def parse_duration(value):
    text = str(value).strip()
    if not text:
        raise RuntimeError("duration must not be empty")
    total = 0.0
    position = 0
    multipliers = {"ms": 0.001, "s": 1.0, "m": 60.0, "h": 3600.0}
    for match in DURATION_PART.finditer(text):
        if match.start() != position:
            raise RuntimeError(f"invalid duration: {value}")
        total += float(match.group("value")) * multipliers[match.group("unit")]
        position = match.end()
    if position != len(text) or not math.isfinite(total) or total <= 0:
        raise RuntimeError(f"invalid duration: {value}")
    return total


def seconds_environment(name, default):
    raw = os.environ.get(name, str(default))
    try:
        value = float(raw)
    except ValueError as error:
        raise RuntimeError(f"{name} must be a non-negative number of seconds") from error
    if not math.isfinite(value) or value < 0:
        raise RuntimeError(f"{name} must be a non-negative number of seconds")
    return value


def expected_duration(name, duration=None):
    if name == "outage":
        # The restore and recovery scenarios start shortly after the outage.
        return seconds_environment("OUTAGE_SECONDS", 20) + seconds_environment("RECOVERY_SECONDS", 30) + 2
    selected = duration if duration is not None else os.environ.get("DURATION")
    return parse_duration(selected or DEFAULT_DURATIONS[name])


def runner_timeout(name, duration=None):
    minimum = expected_duration(name, duration) + RUNNER_TIMEOUT_MARGIN_SECONDS
    configured = os.environ.get("K6_RUNNER_TIMEOUT")
    if configured:
        timeout = parse_duration(configured)
        if timeout < minimum:
            raise RuntimeError(
                f"K6_RUNNER_TIMEOUT for {name} must be at least {minimum:g}s "
                "(scenario duration plus 60s margin)"
            )
        return timeout
    return max(DEFAULT_RUNNER_TIMEOUT_SECONDS, minimum)


def docker_command(name, overrides):
    args = ["docker", "compose", "run", "--rm", "--no-deps", "--env-from-file", "./env/demo.env"]
    for key in sorted(overrides):
        args += ["-e", f"{key}={overrides[key]}"]
    args += ["--volume", f"{RESULTS.as_posix()}:/results", "loadtest", "run", "--quiet", "--summary-export",
             f"/results/{name}.json", f"/scripts/{name}.js"]
    return args


def native_environment(overrides):
    process_env = os.environ.copy()
    for line in (ROOT / "env/demo.env").read_text().splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            process_env[key] = value
    process_env.update(API_URL="http://127.0.0.1:8080", AUTH_URL="http://127.0.0.1:8090",
                       PVMBG_URL="http://127.0.0.1:8082", CLIENTS_FILE=str(ROOT / "env/demo-clients.json"))
    return apply_overrides(process_env, overrides)


def run(name, duration=None, connections=False):
    overrides = scenario_overrides(duration)
    deadline = runner_timeout(name, overrides.get("DURATION"))
    args = docker_command(name, overrides)
    process_env = None
    if NATIVE_K6:
        process_env = native_environment(overrides)
        args = [NATIVE_K6, "run", "--quiet", "--summary-export", str(RESULTS / f"{name}.json"),
                str(ROOT / "scripts/loadtest" / f"{name}.js")]
    samples = []
    started = time.monotonic()
    with (RESULTS / f"{name}.txt").open("w", encoding="utf-8") as output:
        process = subprocess.Popen(args, cwd=ROOT, env=process_env, stdout=output, stderr=subprocess.STDOUT)
        try:
            while process.poll() is None:
                if time.monotonic() - started > deadline:
                    raise RuntimeError(f"{name} exceeded runner deadline of {deadline:.0f}s")
                if connections:
                    raw = docker("exec", "-T", "client-api", "cat", "/proc/net/tcp", "/proc/net/tcp6",
                                 capture_output=True, text=True).stdout
                    count = sum(1 for line in raw.splitlines()
                                if len(line.split()) > 3 and line.split()[1].endswith(":1F90")
                                and line.split()[3] == "01")
                    samples.append({"elapsed_seconds": round(time.monotonic()-started, 2),
                                    "established_api_connections": count})
                time.sleep(1)
        finally:
            if process.poll() is None:
                process.terminate()
                process.wait(timeout=20)
    if samples:
        (RESULTS / "connections.json").write_text(json.dumps(samples, indent=2)+"\n", encoding="utf-8")
    if process.returncode:
        raise RuntimeError(f"{name} failed; inspect its preserved output")
    metrics = json.loads((RESULTS / f"{name}.json").read_text())["metrics"]
    invalid = [key for key, metric in metrics.items()
               if (key.startswith("http_req_") or key.startswith("business_latency"))
               and metric.get("min", 0) < 0]
    if invalid:
        raise RuntimeError(f"{name} recorded negative timings ({', '.join(invalid)}); results are invalid, check the load-generator clock")
    if connections:
        # Confirm at least fifty live TCP sessions for a contiguous minute,
        # rather than treating VU count as measured connection count.
        beginning = None
        longest = 0
        for sample in samples:
            if sample["established_api_connections"] >= 50:
                if beginning is None:
                    beginning = sample["elapsed_seconds"]
                longest = max(longest, sample["elapsed_seconds"]-beginning)
            else:
                beginning = None
        if longest < 60:
            raise RuntimeError(f"50 TCP sessions sustained only {longest:.2f}s; expected >=60s")
    print(f"PASS {name}", flush=True)


def main():
    global RESULTS
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", default="docs/evidence/integration/load",
                        help="Output directory relative to repo root; use a new directory to preserve previous evidence")
    parser.add_argument("--skip-connection-check", action="store_true",
                        help="Skip the 50-TCP sustained check; for short smoke tests only")
    args = parser.parse_args()
    # Reject invalid deadlines before writing evidence or changing the mock.
    for name in (*DEFAULT_DURATIONS, "outage"):
        runner_timeout(name)
    RESULTS = (ROOT / args.output).resolve()
    RESULTS.mkdir(parents=True, exist_ok=True)
    local = ROOT / ".local"
    local.mkdir(exist_ok=True)
    override = local / "loadtest-delay.yml"
    override.write_text("services:\n  pvmbg-mock:\n    environment:\n      PVMBG_DELAY_MIN: 3s\n      PVMBG_DELAY_MAX: 3s\n", encoding="utf-8")
    with (RESULTS / "environment.txt").open("w", encoding="utf-8") as output:
        if NATIVE_K6:
            subprocess.run([NATIVE_K6, "version"], check=True, stdout=output, stderr=subprocess.STDOUT)
        else:
            docker("run", "--rm", "--no-deps", "loadtest", "version", stdout=output, stderr=subprocess.STDOUT)
        subprocess.run(["docker", "info", "--format", "Docker {{.ServerVersion}}; CPUs={{.NCPU}}; MemoryBytes={{.MemTotal}}; OS={{.OperatingSystem}}"], cwd=ROOT, check=True, stdout=output)
    try:
        docker("-f", "docker-compose.yml", "-f", str(override), "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pvmbg-mock")
        time.sleep(10)
        run("seismic-only")
        run("sustained", connections=not args.skip_connection_check)
        run("outage")
    finally:
        docker("up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pvmbg-mock")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, subprocess.SubprocessError) as error:
        print(str(error), file=sys.stderr)
        sys.exit(1)
