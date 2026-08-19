# Graph Report - .  (2026-08-19)

## Corpus Check
- 58 files · ~75,441 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 410 nodes · 575 edges · 49 communities (25 shown, 24 thin omitted)
- Extraction: 91% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 48 edges (avg confidence: 0.82)
- Token cost: 327,233 input · 0 output

## Community Hubs (Navigation)
- Webhook Data Explorer UI
- Monitoring Stack Services & Alerts
- Webhook Build Tooling
- Generator Fake-Data Engine
- Webhook npm Dependencies
- TypeScript Config & Next Types
- Generator Config Schema
- Webhook Receiver API
- Webhook UI Component Library
- Generator Main Loop & Stats
- Kafka Producer Client
- Ruby Kafka Consumer Base
- Flink Dashboard Panels & Jobs
- Beads/Agent Instructions Docs
- Dremio/MinIO/Spark Bootstrap Services
- Flink Pipeline Deployment Config
- Webhook Dashboard Page
- Dremio Query UI & Tables
- Spark Iceberg Compaction Script
- Explorer Data API
- Webhook Metrics API
- Karafka App Entrypoint
- Explorer Tables API
- Webhook Root Layout
- post-checkout Git Hook
- post-merge Git Hook
- pre-commit Git Hook
- pre-push Git Hook
- prepare-commit-msg Git Hook
- MinIO Warehouse Bucket
- Dremio Bootstrap Script
- Kafka Topic Init Script
- MinIO Bucket Init Script
- Spark Entrypoint Script
- Next.js Config File
- Next.js Env Types
- Tailwind Config File
- Beads OpenAI Agent Config
- Kafka Dashboard Screenshot
- Generator Duration Type
- OrderDetail Model (Go)
- OrderDetail Model (Ruby)
- Generator Mutex Type
- Generator Binary/Service
- Generator Time Type

## God Nodes (most connected - your core abstractions)
1. `cn()` - 30 edges
2. `compilerOptions` - 16 edges
3. `main()` - 11 edges
4. `Config` - 11 edges
5. `Load()` - 11 edges
6. `Registry` - 11 edges
7. `Producer` - 11 edges
8. `EventID` - 10 edges
9. `New()` - 10 edges
10. `errorTracker` - 9 edges

## Surprising Connections (you probably didn't know these)
- `Apache Flink Dashboard Overview Screenshot` --conceptually_related_to--> `compose/flink.yml`  [INFERRED]
  docs/screenshots/flink.png → compose/flink.yml
- `Beads Skill - Task Tracking Workflow` --semantically_similar_to--> `Agent Instructions (AGENTS.md)`  [INFERRED] [semantically similar]
  .agents/skills/beads/SKILL.md → AGENTS.md
- `shopify-lakehouse README` --references--> `dev-guards config (.dev-guards/config.yml)`  [AMBIGUOUS]
  README.md → .dev-guards/config.yml
- `OrderDetailsConsumer (Karafka consumer)` --semantically_similar_to--> `Webhook Service (Next.js Shopify webhook receiver)`  [INFERRED] [semantically similar]
  ruby-kafka-consumer/README.md → webhook-service/README.md
- `Apache Flink Dashboard Overview Screenshot` --conceptually_related_to--> `flink/sql/ingest.sql`  [INFERRED]
  docs/screenshots/flink.png → flink/sql/ingest.sql

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Monitoring & alerting stack (Prometheus, Grafana, Kafka exporters)** — compose_grafana_grafana, compose_grafana_prometheus, compose_kafka_kafka_jmx_exporter, compose_kafka_kafka_exporter, kafka_jmx_exporter_config_config, compose_prometheus_scrape_config [EXTRACTED 1.00]
- **Flink streaming ingest pipeline into Nessie/Iceberg** — compose_flink_flink_jobmanager, compose_flink_flink_taskmanager, compose_flink_flink_sql_submit, compose_nessie_nessie, compose_kafka_kafka [EXTRACTED 1.00]
- **Grafana alert rules for the lakehouse pipeline** — compose_grafana_provisioning_alerting_rules_flink_job_restarts, compose_grafana_provisioning_alerting_rules_kafka_consumer_lag_high, compose_grafana_provisioning_alerting_rules_generator_backpressure_active [EXTRACTED 1.00]
- **Beads Task Tracking System: Skill + Config + Orchestrator** — agents_skills_beads_skill_beads_skill, _beads_config_beads_config, _beads_orchestrator_beads_orchestrator, concept_beads_dolt_sync [EXTRACTED 1.00]

