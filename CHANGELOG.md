# Changelog

Notable changes to shopify-lakehouse. Moved out of README.md, which should describe what the project is rather than track recent activity.

## Referential integrity, fraud injection, and batch reporting (dbt + Airflow)

- **Generator referential integrity + fraud injection** (see [Fraud injection](README.md#fraud-injection)):
  - `order_details.customer_id` now links to a real customer registry (previously two independent random streams with no shared key)
  - Ground-truth-labeled synthetic fraud episodes (`is_synthetic_fraud`, `fraud_pattern`): a random customer's order volume spikes for a configurable window, giving downstream fraud detection a known answer key
  - `scripts/reset-pipeline.sh`: reusable Kafka/Iceberg/Flink reset procedure for schema changes (default mode for additive changes, `--full` for breaking ones)

- **Batch reporting: dbt + Airflow** (see [Batch Reporting](README.md#batch-reporting-dbt--airflow)):
  - Standalone `dbt` service (dbt-core + dbt-dremio) against the existing Dremio/Nessie source: 5 deduplicated staging models + 11 marts (revenue, inventory health, customer analytics) in a new `nessie.marts` space, kept separate from Flink-owned `nessie.lakehouse.*`
  - 44 schema tests (`not_null`, `unique`, `relationships`, `accepted_values`), two configured at `warn` severity for expected eventual-consistency lag between the independent `order_details`/`customers` Kafka streams
  - `dbt docs generate` output served statically via a small `dbt-docs` nginx container (`:8095`)
  - Standalone Airflow (`airflow standalone`, `:8090`) with an [astronomer-cosmos](https://astronomer.github.io/astronomer-cosmos/) DAG running the dbt project hourly, one Airflow task per dbt model
  - Fixed a real Docker Compose dependency pitfall: a one-shot dependency (`flink-sql-submit`) gets re-run on every invocation of a service that depends on it, which would otherwise wipe/recreate the Iceberg tables before every dbt/Airflow run — neither service depends on it

## Monitoring stack, webhook-service, generator hardening

- **Monitoring stack** — Prometheus + Grafana now cover the full pipeline (see [Monitoring](README.md#monitoring)):
  - Flink's built-in Prometheus reporter (job uptime, checkpoint duration/failures, records in/out)
  - Kafka broker metrics via a JMX exporter + consumer-group lag via `kafka-exporter` (two exporters, different metric families — documented in the README)
  - Generator instrumented with `prometheus/client_golang` (events produced, delivery errors, backpressure, all by topic)
  - 4 dashboards + 3 Grafana-native alert rules (Flink restarts, consumer lag > 1000, generator backpressure)

- **Webhook-service deployed** — Next.js 15 app on `:3456` with:
  - HMAC-verified Shopify webhook receiver (`POST /api/webhooks`)
  - Dashboard with resource counts, health status, recent events
  - Data Explorer with table browser, SQL editor, JSON/CSV export
  - Prisma schema with 6 tables (Products, Variants, InventoryLevels, OrderDetails, Customers, WebhookEvents)

- **Generator hardening** — Production-grade reliability:
  - Bounded variant lookup (5000 entry map, O(1) access)
  - Event deduplication via UUID v4 `event_id` + `committed-offset` startup
  - Backpressure with sliding window (5 errors/10s → 5s cooldown)
  - Handle collision detection with auto-suffix
  - Dedicated customer ID counter (no variant ID reuse)

- **Flink SQL improvements**:
  - Single `products_source` consumer group (removed duplicate `products_source_variants`)
  - Non-destructive table creation (`CREATE TABLE IF NOT EXISTS`)
  - All 4 Kafka topics: products, inventory, order_details, customers

- **Iceberg maintenance**:
  - Spark compaction now includes `customers` table
  - Snapshot expiry reduced from 30d → 1h (prevents metadata bloat)
  - Partitioning simplified: only `location_id` on `inventory_levels` (low-cardinality enums removed)

- **Infrastructure fixes**:
  - Dremio S3 endpoint: bare `host:port` (not `http://` URL) — prevents silent AWS fallback
  - Flink/S3A credentials from environment (no hardcoded secrets in `core-site.xml`)
  - Nessie depends on `minio-init` (ensures bucket exists)
