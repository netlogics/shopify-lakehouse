# Graph Report - .  (2026-09-01)

## Corpus Check
- 42 files · ~78,669 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 350 nodes · 521 edges · 65 communities (39 shown, 26 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 48 edges (avg confidence: 0.81)
- Token cost: 84,916 input · 0 output

## Community Hubs (Navigation)
- Webhook Explorer UI (Next.js)
- Generator Core (event synthesis + fraud)
- Generator Configuration
- Monitoring & Ops Docs (alerts, changelog, hardening)
- Core Pipeline Services (Flink/Kafka/Grafana)
- Webhook Ingestion API
- Kafka Producer Client
- Generator Main Loop & Error Tracking
- Generator Test Suite
- Grafana Flink Dashboard
- Flink SQL Ingest (Iceberg sink DDL)
- Ruby Order Details Consumer
- Dremio/MinIO/Spark Bootstrap Services
- Flink Deployment (Dockerfile/compose)
- Webhook Dashboard Page
- dbt & Airflow Compose Wiring
- Dremio Query UI / Iceberg Tables
- Spark Compaction Script
- Webhook Explorer Data API
- Agent/Beads Instructions
- Webhook Prometheus Metrics API
- Karafka App Bootstrap
- Ruby OrderDetail Model
- Ruby Base Consumer
- Generator Metrics Dashboard
- Webhook Explorer Tables API
- Webhook App Root Layout
- Airflow Cosmos DAG
- dbt Entrypoint Script
- Airflow Entrypoint Script
- MinIO Warehouse Bucket
- Dremio Bootstrap Script
- Kafka Topic Init Script
- MinIO Bucket Init Script
- Pipeline Reset Script
- Grafana Bootstrap Entrypoint
- Grafana Dashboard Provisioning
- Grafana Prometheus Datasource
- Grafana Kafka Lag Dashboard
- Generator Duration Type
- Generator Event Type
- Ruby OrderDetail (dup ref)
- Ruby OrderDetail (dup ref)
- Generator Mutex Type
- Generator Service Node
- Generator Time Type

## God Nodes (most connected - your core abstractions)
1. `cn()` - 30 edges
2. `Registry` - 17 edges
3. `Load()` - 15 edges
4. `NewGenerator()` - 14 edges
5. `Config` - 12 edges
6. `NewRegistry()` - 12 edges
7. `Producer` - 11 edges
8. `Generator` - 10 edges
9. `EventID` - 10 edges
10. `errorTracker` - 9 edges

## Surprising Connections (you probably didn't know these)
- `OrderDetailsConsumer (Karafka consumer)` --semantically_similar_to--> `Webhook Service (Next.js Shopify webhook receiver)`  [INFERRED] [semantically similar]
  ruby-kafka-consumer/README.md → webhook-service/README.md
- `Apache Flink Dashboard Overview Screenshot` --conceptually_related_to--> `compose/flink.yml`  [INFERRED]
  docs/screenshots/flink.png → compose/flink.yml
- `Agent Instructions (AGENTS.md)` --references--> `Project Instructions for AI Agents (CLAUDE.md)`  [INFERRED]
  AGENTS.md → CLAUDE.md
- `Dremio Manual Setup Guide` --references--> `MinIO Object Storage Service`  [EXTRACTED]
  dremio/setup.md → compose/minio.yml
- `jmx_exporter httpserver config for Kafka` --references--> `kafka broker service`  [EXTRACTED]
  kafka/jmx-exporter/config.yml → compose/kafka.yml

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Services that depend on dremio-bootstrap to query the Nessie/Dremio source** — compose_dbt_doc, compose_airflow_doc, dbt_profiles_doc [INFERRED 0.85]
- **Three Grafana alert rules covering the pipeline (Flink restarts, Kafka lag, generator backpressure)** — compose_grafana_provisioning_alerting_rules_flink_job_restarts, compose_grafana_provisioning_alerting_rules_kafka_consumer_lag_high, compose_grafana_provisioning_alerting_rules_generator_backpressure_active, readme_monitoring_stack [EXTRACTED 1.00]
- **Files that pin dbt-core/dbt-dremio versions consistently across the dbt-running services** — airflow_requirements_doc, dbt_requirements_doc, dbt_dbt_project_doc [INFERRED 0.85]
- **Flink streaming ingest pipeline into Nessie/Iceberg** — compose_flink_flink_jobmanager, compose_flink_flink_taskmanager, compose_flink_flink_sql_submit, compose_nessie_nessie, compose_kafka_kafka [EXTRACTED 1.00]

## Communities (65 total, 26 thin omitted)

### Community 0 - "Webhook Explorer UI (Next.js)"
Cohesion: 0.09
Nodes (31): DataRow, QueryResult, TableInfo, DialogContent(), DialogDescription(), DialogFooter(), DialogHeader(), DialogOverlay() (+23 more)

### Community 1 - "Generator Core (event synthesis + fraud)"
Cohesion: 0.14
Nodes (17): Faker, FraudParams, Generator, Registry, VariantRef, Duration, Mutex, Time (+9 more)

### Community 2 - "Generator Configuration"
Cohesion: 0.18
Nodes (23): Config, CustomersConfig, FraudConfig, InventoryConfig, KafkaConfig, OrderDetailsConfig, ProductsConfig, applyEnvOverrides() (+15 more)

### Community 3 - "Monitoring & Ops Docs (alerts, changelog, hardening)"
Cohesion: 0.10
Nodes (22): airflow/requirements.txt (astronomer-cosmos, dbt-core, dbt-dremio pins), Rationale: Dremio S3 endpoint must be bare host:port, not http:// URL, to prevent silent AWS fallback, Generator hardening (bounded variant lookup, dedup, backpressure, collision detection), compose/grafana/provisioning/alerting/rules.yml (Grafana unified alerting rules), Alert rule: Flink job restarting repeatedly, Alert rule: Generator backpressure active, Alert rule: Kafka consumer lag above 1000, Rationale: Reduce stage required before Threshold condition in Grafana alert rules (+14 more)

### Community 4 - "Core Pipeline Services (Flink/Kafka/Grafana)"
Cohesion: 0.13
Nodes (22): flink-jobmanager service, flink-sql-submit service, flink-taskmanager service, generator service, grafana service, prometheus service (defined in grafana.yml), kafka broker service, kafka-exporter service (+14 more)

### Community 5 - "Webhook Ingestion API"
Cohesion: 0.19
Nodes (18): getOperation(), payloadHash(), POST(), prisma, ShopifyCustomer, ShopifyInventoryLevel, ShopifyOrder, ShopifyOrderLineItem (+10 more)

### Community 6 - "Kafka Producer Client"
Cohesion: 0.23
Nodes (7): Client, encode(), Event, loadAndRegister(), New(), Producer, Schema

### Community 7 - "Generator Main Loop & Error Tracking"
Cohesion: 0.26
Nodes (9): Event, Duration, Mutex, Time, logDeliveryEvents(), main(), newErrorTracker(), errorTracker (+1 more)

### Community 8 - "Generator Test Suite"
Cohesion: 0.46
Nodes (12): NewGenerator(), NewRegistry(), T, TestFraudEpisodeConcentratesOrdersOnTargetCustomer(), TestMaybeTriggerFraudDoesNotOverlapEpisodes(), TestNewInventoryLevelEmptyRegistry(), TestNewInventoryLevelReferencesKnownVariant(), TestNewOrderDetailEmptyCustomerRegistry() (+4 more)

### Community 9 - "Grafana Flink Dashboard"
Cohesion: 0.47
Nodes (10): Grafana Flink Dashboard Provisioning Config, Checkpoints Completed vs Failed Panel, Flink Job Health Dashboard (Grafana), Job Restarts Panel, Job Uptime Panel, Last Checkpoint Duration Panel, insert_into_nessie_lakehouse_customers (Flink job), insert_into_nessie_lakehouse_order_details (Flink job) (+2 more)

### Community 10 - "Flink SQL Ingest (Iceberg sink DDL)"
Cohesion: 0.20
Nodes (9): customers_source, inventory_source, nessie.lakehouse.customers, nessie.lakehouse.inventory_levels, nessie.lakehouse.order_details, nessie.lakehouse.product_variants, nessie.lakehouse.products, order_details_source (+1 more)

### Community 12 - "Dremio/MinIO/Spark Bootstrap Services"
Cohesion: 0.38
Nodes (7): Dremio Bootstrap Auto-Configuration Service, Dremio Query Engine Service, MinIO Bucket Initialization Service, MinIO Object Storage Service, Spark Iceberg Compaction Service, Spark Compaction of Flink Small Files, Dremio Manual Setup Guide

### Community 13 - "Flink Deployment (Dockerfile/compose)"
Cohesion: 0.40
Nodes (6): compose/flink.yml, Apache Flink Dashboard Overview Screenshot, insert-into_nessie.lakehouse.product_variants Flink Job, insert-into_nessie.lakehouse.products,nessie.lakehouse.inventory_levels Flink Job, Nessie Lakehouse Catalog (nessie.lakehouse), flink/Dockerfile

### Community 14 - "Webhook Dashboard Page"
Cohesion: 0.53
Nodes (5): DashboardPage(), getHealthStats(), getRecentEvents(), getResourceCounts(), prisma

### Community 15 - "dbt & Airflow Compose Wiring"
Cohesion: 0.60
Nodes (5): compose/airflow.yml (Airflow service definition), Rationale: airflow service deliberately does not depend_on flink-sql-submit (identical reasoning to compose/dbt.yml), compose/dbt.yml (dbt service definition), Rationale: dbt service deliberately does not depend_on flink-sql-submit (would wipe/recreate Iceberg tables before every dbt run), docker-compose.yml (root compose, includes all service files)

### Community 16 - "Dremio Query UI / Iceberg Tables"
Cohesion: 0.50
Nodes (5): Dremio Query UI (products table), inventory_levels Table, nessie.lakehouse Catalog (Nessie/Iceberg), product_variants Table, products Table

### Community 17 - "Spark Compaction Script"
Cohesion: 0.50
Nodes (4): compact_table(), Spark compaction script for Iceberg/Nessie tables.  Runs three maintenance proce, Execute a SQL statement, print a summary, and return the result DataFrame., run_sql()

### Community 18 - "Webhook Explorer Data API"
Cohesion: 0.50
Nodes (4): DATE_FIELDS, GET(), MODEL_FIELDS, prisma

### Community 19 - "Agent/Beads Instructions"
Cohesion: 0.50
Nodes (4): Agent Instructions (AGENTS.md), Project Instructions for AI Agents (CLAUDE.md), Beads Conservative Agent Profile, Beads Dolt Sync Architecture

### Community 20 - "Webhook Prometheus Metrics API"
Cohesion: 0.67
Nodes (3): formatPrometheusMetrics(), GET(), prisma

## Knowledge Gaps
- **53 isolated node(s):** `init-bucket.sh script`, `entrypoint.sh script`, `MinIO Bucket Initialization Service`, `bootstrap.sh script`, `generator` (+48 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **26 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Generator Main Loop & Error Tracking` to `Generator Test Suite`, `Generator Configuration`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Why does `Load()` connect `Generator Configuration` to `Generator Main Loop & Error Tracking`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `NewGenerator()` connect `Generator Test Suite` to `Generator Core (event synthesis + fraud)`, `Generator Main Loop & Error Tracking`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Load()` (e.g. with `main()` and `TestLoadDefaults()`) actually correct?**
  _`Load()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `NewGenerator()` (e.g. with `main()` and `TestFraudEpisodeConcentratesOrdersOnTargetCustomer()`) actually correct?**
  _`NewGenerator()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `init-bucket.sh script`, `entrypoint.sh script`, `MinIO Bucket Initialization Service` to the rest of the system?**
  _53 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Webhook Explorer UI (Next.js)` be split into smaller, more focused modules?**
  _Cohesion score 0.09408033826638477 - nodes in this community are weakly interconnected._