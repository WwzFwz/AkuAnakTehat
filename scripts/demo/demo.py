"""Local demo controls and existing P1-P5 checks. Never print credentials/tokens."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[2]
PORTS = {"client-api": 8080, "bmkg-mock": 8081, "pvmbg-mock": 8082,
         "auth-service": 8090, "dashboard-updater": 8091, "notifier": 8092,
         "pemda-portal": 8093}
SUITES = {
    "p1": "TestIngestPipeline|TestDynamicSchemaAndStaleAPI",
    "p2": "TestDynamicSchemaAndStaleAPI",
    "p3": "TestMockContracts|TestTokenRotationAndAuthorization|TestJWTClaimValidation|TestQueryIntegration|TestNaturalExpiryAndFieldCLI",
    "p4": "TestDynamicSchemaAndStaleAPI|TestIndependentRebuild",
    "p5": "TestEventPipeline",
    "all": ".",
}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def request(port, path, data=None, headers=None, form=False):
    """HTTP only to fixed loopback ports; HTTP errors remain inspectable statuses."""
    headers = dict(headers or {})
    headers.setdefault("X-Correlation-ID", "demo-" + uuid.uuid4().hex)
    payload = None
    if data is not None:
        payload = (urllib.parse.urlencode(data) if form else json.dumps(data)).encode()
        headers["Content-Type"] = "application/x-www-form-urlencoded" if form else "application/json"
    req = urllib.request.Request(f"http://127.0.0.1:{port}{path}", data=payload, headers=headers)
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())
    try:
        response = opener.open(req, timeout=6)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        raw = response.read(4 * 1024 * 1024 + 1)
        if len(raw) > 4 * 1024 * 1024:
            raise RuntimeError("Response exceeds demo display limit")
        try:
            body = json.loads(raw)
        except json.JSONDecodeError:
            body = {"error": "non-JSON body withheld"}
        return response.code, body, response.headers.get("X-Correlation-ID")


def local_env():
    values = {}
    for line in (ROOT / "env/demo.env").read_text(encoding="utf-8-sig").splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            values[key] = value
    return values


def clients():
    return json.loads((ROOT / "env/demo-clients.json").read_text(encoding="utf-8-sig"))


def emit(value):
    # Additional defense for inspected HTTP bodies containing arbitrary attributes.
    text = json.dumps(value, indent=2, ensure_ascii=True)
    secrets = list(local_env().values()) + [c["client_secret"] for c in clients()]
    for secret in secrets:
        if len(secret) >= 16:
            text = text.replace(secret, "[REDACTED]")
    print(text, flush=True)


class Session:
    def __init__(self, identity):
        self.identity, self.access, self.refresh, self.expires = identity, "", "", 0

    def renew(self):
        if self.refresh:
            grant = {"grant_type": "refresh_token", "refresh_token": self.refresh}
        else:
            client = next(c for c in clients() if c["client_id"] == self.identity)
            grant = {"grant_type": "client_credentials", "client_id": self.identity,
                     "client_secret": client["client_secret"]}
        code, body, _ = request(8090, "/oauth/token", grant, form=True)
        if code != 200:
            raise RuntimeError(f"Authentication returned HTTP {code}; response withheld")
        self.access, self.refresh = body["access_token"], body["refresh_token"]
        self.expires = time.monotonic() + body["expires_in"] - 2

    def get(self, path):
        if time.monotonic() >= self.expires:
            self.renew()
        result = request(8080, path, headers={"Authorization": "Bearer " + self.access})
        if result[0] == 401:
            self.renew()
            result = request(8080, path, headers={"Authorization": "Bearer " + self.access})
        return result


def admin(path, body):
    code, result, _ = request(8082, "/admin/" + path, body,
                              {"X-Admin-Key": local_env()["PVMBG_ADMIN_KEY"]})
    if code != 200:
        raise RuntimeError(f"PVMBG admin returned HTTP {code}")
    return result


def compose(*args):
    subprocess.run(["docker", "compose", *args], cwd=ROOT, check=True)


def verify(suite):
    print(f"Running {suite}: {SUITES[suite]}. State-changing suites must run sequentially.", flush=True)
    env = os.environ.copy()
    env.update(GOWORK="off", GOTOOLCHAIN="local", GOTELEMETRY="off",
               GOCACHE=str(ROOT / ".local/go-cache"), GOMODCACHE=str(ROOT / ".local/go-mod-cache"))
    args = ["go", "test", *[f"./scripts/check/{s}_test.go" for s in ("foundation", "ingest", "events", "query")],
            "-run", f"^({SUITES[suite]})$" if suite != "all" else ".", "-v", "-count=1", "-timeout=15m"]
    return subprocess.run(args, cwd=ROOT, env=env).returncode


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("status", help="Show health/readiness for all services including Pemda")
    read = commands.add_parser("read", help="Read the client API without displaying a token")
    read.add_argument("--identity", choices=["media", "field-team", "bnpb-ops"], default="media")
    read.add_argument("--type", choices=["seismic", "volcanic", "all"], default="all")
    read.add_argument("--raw", action="store_true")
    read.add_argument("--limit", type=int, choices=range(1, 501), default=2, metavar="1..500")
    read.add_argument("--expect", type=int, choices=[200, 400, 401, 403, 429, 503], default=200)
    read.add_argument("--cursor", help="Opaque cursor from a previous response")
    read.add_argument("--id", type=uuid.UUID, help="Inspect one hazard, including an older schema record")
    commands.add_parser("schema").add_argument("version", type=int, choices=[1, 2])
    outage = commands.add_parser("outage")
    outage.add_argument("state", choices=["on", "off"])
    outage.add_argument("--mode", choices=["error", "hang"], default="error")
    commands.add_parser("restore", help="Restore normal demo baseline, preserving all volumes")
    commands.add_parser("verify").add_argument("suite", choices=SUITES)
    view = commands.add_parser("view", help="Inspect a consumer's local state")
    view.add_argument("service", choices=["dashboard-updater", "notifier", "pemda-portal"])
    view.add_argument("--after", default="")
    args = parser.parse_args()
    if args.command == "verify":
        return verify(args.suite)
    if args.command == "status":
        rows = []
        for service, port in PORTS.items():
            row = {"service": service}
            for endpoint in ("health", "ready"):
                try:
                    code, body, _ = request(port, "/" + endpoint)
                    row[endpoint] = code
                    if service == "pvmbg-mock" and "simulation" in body:
                        row["simulation"] = body["simulation"]
                except (OSError, urllib.error.URLError):
                    row[endpoint] = "unreachable"
            rows.append(row)
        emit(rows)
        return 0 if all(r["health"] == r["ready"] == 200 for r in rows) else 1
    if args.command == "read":
        path = "/v1/hazards" + ("/" + args.type if args.type != "all" else "")
        params = {"limit": args.limit}
        if args.raw:
            params["include"] = "raw"
        if args.cursor:
            params["cursor"] = args.cursor
        if args.id:
            path = "/v1/hazards/" + str(args.id) + ("/raw" if args.raw else "")
            params = {}
        code, body, trace = Session(args.identity).get(path + "?" + urllib.parse.urlencode(params))
        emit({"identity": args.identity, "http_status": code, "correlation_id": trace, "body": body})
        return 0 if code == args.expect else 1
    if args.command == "schema":
        emit(admin("schema-version", {"version": args.version}))
    elif args.command == "outage":
        emit(admin("outage", {"enabled": args.state == "on", "mode": args.mode}))
    elif args.command == "view":
        path = "/processed" if args.service == "notifier" else "/view"
        code, body, _ = request(PORTS[args.service], path + "?" + urllib.parse.urlencode({"limit": 2, "after": args.after}))
        emit({"service": args.service, "http_status": code, "body": body})
        return 0 if code == 200 else 1
    elif args.command == "restore":
        compose("--profile", "demo", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "180",
                "canonical-db", "auth-store", "kafka", "bmkg-mock", "pvmbg-mock", "aggregator",
                "auth-service", "client-api", "dashboard-updater", "notifier", "pemda-portal")
        admin("outage", {"enabled": False, "mode": "error"})
        emit(admin("schema-version", {"version": 1}))
        print("Baseline restored. Allow polling/breaker recovery before checking source status.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print("Interrupted. Run: py scripts/demo/demo.py restore", file=sys.stderr)
        raise SystemExit(130)
    except (OSError, ValueError, KeyError, StopIteration, RuntimeError, subprocess.SubprocessError, urllib.error.URLError) as error:
        # Never interpolate arbitrary response data or local secret values into errors.
        print(f"Demo failed ({type(error).__name__}). Check stack/bootstrap and run status. "
              "After an interrupted scenario, run restore.", file=sys.stderr)
        raise SystemExit(1)
