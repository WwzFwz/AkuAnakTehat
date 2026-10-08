"""Regenerate report figures/excerpts from the pinned repository revision.

Requires Python 3, matplotlib, PyMuPDF, Java, and a local PlantUML jar.
Does not run application tests or read generated credentials.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

HERE = Path(__file__).resolve().parents[1]
ROOT = HERE.parents[1]
REV = re.search(r"\\newcommand\{\\CodeRevision\}\{([^}]+)\}", (HERE / "metadata.tex").read_text()).group(1)
FINAL_REV = re.search(r"\\newcommand\{\\FinalEvidenceRevision\}\{([^}]+)\}", (HERE / "metadata.tex").read_text()).group(1)
FINAL_PATH = "docs/evidence/final-2026-10-08"
ASSETS = HERE / "assets"
manifest = {"revision": REV, "files": []}


def original(path, revision=REV):
    raw = subprocess.check_output(["git", "show", f"{revision}:{path}"], cwd=ROOT)
    manifest["files"].append({"path": path, "revision": revision, "sha256": hashlib.sha256(raw).hexdigest()})
    return raw.decode("utf-8-sig")


def tex(text):
    return str(text).replace("\\", r"\textbackslash{}").replace("_", r"\_").replace("%", r"\%").replace("&", r"\&").replace("#", r"\#")


def main():
    import fitz
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt

    for directory in ["figures", "code", "evidence"]:
        (ASSETS / directory).mkdir(parents=True, exist_ok=True)
    excerpts = [
        ("services/aggregator/internal/application/canonicalize/mapper.go", "func MapVolcanic", "p1-map-volcanic.txt"),
        ("services/client-api/internal/projection/projector.go", "func Project", "p3-projector.txt"),
        ("services/notifier/internal/application/apply.go", "// Send precedes", "p5-notifier.txt"),
    ]
    for path, marker, target in excerpts:
        text = original(path)
        start = text.index(marker)
        (ASSETS / "code" / target).write_text(text[start:], encoding="utf-8")
        manifest["files"][-1].update(excerpt=target, first_line=text[:start].count("\n")+1)

    log = original(f"{FINAL_PATH}/regression.txt", FINAL_REV)
    selected = [line for line in log.splitlines() if line.startswith("--- PASS:") or line == "PASS" or line.startswith("ok ")]
    (ASSETS / "evidence/regression-excerpt.txt").write_text("\n".join(selected)+"\n", encoding="utf-8")

    datasets = {name: json.loads(original(f"{FINAL_PATH}/load/{name}.json", FINAL_REV))["metrics"]
                for name in ["seismic-only", "sustained", "outage"]}
    rows = [r"\begin{table}[H]\centering\small",
            r"\begin{tabular}{lrrrrr}\toprule",
            r"Skenario & Bisnis/s & Sukses/s & p50 (ms) & p95 (ms) & p99 (ms)\\\midrule"]
    for name, metrics in datasets.items():
        latency = metrics["business_latency"]
        values = [metrics["business_requests"]["rate"], metrics["successful_requests"]["rate"], latency["med"], latency["p(95)"], latency["p(99)"]]
        rows.append(name + " & " + " & ".join(f"{v:.2f}".replace(".", ",") for v in values) + r"\\")
    rows += [r"\bottomrule\end{tabular}", r"\caption{Pengujian ulang k6 pada 8 Oktober 2026. Latency hanya HTTP 200; bisnis/s mencakup 429.}\end{table}"]
    (ASSETS / "evidence/load-table.tex").write_text("\n".join(rows)+"\n", encoding="utf-8")

    connection = json.loads(original(f"{FINAL_PATH}/load/connections.json", FINAL_REV))
    plt.rcParams.update({"font.family": "DejaVu Sans", "font.size": 10, "axes.spines.top": False,
                         "axes.spines.right": False, "axes.titleweight": "bold", "axes.labelcolor": "#17324d"})
    fig, axes = plt.subplots(2, 1, figsize=(8, 6.3), layout="constrained")
    labels = ["Seismic + volcanic", "Sustained", "Outage/recovery"]
    for j, (metric, color) in enumerate([("med", "#17324d"), ("p(95)", "#087f8c"), ("p(99)", "#ce9250")]):
        values = [d["business_latency"][metric] for d in datasets.values()]
        axes[0].bar([i+(j-1)*0.24 for i in range(3)], values, width=0.22, color=color,
                    label={"med":"p50", "p(95)":"p95", "p(99)":"p99"}[metric])
    axes[0].set_xticks(range(3), labels)
    axes[0].set_ylabel("Latency sukses (ms)")
    axes[0].set_title("A. Distribusi respons HTTP 200")
    axes[0].legend(frameon=False, ncols=3)
    axes[0].grid(axis="y", alpha=0.15)
    axes[1].step([r["elapsed_seconds"] for r in connection], [r["established_api_connections"] for r in connection],
                 where="post", color="#087f8c", linewidth=2)
    axes[1].axhline(50, color="#ce9250", linestyle="--", label="Ambang 50 koneksi")
    axes[1].set(xlabel="Waktu sejak runner mulai (s)", ylabel="TCP ESTABLISHED", ylim=(0, 56), title="B. Koneksi client-api selama sustained")
    axes[1].legend(frameon=False, loc="lower right")
    axes[1].grid(alpha=0.15)
    fig.savefig(ASSETS / "figures/06-load-results.pdf")
    fig.savefig(ASSETS / "figures/06-load-results.png", dpi=160)
    plt.close(fig)

    records = [json.loads(line) for line in original("docs/evidence/integration/event-trace.jsonl").splitlines()]
    corr = records[0]["correlation_id"]
    rows = [r"Correlation ID: \nolinkurl{"+corr+r"}.\par\smallskip", r"\begin{tabularx}{\linewidth}{lYr}\toprule",
            r"Service & Pesan log & Latency (ms)\\\midrule"]
    for record in records:
        latency = record.get("publish_latency_ms", record.get("latency_ms", "--"))
        rows.append(tex(record["service"])+" & "+tex(record["msg"])+" & "+tex(latency)+r"\\")
    rows.append(r"\bottomrule\end{tabularx}")
    (ASSETS / "evidence/trace-table.tex").write_text("\n".join(rows)+"\n", encoding="utf-8")

    jar = Path(os.environ.get("PLANTUML_JAR", ROOT / ".local/report-tools/plantuml.jar")).resolve()
    if not jar.is_file():
        raise SystemExit("Set PLANTUML_JAR to an installed jar or run scripts/bootstrap-tools.ps1")
    sources = sorted((HERE / "diagrams").glob("[0-9]*.puml"))
    subprocess.run(["java", "-jar", str(jar), "-tsvg", "-failfast2", *[str(p) for p in sources]], check=True, cwd=HERE)
    for source in sources:
        svg = source.with_suffix(".svg")
        shutil.copyfile(svg, ASSETS / "figures" / svg.name)
        with fitz.open(svg) as doc:
            (ASSETS / "figures" / source.with_suffix(".pdf").name).write_bytes(doc.convert_to_pdf())
        svg.unlink()
    (ASSETS / "provenance.json").write_text(json.dumps(manifest, indent=2)+"\n", encoding="utf-8")
    print(f"Generated {len(sources)} diagrams, load chart, source excerpts and evidence tables at revision", REV[:7])


if __name__ == "__main__":
    main()
