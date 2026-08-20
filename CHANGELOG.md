# Changelog

Notable changes to shopify-lakehouse. Moved out of README.md, which should describe what the project is rather than track recent activity.

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
