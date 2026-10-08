"""Verify a fresh checkout, fresh secrets and isolated Compose volumes; restore the primary stack."""
import argparse
from datetime import datetime, timezone
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import time
import uuid
import zipfile

ROOT = Path(__file__).resolve().parents[2]


def capture(args, cwd=ROOT, env=None):
    return subprocess.check_output(args, cwd=cwd, env=env, text=True).strip()


def volumes(project):
    names = capture(["docker", "volume", "ls", "--filter", f"label=com.docker.compose.project={project}", "--format", "{{.Name}}"]).split()
    if not names:
        return {}
    data = json.loads(capture(["docker", "volume", "inspect", *names]))
    return {v["Name"]: v["CreatedAt"] for v in data}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    output = (ROOT / args.output).resolve()
    output.mkdir(parents=True, exist_ok=False)
    run_id = uuid.uuid4().hex[:8]
    project = "bnpb-bootstrap-" + run_id
    checkout = (ROOT / ".local" / project).resolve()
    assert checkout.is_relative_to(ROOT / ".local") and not checkout.exists()
    revision = capture(["git", "rev-parse", "HEAD"])
    archive = subprocess.check_output(["git", "archive", "--format=zip", revision], cwd=ROOT)
    with zipfile.ZipFile(io.BytesIO(archive)) as bundle:
        assert all((checkout / entry).resolve().is_relative_to(checkout) for entry in bundle.namelist())
        bundle.extractall(checkout)
    assert not list((checkout / "env").glob("*.env"))
    assert not list((checkout / "env/keys").glob("*.pem"))
    original = volumes("bnpb-m1")
    assert original and not volumes(project)
    primary_spec = importlib.util.spec_from_file_location("primary_demo", ROOT / "scripts/demo/demo.py")
    primary_demo = importlib.util.module_from_spec(primary_spec)
    primary_spec.loader.exec_module(primary_demo)
    primary_session = primary_demo.Session("field-team")
    status, original_page, _ = primary_session.get("/v1/hazards?limit=1")
    assert status == 200 and original_page["data"]
    original_hazard = original_page["data"][0]["hazard_id"]
    primary_secrets = {p: hashlib.sha256(p.read_bytes()).digest() for p in (ROOT / "env").rglob("*")
                       if p.is_file() and p.suffix in (".env", ".json", ".pem")}
    env = os.environ.copy()
    env.update(GOWORK="off", GOTOOLCHAIN="local", GOTELEMETRY="off", COMPOSE_PROJECT_NAME=project)
    compose = ["docker", "compose", "-p", project, "-f", str(checkout / "docker-compose.yml"), "--profile", "demo"]
    primary = ["docker", "compose", "-p", "bnpb-m1", "--profile", "demo"]
    result = {"started_at": datetime.now(timezone.utc).isoformat(), "revision": revision,
              "project": project, "checkout": str(checkout.relative_to(ROOT)),
              "original_volumes_before": original, "isolated_volumes_before": {}, "bootstrap_passed": False}
    stopped = False
    try:
        with (output / "bootstrap.txt").open("w", encoding="utf-8") as log:
            cmd = ["go", "run", "./scripts/secrets/generate.go", "-root", str(checkout)]
            subprocess.run(cmd, cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=180)
            secret_files = sorted(p for p in (checkout / "env").rglob("*") if p.is_file() and p.suffix in (".env", ".json", ".pem"))
            hashes = {p: hashlib.sha256(p.read_bytes()).digest() for p in secret_files}
            assert len(secret_files) == 12
            assert all(p.read_bytes() != (ROOT / p.relative_to(checkout)).read_bytes() for p in secret_files if p.suffix == ".pem")
            subprocess.run(cmd, cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=180)
            assert all(hashlib.sha256(p.read_bytes()).digest() == h for p, h in hashes.items())
            result["fresh_secret_files"] = len(secret_files)
            result["bootstrap_idempotent"] = True
        config = json.loads(capture([*compose, "config", "--format", "json"], cwd=checkout, env=env))
        assert config["name"] == project
        assert all(not v.get("external") and v["name"].startswith(project+"_") for v in config["volumes"].values())
        assert all(not n.get("external") and n["name"].startswith(project+"_") for n in config["networks"].values())
        result["isolated_volume_names"] = [v["name"] for v in config["volumes"].values()]
        result["isolated_network_names"] = [v["name"] for v in config["networks"].values()]
        with (output / "primary-stop.txt").open("w", encoding="utf-8") as log:
            stopped = True
            subprocess.run([*primary, "stop"], cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=120)
        print("Primary stack stopped; original volumes retained. Starting isolated stack.", flush=True)
        with (output / "build.txt").open("w", encoding="utf-8") as log:
            subprocess.run([*compose, "up", "-d", "--build", "--wait", "--wait-timeout", "240"],
                           cwd=checkout, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=600)
        result["isolated_volumes_created"] = volumes(project)
        assert len(result["isolated_volumes_created"]) == 6
        migration = capture([*compose, "exec", "-T", "canonical-db", "sh", "-ec",
                             'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT version, dirty FROM schema_migrations"'],
                            cwd=checkout, env=env)
        version, dirty = migration.split("|")
        assert int(version) > 0 and dirty == "f"
        result["migration_version"] = int(version)
        result["migration_dirty"] = False
        spec = importlib.util.spec_from_file_location("fresh_demo", checkout / "scripts/demo/demo.py")
        demo = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(demo)
        session = demo.Session("field-team")
        deadline = time.monotonic()+120
        while True:
            code, body, _ = session.get("/v1/hazards?limit=100&include=raw")
            sources = body.get("sources", [])
            if code == 200 and len(body.get("data", [])) >= 40 and {s["source"] for s in sources} == {"BMKG", "PVMBG"} and all(s["status"] == "HEALTHY" for s in sources):
                break
            if time.monotonic() > deadline:
                raise RuntimeError("fresh source ingest did not become healthy")
            time.sleep(3)
        hazards = {h["hazard_id"] for h in body["data"]}
        result["canonical_count"] = len(hazards)
        result["sources"] = body["sources"]
        result["service_health"] = {}
        for service, port in demo.PORTS.items():
            for path in ("/health", "/ready"):
                # Source mocks expose only /health.
                if path == "/ready" and service in ("bmkg-mock", "pvmbg-mock"):
                    continue
                status, _, _ = demo.request(port, path)
                assert status == 200, (service, path, status)
                result["service_health"][service+path] = status
        media = demo.Session("media")
        status, _, _ = media.get("/v1/hazards?include=raw")
        assert status == 403
        result["media_raw_status"] = status
        result["consumer_counts"] = {}
        for name, port, path in (("dashboard", 8091, "/view"), ("notifier", 8092, "/processed"), ("pemda", 8093, "/view")):
            deadline = time.monotonic()+90
            while True:
                status, view, _ = demo.request(port, path+"?limit=200")
                if status not in (200, 503):
                    raise RuntimeError(f"fresh {name} inspection returned HTTP {status}")
                records = view.get("data", [])
                ids = {record["hazard_id"] for record in records}
                if status == 200 and hazards <= ids:
                    result["consumer_counts"][name] = len(records)
                    break
                if time.monotonic() > deadline:
                    raise RuntimeError(f"fresh {name} did not receive initial hazards")
                time.sleep(3)
        result["bootstrap_passed"] = True
        print("Fresh secrets, migrations, healthy API, authorization and all three consumers passed.", flush=True)
    finally:
        try:
            # Only the uniquely named test project may have its volumes removed.
            isolated = volumes(project)
            assert project.startswith("bnpb-bootstrap-") and all(name.startswith(project+"_") for name in isolated)
            with (output / "cleanup.txt").open("w", encoding="utf-8") as log:
                subprocess.run([*compose, "down", "--volumes", "--remove-orphans"], cwd=checkout, env=env,
                               stdout=log, stderr=subprocess.STDOUT, check=True, timeout=180)
            result["isolated_volumes_after_cleanup"] = volumes(project)
        finally:
            if stopped:
                with (output / "primary-restore.txt").open("w", encoding="utf-8") as log:
                    subprocess.run([*primary, "up", "-d", "--wait", "--wait-timeout", "240"],
                                   cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=300)
            result["original_volumes_after"] = volumes("bnpb-m1")
            result["original_volumes_preserved"] = result["original_volumes_after"] == original
            result["original_secret_files_preserved"] = all(hashlib.sha256(p.read_bytes()).digest() == h for p, h in primary_secrets.items())
            deadline = time.monotonic()+90
            while True:
                status, original_detail, _ = primary_session.get("/v1/hazards/"+original_hazard)
                if status == 200 and original_detail.get("hazard_id") == original_hazard:
                    result["original_hazard_preserved"] = True
                    break
                if time.monotonic() > deadline:
                    result["original_hazard_preserved"] = False
                    break
                time.sleep(3)
            result["finished_at"] = datetime.now(timezone.utc).isoformat()
            (output / "result.json").write_text(json.dumps(result, indent=2)+"\n", encoding="utf-8")
    assert result["original_volumes_preserved"] and result["original_secret_files_preserved"] and result["original_hazard_preserved"] and not result["isolated_volumes_after_cleanup"]
    print("PASS clean bootstrap; primary stack restored and original volumes preserved.", flush=True)


if __name__ == "__main__":
    main()
