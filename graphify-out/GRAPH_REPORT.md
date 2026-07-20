# Graph Report - .  (2026-07-19)

## Corpus Check
- Corpus is ~10,120 words - fits in a single context window. You may not need a graph.

## Summary
- 119 nodes · 184 edges · 19 communities (7 shown, 12 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 32 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Docker Service Orchestration
- Shopify Event Producer
- Config & Kafka Settings
- Data Generators & Faker
- Generator CLI Entrypoint
- Agent & Beads Workflow Config
- Spark Iceberg Compaction
- Beads Post-Checkout Hook
- Beads Post-Merge Hook
- Beads Pre-Commit Hook
- Beads Pre-Push Hook
- Beads Prepare-Commit-Msg Hook
- Dremio Bootstrap Script
- Kafka Init Topics Script
- MinIO Init Bucket Script
- Spark Entrypoint Script
- OpenAI Beads Agent
- Generator Package Root
- Project README

## God Nodes (most connected - your core abstractions)
1. `Load()` - 11 edges
2. `Registry` - 10 edges
3. `Config` - 9 edges
4. `NewGenerator()` - 9 edges
5. `Producer` - 9 edges
6. `Nessie Iceberg Catalog Service` - 8 edges
7. `main()` - 7 edges
8. `NewRegistry()` - 7 edges
9. `MinIO Object Storage Service` - 7 edges
10. `Shopify Lakehouse Docker Compose Stack` - 7 edges

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

## Communities (19 total, 12 thin omitted)

### Community 0 - "Docker Service Orchestration"
Cohesion: 0.18
Nodes (21): Dremio Bootstrap Auto-Configuration Service, Dremio Query Engine Service, Flink JobManager Service, Flink SQL Submit Service (ingest.sql runner), Flink TaskManager Service, Shopify Data Generator Service, Kafka Topic Initialization Service, Kafka Broker Service (KRaft mode) (+13 more)

### Community 1 - "Shopify Event Producer"
Cohesion: 0.19
Nodes (10): Client, encode(), Event, loadAndRegister(), New(), InventoryLevel, Product, Variant (+2 more)

### Community 2 - "Config & Kafka Settings"
Cohesion: 0.26
Nodes (15): Config, InventoryConfig, KafkaConfig, ProductsConfig, applyEnvOverrides(), applyNonZero(), defaults(), Load() (+7 more)

### Community 3 - "Data Generators & Faker"
Cohesion: 0.21
Nodes (10): Duration, Faker, Generator, Registry, VariantRef, handle(), shopifyTime(), strPtr() (+2 more)

### Community 4 - "Generator CLI Entrypoint"
Cohesion: 0.29
Nodes (11): Event, logDeliveryEvents(), main(), NewGenerator(), NewRegistry(), T, TestNewInventoryLevelEmptyRegistry(), TestNewInventoryLevelReferencesKnownVariant() (+3 more)

### Community 5 - "Agent & Beads Workflow Config"
Cohesion: 0.25
Nodes (8): Beads Configuration, Beads Orchestrator Config, Beads Issue Tracker - AI-Native Issue Tracking, Agent Instructions (AGENTS.md), Beads Skill - Task Tracking Workflow, Project Instructions for AI Agents (CLAUDE.md), Beads Conservative Agent Profile, Beads Dolt Sync Architecture

### Community 6 - "Spark Iceberg Compaction"
Cohesion: 0.50
Nodes (4): compact_table(), Spark compaction script for Iceberg/Nessie tables.  Runs three maintenance proce, Execute a SQL statement, print a summary, and return the result DataFrame., run_sql()

## Knowledge Gaps
- **12 isolated node(s):** `bootstrap.sh script`, `generator`, `init-topics.sh script`, `init-bucket.sh script`, `entrypoint.sh script` (+7 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Generator CLI Entrypoint` to `Shopify Event Producer`, `Config & Kafka Settings`?**
  _High betweenness centrality (0.120) - this node is a cross-community bridge._
- **Why does `New()` connect `Shopify Event Producer` to `Config & Kafka Settings`, `Generator CLI Entrypoint`?**
  _High betweenness centrality (0.073) - this node is a cross-community bridge._
- **Why does `Load()` connect `Config & Kafka Settings` to `Generator CLI Entrypoint`?**
  _High betweenness centrality (0.065) - this node is a cross-community bridge._
- **Are the 5 inferred relationships involving `Load()` (e.g. with `main()` and `TestLoadDefaults()`) actually correct?**
  _`Load()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **Are the 5 inferred relationships involving `NewGenerator()` (e.g. with `main()` and `TestNewInventoryLevelEmptyRegistry()`) actually correct?**
  _`NewGenerator()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `bootstrap.sh script`, `generator`, `init-topics.sh script` to the rest of the system?**
  _12 weakly-connected nodes found - possible documentation gaps or missing edges._