## Communities (49 total, 24 thin omitted)

### Community 0 - "Webhook Data Explorer UI"
Cohesion: 0.09
Nodes (31): DataRow, QueryResult, TableInfo, DialogContent(), DialogDescription(), DialogFooter(), DialogHeader(), DialogOverlay() (+23 more)

### Community 1 - "Monitoring Stack Services & Alerts"
Cohesion: 0.10
Nodes (36): flink-jobmanager service, flink-sql-submit service, flink-taskmanager service, generator service, grafana service, prometheus service (defined in grafana.yml), Alert rule: Flink job restarting repeatedly, Alert rule: Generator backpressure active (+28 more)

### Community 2 - "Webhook Build Tooling"
Cohesion: 0.06
Nodes (34): autoprefixer, postcss, prisma, tailwindcss, tsx, @types/better-sqlite3, @types/crypto-js, @types/node (+26 more)

### Community 3 - "Generator Fake-Data Engine"
Cohesion: 0.14
Nodes (22): Faker, Generator, Registry, VariantRef, Mutex, Time, handle(), NewGenerator() (+14 more)

### Community 4 - "Webhook npm Dependencies"
Cohesion: 0.07
Nodes (29): better-sqlite3, class-variance-authority, clsx, crypto-js, lucide-react, next, @prisma/client, @radix-ui/react-dialog (+21 more)

### Community 5 - "TypeScript Config & Next Types"
Cohesion: 0.07
Nodes (26): dom, dom.iterable, esnext, next-env.d.ts, .next/types/**/*.ts, node_modules, **/*.ts, **/*.tsx (+18 more)

### Community 6 - "Generator Config Schema"
Cohesion: 0.21
Nodes (18): Config, CustomersConfig, InventoryConfig, KafkaConfig, OrderDetailsConfig, ProductsConfig, applyEnvOverrides(), applyNonZero() (+10 more)

### Community 7 - "Webhook Receiver API"
Cohesion: 0.19
Nodes (18): getOperation(), payloadHash(), POST(), prisma, ShopifyCustomer, ShopifyInventoryLevel, ShopifyOrder, ShopifyOrderLineItem (+10 more)

### Community 8 - "Webhook UI Component Library"
Cohesion: 0.12
Nodes (16): aliases, components, hooks, lib, ui, utils, iconLibrary, rsc (+8 more)

### Community 9 - "Generator Main Loop & Stats"
Cohesion: 0.19
Nodes (11): Generator Performance Dashboard, Duration, Event, Mutex, Time, logDeliveryEvents(), main(), newErrorTracker() (+3 more)

### Community 10 - "Kafka Producer Client"
Cohesion: 0.23
Nodes (7): Client, encode(), Event, loadAndRegister(), New(), Producer, Schema

### Community 11 - "Ruby Kafka Consumer Base"
Cohesion: 0.21
Nodes (5): Base, BaseConsumer, ApplicationConsumer, OrderDetailsConsumer, OrderDetail

### Community 12 - "Flink Dashboard Panels & Jobs"
Cohesion: 0.47
Nodes (10): Grafana Flink Dashboard Provisioning Config, Checkpoints Completed vs Failed Panel, Flink Job Health Dashboard (Grafana), Job Restarts Panel, Job Uptime Panel, Last Checkpoint Duration Panel, insert_into_nessie_lakehouse_customers (Flink job), insert_into_nessie_lakehouse_order_details (Flink job) (+2 more)

