# Graph Report - .  (2026-08-20)

## Corpus Check
- 68 files · ~72,957 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 279 nodes · 457 edges · 35 communities (20 shown, 15 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 45 edges (avg confidence: 0.82)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Webhook Data Explorer UI
- Monitoring Stack Services & Alerts
- Generator Fake-Data Engine
- Generator Config Schema
- Webhook Receiver API
- Generator Main Loop & Stats
- Kafka Producer Client
- Ruby Kafka Consumer Base
- Flink Dashboard Panels & Jobs
- Dremio/MinIO/Spark Bootstrap Services
- Flink Pipeline Deployment Config
- Webhook Dashboard Page
- Dremio Query UI & Tables
- Spark Iceberg Compaction Script
- Explorer Data API
- Project Agent Instructions
- Webhook Metrics API
- Karafka App Entrypoint
- Explorer Tables API
- Webhook Root Layout
- MinIO Warehouse Bucket
- Dremio Bootstrap Script
- Kafka Topic Init Script
- MinIO Bucket Init Script
- Spark Entrypoint Script
- Kafka Dashboard Screenshot
- Generator Duration Type
- OrderDetail Model (Go)
- OrderDetail Model (Ruby)
- Generator Mutex Type
- Generator Binary/Service
- Generator Time Type

## God Nodes (most connected - your core abstractions)
1. `cn()` - 30 edges
2. `main()` - 11 edges
3. `Config` - 11 edges
4. `Load()` - 11 edges
5. `Registry` - 11 edges
6. `Producer` - 11 edges
7. `EventID` - 10 edges
8. `New()` - 10 edges
9. `errorTracker` - 9 edges
10. `NewGenerator()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `Apache Flink Dashboard Overview Screenshot` --conceptually_related_to--> `compose/flink.yml`  [INFERRED]
  docs/screenshots/flink.png → compose/flink.yml
- `OrderDetailsConsumer (Karafka consumer)` --semantically_similar_to--> `Webhook Service (Next.js Shopify webhook receiver)`  [INFERRED] [semantically similar]
  ruby-kafka-consumer/README.md → webhook-service/README.md
- `Apache Flink Dashboard Overview Screenshot` --conceptually_related_to--> `flink/sql/ingest.sql`  [INFERRED]
  docs/screenshots/flink.png → flink/sql/ingest.sql
- `insert-into_nessie.lakehouse.product_variants Flink Job` --conceptually_related_to--> `flink/sql/ingest.sql`  [INFERRED]
  docs/screenshots/flink.png → flink/sql/ingest.sql
- `insert-into_nessie.lakehouse.products,nessie.lakehouse.inventory_levels Flink Job` --conceptually_related_to--> `flink/sql/ingest.sql`  [INFERRED]
  docs/screenshots/flink.png → flink/sql/ingest.sql

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Monitoring & alerting stack (Prometheus, Grafana, Kafka exporters)** — compose_grafana_grafana, compose_grafana_prometheus, compose_kafka_kafka_jmx_exporter, compose_kafka_kafka_exporter, kafka_jmx_exporter_config_config, compose_prometheus_scrape_config [EXTRACTED 1.00]
- **Flink streaming ingest pipeline into Nessie/Iceberg** — compose_flink_flink_jobmanager, compose_flink_flink_taskmanager, compose_flink_flink_sql_submit, compose_nessie_nessie, compose_kafka_kafka [EXTRACTED 1.00]
- **Grafana alert rules for the lakehouse pipeline** — compose_grafana_provisioning_alerting_rules_flink_job_restarts, compose_grafana_provisioning_alerting_rules_kafka_consumer_lag_high, compose_grafana_provisioning_alerting_rules_generator_backpressure_active [EXTRACTED 1.00]

## Communities (35 total, 15 thin omitted)

### Community 0 - "Webhook Data Explorer UI"
Cohesion: 0.09
Nodes (31): DataRow, QueryResult, TableInfo, DialogContent(), DialogDescription(), DialogFooter(), DialogHeader(), DialogOverlay() (+23 more)

### Community 1 - "Monitoring Stack Services & Alerts"
Cohesion: 0.10
Nodes (35): flink-jobmanager service, flink-sql-submit service, flink-taskmanager service, generator service, grafana service, prometheus service (defined in grafana.yml), Alert rule: Flink job restarting repeatedly, Alert rule: Generator backpressure active (+27 more)

### Community 2 - "Generator Fake-Data Engine"
Cohesion: 0.14
Nodes (23): Faker, Generator, Registry, VariantRef, Mutex, Time, handle(), NewGenerator() (+15 more)

### Community 3 - "Generator Config Schema"
Cohesion: 0.21
Nodes (18): Config, CustomersConfig, InventoryConfig, KafkaConfig, OrderDetailsConfig, ProductsConfig, applyEnvOverrides(), applyNonZero() (+10 more)

### Community 4 - "Webhook Receiver API"
Cohesion: 0.19
Nodes (18): getOperation(), payloadHash(), POST(), prisma, ShopifyCustomer, ShopifyInventoryLevel, ShopifyOrder, ShopifyOrderLineItem (+10 more)

### Community 5 - "Generator Main Loop & Stats"
Cohesion: 0.19
Nodes (11): Generator Performance Dashboard, Duration, Event, Mutex, Time, logDeliveryEvents(), main(), newErrorTracker() (+3 more)

### Community 6 - "Kafka Producer Client"
Cohesion: 0.22
Nodes (6): Client, encode(), Event, loadAndRegister(), Producer, Schema

### Community 7 - "Ruby Kafka Consumer Base"
Cohesion: 0.21
Nodes (5): Base, BaseConsumer, ApplicationConsumer, OrderDetailsConsumer, OrderDetail

### Community 8 - "Flink Dashboard Panels & Jobs"
Cohesion: 0.47
Nodes (10): Grafana Flink Dashboard Provisioning Config, Checkpoints Completed vs Failed Panel, Flink Job Health Dashboard (Grafana), Job Restarts Panel, Job Uptime Panel, Last Checkpoint Duration Panel, insert_into_nessie_lakehouse_customers (Flink job), insert_into_nessie_lakehouse_order_details (Flink job) (+2 more)

### Community 9 - "Dremio/MinIO/Spark Bootstrap Services"
Cohesion: 0.38
Nodes (7): Dremio Bootstrap Auto-Configuration Service, Dremio Query Engine Service, MinIO Bucket Initialization Service, MinIO Object Storage Service, Spark Iceberg Compaction Service, Spark Compaction of Flink Small Files, Dremio Manual Setup Guide

### Community 10 - "Flink Pipeline Deployment Config"
Cohesion: 0.43
Nodes (7): compose/flink.yml, Apache Flink Dashboard Overview Screenshot, insert-into_nessie.lakehouse.product_variants Flink Job, insert-into_nessie.lakehouse.products,nessie.lakehouse.inventory_levels Flink Job, Nessie Lakehouse Catalog (nessie.lakehouse), flink/Dockerfile, flink/sql/ingest.sql

### Community 11 - "Webhook Dashboard Page"
Cohesion: 0.53
Nodes (5): DashboardPage(), getHealthStats(), getRecentEvents(), getResourceCounts(), prisma

### Community 12 - "Dremio Query UI & Tables"
Cohesion: 0.50
Nodes (5): Dremio Query UI (products table), inventory_levels Table, nessie.lakehouse Catalog (Nessie/Iceberg), product_variants Table, products Table

### Community 13 - "Spark Iceberg Compaction Script"
Cohesion: 0.50
Nodes (4): compact_table(), Spark compaction script for Iceberg/Nessie tables.  Runs three maintenance proce, Execute a SQL statement, print a summary, and return the result DataFrame., run_sql()

### Community 14 - "Explorer Data API"
Cohesion: 0.50
Nodes (4): DATE_FIELDS, GET(), MODEL_FIELDS, prisma

### Community 15 - "Project Agent Instructions"
Cohesion: 0.50
Nodes (4): Agent Instructions (AGENTS.md), Project Instructions for AI Agents (CLAUDE.md), Beads Conservative Agent Profile, Beads Dolt Sync Architecture

### Community 16 - "Webhook Metrics API"
Cohesion: 0.67
Nodes (3): formatPrometheusMetrics(), GET(), prisma

## Knowledge Gaps
- **38 isolated node(s):** `init-bucket.sh script`, `entrypoint.sh script`, `MinIO Bucket Initialization Service`, `bootstrap.sh script`, `generator` (+33 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Generator Main Loop & Stats` to `Generator Fake-Data Engine`, `Generator Config Schema`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `New()` connect `Generator Fake-Data Engine` to `Generator Config Schema`, `Generator Main Loop & Stats`, `Kafka Producer Client`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Why does `Load()` connect `Generator Config Schema` to `Generator Main Loop & Stats`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `main()` (e.g. with `Load()` and `ParseRate()`) actually correct?**
  _`main()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 5 inferred relationships involving `Load()` (e.g. with `main()` and `TestLoadDefaults()`) actually correct?**
  _`Load()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `init-bucket.sh script`, `entrypoint.sh script`, `MinIO Bucket Initialization Service` to the rest of the system?**
  _38 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Webhook Data Explorer UI` be split into smaller, more focused modules?**
  _Cohesion score 0.09408033826638477 - nodes in this community are weakly interconnected._