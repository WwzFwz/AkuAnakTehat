"""Print structured Compose logs with an exact correlation_id match."""
import argparse
import json
from pathlib import Path
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("correlation_id")
    parser.add_argument("--tail", type=int, default=2000, help="lines per container (default: 2000)")
    args = parser.parse_args()
    if not args.correlation_id or len(args.correlation_id) > 256 or args.tail < 1:
        parser.error("supply a correlation ID of 1-256 characters and a positive tail")
    result = subprocess.run(
        ["docker", "compose", "--profile", "demo", "logs", "--no-color", "--no-log-prefix", f"--tail={args.tail}"],
        cwd=Path(__file__).resolve().parents[2], capture_output=True, text=True,
        encoding="utf-8", errors="replace", check=True,
    )
    matches = []
    for line in result.stdout.splitlines():
        try:
            record = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(record, dict) and record.get("correlation_id") == args.correlation_id:
            matches.append(record)
    for record in sorted(matches, key=lambda record: record.get("time", "")):
        print(json.dumps(record, ensure_ascii=True))
    return 0 if matches else 1


if __name__ == "__main__":
    raise SystemExit(main())