### Community 13 - "Beads/Agent Instructions Docs"
Cohesion: 0.25
Nodes (8): Beads Configuration, Beads Orchestrator Config, Beads Issue Tracker - AI-Native Issue Tracking, Agent Instructions (AGENTS.md), Beads Skill - Task Tracking Workflow, Project Instructions for AI Agents (CLAUDE.md), Beads Conservative Agent Profile, Beads Dolt Sync Architecture

### Community 14 - "Dremio/MinIO/Spark Bootstrap Services"
Cohesion: 0.38
Nodes (7): Dremio Bootstrap Auto-Configuration Service, Dremio Query Engine Service, MinIO Bucket Initialization Service, MinIO Object Storage Service, Spark Iceberg Compaction Service, Spark Compaction of Flink Small Files, Dremio Manual Setup Guide

### Community 15 - "Flink Pipeline Deployment Config"
Cohesion: 0.43
Nodes (7): compose/flink.yml, Apache Flink Dashboard Overview Screenshot, insert-into_nessie.lakehouse.product_variants Flink Job, insert-into_nessie.lakehouse.products,nessie.lakehouse.inventory_levels Flink Job, Nessie Lakehouse Catalog (nessie.lakehouse), flink/Dockerfile, flink/sql/ingest.sql

### Community 16 - "Webhook Dashboard Page"
Cohesion: 0.53
Nodes (5): DashboardPage(), getHealthStats(), getRecentEvents(), getResourceCounts(), prisma

### Community 17 - "Dremio Query UI & Tables"
Cohesion: 0.50
Nodes (5): Dremio Query UI (products table), inventory_levels Table, nessie.lakehouse Catalog (Nessie/Iceberg), product_variants Table, products Table

### Community 18 - "Spark Iceberg Compaction Script"
Cohesion: 0.50
Nodes (4): compact_table(), Spark compaction script for Iceberg/Nessie tables.  Runs three maintenance proce, Execute a SQL statement, print a summary, and return the result DataFrame., run_sql()

### Community 19 - "Explorer Data API"
Cohesion: 0.50
Nodes (4): DATE_FIELDS, GET(), MODEL_FIELDS, prisma

### Community 20 - "Webhook Metrics API"
Cohesion: 0.67
Nodes (3): formatPrometheusMetrics(), GET(), prisma

## Ambiguous Edges - Review These
- `dev-guards config (.dev-guards/config.yml)` → `shopify-lakehouse README`  [AMBIGUOUS]
  README.md · relation: references

## Knowledge Gaps
- **116 isolated node(s):** `init-bucket.sh script`, `entrypoint.sh script`, `Beads OpenAI Agent Interface`, `Beads Issue Tracker - AI-Native Issue Tracking`, `Beads Configuration` (+111 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `dev-guards config (.dev-guards/config.yml)` and `shopify-lakehouse README`?**
  _Edge tagged AMBIGUOUS (relation: references) - confidence is low._
- **Why does `main()` connect `Generator Main Loop & Stats` to `Kafka Producer Client`, `Generator Fake-Data Engine`, `Generator Config Schema`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `New()` connect `Kafka Producer Client` to `Generator Main Loop & Stats`, `Generator Fake-Data Engine`, `Generator Config Schema`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `dependencies` connect `Webhook npm Dependencies` to `Webhook Build Tooling`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `main()` (e.g. with `Load()` and `ParseRate()`) actually correct?**
  _`main()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `init-bucket.sh script`, `entrypoint.sh script`, `Beads OpenAI Agent Interface` to the rest of the system?**
  _116 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Webhook Data Explorer UI` be split into smaller, more focused modules?**
  _Cohesion score 0.09408033826638477 - nodes in this community are weakly interconnected._