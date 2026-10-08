"""Run a sustained PVMBG outage with native k6 and sampled recovery evidence."""
import argparse
from datetime import datetime, timezone
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("demo", ROOT / "scripts/demo/demo.py")
demo = importlib.util.module_from_spec(spec)
spec.loader.exec_module(demo)


def timestamp():
    return datetime.now(timezone.utc).isoformat()


def containers():
    ids = subprocess.check_output(
        ["docker", "compose", "--profile", "demo", "ps", "-q"], cwd=ROOT, text=True).split()
    raw = json.loads(subprocess.check_output(["docker", "inspect", *ids], text=True))
    return {c["Config"]["Labels"]["com.docker.compose.service"]: {
        "id": c["Id"], "restart_count": c["RestartCount"],
        "started_at": c["State"]["StartedAt"],
        "state": c["State"]["Status"], "health": c["State"].get("Health", {}).get("Status")
    } for c in raw}


def snapshot(session):
    result = {"at": timestamp()}
    code, health, _ = demo.request(8082, "/health")
    result["pvmbg_http"] = code
    result["simulation"] = health.get("simulation")
    for hazard_type in ("seismic", "volcanic"):
        code, body, _ = session.get(f"/v1/hazards/{hazard_type}?limit=1")
        result[hazard_type] = {"http": code, "data_count": len(body.get("data", [])),
                               "sources": body.get("sources", [])}
    return result


def source(row, hazard_type, name):
    return next((s for s in row[hazard_type]["sources"] if s["source"] == name), {})


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True)
    parser.add_argument("--seconds", type=int, default=600)
    parser.add_argument("--recovery-seconds", type=int, default=90)
    args = parser.parse_args()
    if args.seconds <= 0 or args.recovery_seconds <= 0:
        parser.error("durations must be positive")
    binary = os.environ.get("K6_BINARY")
    if not binary:
        parser.error("set K6_BINARY to the native k6 executable")
    output = (ROOT / args.output).resolve()
    output.mkdir(parents=True, exist_ok=False)
    env = os.environ.copy()
    env.update(demo.local_env())
    env.update(API_URL="http://127.0.0.1:8080", AUTH_URL="http://127.0.0.1:8090",
               PVMBG_URL="http://127.0.0.1:8082", CLIENTS_FILE=str(ROOT / "env/demo-clients.json"),
               OUTAGE_SECONDS=str(args.seconds), RECOVERY_SECONDS=str(args.recovery_seconds),
               OUTAGE_VUS="5", RECOVERY_VUS="5", SLEEP="0.2")
    session = demo.Session("field-team")
    before = containers()
    process = None
    rows = []
    try:
        demo.admin("outage", {"enabled": False})
        warmup_deadline = time.monotonic() + 90
        while True:
            baseline = snapshot(session)
            if all(baseline[t]["http"] == 200 and baseline[t]["data_count"] > 0
                   and source(baseline, t, name).get("status") == "HEALTHY"
                   for t, name in (("seismic", "BMKG"), ("volcanic", "PVMBG"))):
                break
            if time.monotonic() > warmup_deadline:
                raise RuntimeError("sources not healthy before outage")
            time.sleep(3)
        version = subprocess.check_output([binary, "version"], text=True).strip()
        config = {"started_at": timestamp(), "outage_seconds": args.seconds,
                  "recovery_seconds": args.recovery_seconds, "outage_vus": 5,
                  "recovery_vus": 5, "sample_interval_seconds": 5,
                  "k6": version, "baseline": baseline, "containers_before": before}
        (output / "configuration.json").write_text(json.dumps(config, indent=2)+"\n", encoding="utf-8")
        started = time.monotonic()
        with (output / "k6.txt").open("w", encoding="utf-8") as log, (output / "samples.jsonl").open("w", encoding="utf-8") as samples:
            process = subprocess.Popen([binary, "run", "--quiet", "--summary-export", str(output / "summary.json"),
                                        str(ROOT / "scripts/loadtest/outage.js")], cwd=ROOT, env=env,
                                       stdout=log, stderr=subprocess.STDOUT)
            while process.poll() is None:
                elapsed = time.monotonic()-started
                if elapsed > args.seconds + args.recovery_seconds + 120:
                    raise RuntimeError("outage test exceeded deadline")
                row = snapshot(session)
                row["elapsed_seconds"] = round(time.monotonic()-started, 3)
                rows.append(row)
                samples.write(json.dumps(row)+"\n")
                samples.flush()
                time.sleep(5)
        metrics = json.loads((output / "summary.json").read_text())["metrics"]
        enabled = [r for r in rows if r["simulation"]["outage"]]
        disabled = [r for r in rows if not r["simulation"]["outage"] and r["elapsed_seconds"] >= args.seconds]
        stale = [r for r in enabled if r["elapsed_seconds"] >= 30]
        after = containers()
        checks = {
            "k6_passed": process.returncode == 0,
            "nonnegative_timings": all(m.get("min", 0) >= 0 for name, m in metrics.items()
                                       if name.startswith(("http_req_", "business_latency"))),
            "outage_duration_observed": bool(enabled) and enabled[-1]["elapsed_seconds"]-enabled[0]["elapsed_seconds"] >= args.seconds-15,
            "seismic_healthy_during_outage": bool(enabled) and all(r["seismic"]["http"] == 200 and source(r, "seismic", "BMKG").get("status") == "HEALTHY" for r in enabled),
            "volcanic_stale_after_detection": bool(stale) and all(r["volcanic"]["http"] == 200 and r["volcanic"]["data_count"] > 0 and source(r, "volcanic", "PVMBG").get("stale_since") and source(r, "volcanic", "PVMBG").get("status") != "HEALTHY" for r in stale),
            "volcanic_recovers": any(r["volcanic"]["http"] == 200 and source(r, "volcanic", "PVMBG").get("status") == "HEALTHY" for r in disabled),
            "no_restarts_or_replacements": before == after,
        }
        result = {"finished_at": timestamp(), "checks": checks, "sample_count": len(rows),
                  "outage_samples": len(enabled), "containers_after": after}
        (output / "result.json").write_text(json.dumps(result, indent=2)+"\n", encoding="utf-8")
        print(json.dumps(result["checks"]), flush=True)
        if not all(checks.values()):
            raise RuntimeError("long outage verification failed; inspect preserved evidence")
    finally:
        if process is not None and process.poll() is None:
            process.terminate()
            process.wait(timeout=20)
        demo.admin("outage", {"enabled": False})


if __name__ == "__main__":
    main()
