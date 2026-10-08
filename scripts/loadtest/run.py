"""Run pinned local k6 scenarios and record actual client-api TCP connections.

Requires Python 3 and Docker Compose. Credentials remain in ignored env files.
The three-second mock override is temporary; only PVMBG is recreated/restored.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / "docs/evidence/integration/load"
NATIVE_K6 = os.environ.get("K6_BINARY")


def docker(*args, **kwargs):
    return subprocess.run(["docker", "compose", *args], cwd=ROOT, check=True, **kwargs)


def run(name, duration=None, connections=False):
    args = ["docker", "compose", "run", "--rm", "--no-deps", "--env-from-file", "./env/demo.env",
            "--volume", f"{RESULTS.as_posix()}:/results"]
    if duration:
        args += ["-e", f"DURATION={duration}"]
    args += ["loadtest", "run", "--quiet", "--summary-export",
             f"/results/{name}.json", f"/scripts/{name}.js"]
    process_env = None
    if NATIVE_K6:
        process_env = os.environ.copy()
        for line in (ROOT / "env/demo.env").read_text().splitlines():
            if line and not line.startswith("#") and "=" in line:
                key, value = line.split("=", 1)
                process_env[key] = value
        process_env.update(API_URL="http://127.0.0.1:8080", AUTH_URL="http://127.0.0.1:8090",
                           PVMBG_URL="http://127.0.0.1:8082", CLIENTS_FILE=str(ROOT / "env/demo-clients.json"))
        if duration:
            process_env["DURATION"] = duration
        args = [NATIVE_K6, "run", "--quiet", "--summary-export", str(RESULTS / f"{name}.json"),
                str(ROOT / "scripts/loadtest" / f"{name}.js")]
    samples = []
    started = time.monotonic()
    with (RESULTS / f"{name}.txt").open("w", encoding="utf-8") as output:
        process = subprocess.Popen(args, cwd=ROOT, env=process_env, stdout=output, stderr=subprocess.STDOUT)
        try:
            while process.poll() is None:
                if time.monotonic() - started > 240:
                    raise RuntimeError(f"{name} exceeded runner deadline")
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
    args = parser.parse_args()
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
        run("sustained", connections=True)
        run("outage")
    finally:
        docker("up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pvmbg-mock")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, subprocess.SubprocessError) as error:
        print(str(error), file=sys.stderr)
        sys.exit(1)
