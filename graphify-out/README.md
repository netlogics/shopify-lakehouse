# Codebase Knowledge Graph

This directory holds a machine-generated knowledge graph of the repository,
produced by [graphify](https://pypi.org/project/graphifyy/). It maps the code,
compose files, and agent/docs config into nodes (functions, services, concepts)
and edges (calls, references, semantic relationships), then detects communities
and surfaces the most-connected "god nodes."

Snapshot: **119 nodes · 184 edges · 19 communities** across 41 files (~10,120
words). Extraction was 83% structural (AST) and 17% model-inferred.

## Artifacts

| File | What it is |
|---|---|
| [`graph.html`](graph.html) | Interactive visualization — **open in a browser**; no server needed |
| [`GRAPH_REPORT.md`](GRAPH_REPORT.md) | Plain-language report: god-nodes, communities, surprising connections, suggested questions |
| `graph.json` | The graph itself (nodes/edges/communities) in GraphRAG-ready JSON — feed to agents or graph tooling |
| `manifest.json` | Per-file AST + semantic hashes; drives incremental `--update` re-extraction |
| `cost.json` | Per-run file count and token usage |

Not committed (see `.gitignore`): `cache/` (regenerable) and `.graphify_*`
(machine-specific — they store absolute local paths from the run host).

## Viewing

```bash
# macOS
open graphify-out/graph.html
# Linux
xdg-open graphify-out/graph.html
```

`GRAPH_REPORT.md` is the fastest way to orient — it lists the core abstractions
and the cross-community bridges worth understanding first.

## Highlights (from the report)

- **God nodes** (most connected): `Load()`, `Registry`, `Config`,
  `NewGenerator()`, `Producer`, `Nessie Iceberg Catalog Service`, `main()`.
- **Hyperedges** (group relationships graphify inferred):
  - *Lakehouse Storage Layer* — MinIO + Nessie + Iceberg
  - *Kafka Ingestion Pipeline* — Generator → Kafka → Schema Registry → Flink
  - *Beads Task Tracking System* — Skill + Config + Orchestrator
- **Knowledge gaps**: 12 weakly-connected nodes (e.g. the init shell scripts)
  with ≤1 edge — candidates for missing links or undocumented wiring.

## Regenerating

The graph is a point-in-time snapshot; it does not auto-update as code changes.
To refresh it after edits:

```bash
# Full rebuild (Claude Code: the /graphify skill; or the CLI directly)
graphify .

# Incremental — only re-extract new/changed files (uses manifest.json)
graphify . --update
```

Install the CLI with `pipx install graphifyy` (or `uv tool install graphifyy`,
or `mise use -g pipx:graphifyy`).

## Querying

Ask questions against the built graph without rebuilding it:

```bash
graphify query "How does an inventory event get from the generator to Iceberg?"
graphify path "Producer" "Nessie Iceberg Catalog Service"
graphify explain "NewGenerator()"
```
