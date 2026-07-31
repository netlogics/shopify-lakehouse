# Graph Report - .  (2026-07-31)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 142 nodes · 212 edges · 24 communities (11 shown, 13 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 33 edges (avg confidence: 0.84)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8f629c84`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Nessie Iceberg Catalog Service
- Registry
- Load
- Producer
- OrderDetailsConsumer
- .NewOrderDetail
- Agent Instructions (AGENTS.md)
- main
- compact.py
- KarafkaApp
- post-checkout
- post-merge
- pre-commit
- pre-push
- prepare-commit-msg
- bootstrap.sh
- init-topics.sh
- init-bucket.sh
- entrypoint.sh
- Beads OpenAI Agent Interface
- generator
- shopify-lakehouse Project README

## God Nodes (most connected - your core abstractions)
1. `Load()` - 11 edges
2. `Config` - 10 edges
3. `Registry` - 10 edges
4. `Producer` - 10 edges
5. `NewGenerator()` - 9 edges
6. `Nessie Iceberg Catalog Service` - 8 edges
7. `MinIO Object Storage Service` - 7 edges
8. `Shopify Lakehouse Docker Compose Stack` - 7 edges
9. `main()` - 7 edges
10. `NewRegistry()` - 7 edges

## Surprising Connections (you probably didn't know these)
- `Beads Skill - Task Tracking Workflow` --semantically_similar_to--> `Agent Instructions (AGENTS.md)`  [INFERRED] [semantically similar]
  .agents/skills/beads/SKILL.md → AGENTS.md
- `Beads Orchestrator Config` --references--> `Beads Skill - Task Tracking Workflow`  [INFERRED]
  .beads/orchestrator.yml → .agents/skills/beads/SKILL.md
- `Beads Skill - Task Tracking Workflow` --conceptually_related_to--> `Beads Issue Tracker - AI-Native Issue Tracking`  [INFERRED]
  .agents/skills/beads/SKILL.md → .beads/README.md
- `Agent Instructions (AGENTS.md)` --references--> `Project Instructions for AI Agents (CLAUDE.md)`  [INFERRED]
  AGENTS.md → CLAUDE.md
- `Lakehouse Data Flow: Generator -> Kafka -> Flink -> Iceberg -> Dremio` --conceptually_related_to--> `Dremio Query Engine Service`  [INFERRED]
  README.md → compose/dremio.yml

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Lakehouse Storage Layer: MinIO + Nessie + Iceberg** — compose_minio_minio_service, compose_nessie_nessie_service, concept_iceberg_table_format [INFERRED 0.95]
- **Kafka Ingestion Pipeline: Generator -> Kafka -> Schema Registry -> Flink** — compose_generator_generator_service, compose_kafka_kafka_service, compose_kafka_schema_registry, compose_flink_flink_sql_submit [INFERRED 0.95]
- **Beads Task Tracking System: Skill + Config + Orchestrator** — agents_skills_beads_skill_beads_skill, _beads_config_beads_config, _beads_orchestrator_beads_orchestrator, concept_beads_dolt_sync [EXTRACTED 1.00]

## Communities (24 total, 13 thin omitted)

### Community 0 - "Nessie Iceberg Catalog Service"
Cohesion: 0.18
Nodes (21): Dremio Bootstrap Auto-Configuration Service, Dremio Query Engine Service, Flink JobManager Service, Flink SQL Submit Service (ingest.sql runner), Flink TaskManager Service, Shopify Data Generator Service, Kafka Topic Initialization Service, Kafka Broker Service (KRaft mode) (+13 more)

### Community 1 - "Registry"
Cohesion: 0.20
Nodes (14): Faker, Generator, Registry, VariantRef, handle(), NewGenerator(), NewRegistry(), strPtr() (+6 more)

### Community 2 - "Load"
Cohesion: 0.24
Nodes (16): Config, InventoryConfig, KafkaConfig, OrderDetailsConfig, ProductsConfig, applyEnvOverrides(), applyNonZero(), defaults() (+8 more)

### Community 3 - "Producer"
Cohesion: 0.23
Nodes (8): Client, encode(), Event, OrderDetail, loadAndRegister(), New(), Producer, Schema

### Community 4 - "OrderDetailsConsumer"
Cohesion: 0.21
Nodes (5): Base, BaseConsumer, ApplicationConsumer, OrderDetailsConsumer, OrderDetail

### Community 5 - ".NewOrderDetail"
Cohesion: 0.20
Nodes (8): Duration, OrderDetail, shopifyTime(), InventoryLevel, OrderDetail, Product, Variant, Time

### Community 6 - "Agent Instructions (AGENTS.md)"
Cohesion: 0.25
Nodes (8): Beads Configuration, Beads Orchestrator Config, Beads Issue Tracker - AI-Native Issue Tracking, Agent Instructions (AGENTS.md), Beads Skill - Task Tracking Workflow, Project Instructions for AI Agents (CLAUDE.md), Beads Conservative Agent Profile, Beads Dolt Sync Architecture

### Community 7 - "main"
Cohesion: 0.50
Nodes (4): Event, logDeliveryEvents(), main(), Int64

### Community 8 - "compact.py"
Cohesion: 0.50
Nodes (4): compact_table(), Spark compaction script for Iceberg/Nessie tables.  Runs three maintenance proce, Execute a SQL statement, print a summary, and return the result DataFrame., run_sql()

## Knowledge Gaps
- **13 isolated node(s):** `generator`, `init-bucket.sh script`, `entrypoint.sh script`, `Beads OpenAI Agent Interface`, `Beads Issue Tracker - AI-Native Issue Tracking` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `main` to `Registry`, `Load`, `Producer`?**
  _High betweenness centrality (0.090) - this node is a cross-community bridge._
- **Why does `New()` connect `Producer` to `Load`, `main`?**
  _High betweenness centrality (0.060) - this node is a cross-community bridge._
- **Why does `Producer` connect `Producer` to `.NewOrderDetail`?**
  _High betweenness centrality (0.050) - this node is a cross-community bridge._
- **Are the 5 inferred relationships involving `Load()` (e.g. with `main()` and `TestLoadDefaults()`) actually correct?**
  _`Load()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **Are the 5 inferred relationships involving `NewGenerator()` (e.g. with `main()` and `TestNewInventoryLevelEmptyRegistry()`) actually correct?**
  _`NewGenerator()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `generator`, `init-bucket.sh script`, `entrypoint.sh script` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._