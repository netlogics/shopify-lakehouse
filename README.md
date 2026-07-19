# shopify-lakehouse

A self-contained reference implementation of a streaming ELT lakehouse pipeline using simulated Shopify data. Useful for testing and validating Iceberg-based lakehouse architectures end-to-end without a live Shopify store.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Docker network: lakehouse                       │
│                                                                         │
│  ┌─────────────┐   Avro/    ┌───────────┐  ┌──────────────────────┐   │
│  │  Generator  │──────────▶ │   Kafka   │  │   Schema Registry    │   │
│  │   (Go)      │  Confluent │  (KRaft)  │  │   (Confluent CP)     │   │
│  │             │   wire fmt │           │◀─│                      │   │
│  │  products   │            │shopify.   │  │  shopify.products    │   │
│  │  1 event/s  │            │products   │  │  -value (Avro)       │   │
│  │             │            │           │  │  shopify.inventory   │   │
│  │  inventory  │            │shopify.   │  │  -value (Avro)       │   │
│  │  10 events/s│            │inventory  │  └──────────────────────┘   │
│  └─────────────┘            └─────┬─────┘                             │
│                                   │                                    │
│                            reads (avro-confluent)                      │
│                                   │                                    │
│                                   ▼                                    │
│                    ┌──────────────────────────────┐                   │
│                    │         Flink 1.20            │                   │
│                    │   (JobManager + TaskManager)  │                   │
│                    │                               │                   │
│                    │  Job 1 (STATEMENT SET)        │                   │
│                    │  ├── products_source          │                   │
│                    │  │   → lakehouse.products     │                   │
│                    │  └── inventory_source         │                   │
│                    │      → lakehouse.inventory_   │                   │
│                    │        levels                 │                   │
│                    │                               │                   │
│                    │  Job 2                        │                   │
│                    │  └── products_source          │                   │
│                    │      → UNNEST(variants)       │                   │
│                    │      → lakehouse.product_     │                   │
│                    │        variants               │                   │
│                    └──────────────┬───────────────┘                   │
│                                   │ HadoopFileIO / s3a://             │
│                    ┌──────────────▼───────────────┐                   │
│                    │     Nessie  (Iceberg catalog) │                   │
│                    │     + version store (RocksDB) │                   │
│                    └──────────────┬───────────────┘                   │
│                                   │                                    │
│                    ┌──────────────▼───────────────┐                   │
│                    │     MinIO  (s3a://warehouse/) │                   │
│                    │     Iceberg data + metadata   │                   │
│                    └──────────────┬───────────────┘                   │
│                                   │                                    │
│          ┌────────────────────────┴────────────────────────┐          │
│          │                                                  │          │
│  ┌───────▼────────┐                              ┌─────────▼────────┐ │
│  │ Spark 3.5      │                              │  Dremio OSS      │ │
│  │ (compaction)   │                              │  (query engine)  │ │
│  │                │                              │                  │ │
│  │ Every 600s:    │                              │  SQL over        │ │
│  │ rewrite_data_  │                              │  Iceberg via     │ │
│  │ files          │                              │  Nessie source   │ │
│  │ rewrite_mani-  │                              │  :9047           │ │
│  │ fests          │                              │                  │ │
│  │ expire_snap-   │                              │  ⚠ see setup.md  │ │
│  │ shots          │                              │                  │ │
│  └────────────────┘                              └──────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Data flow summary

1. **Generator** seeds 100 products on startup, then continuously emits product and inventory events to two Kafka topics encoded as Avro using the Confluent wire format. Schemas are registered with Schema Registry on first run.
2. **Flink** runs two streaming jobs that read from Kafka, apply transformations (type casting, timestamp parsing, `body_html` dropped, variants unnested), and sink to three Iceberg tables via the Nessie catalog.
3. **Spark** runs a compaction loop every 10 minutes — bin-packing small files written by Flink's streaming checkpoints, rewriting manifests, and expiring old snapshots — across all three Iceberg tables.
4. **Dremio** provides an ad-hoc SQL query interface over the Iceberg tables via the Nessie source. See [`dremio/setup.md`](dremio/setup.md) for first-run configuration.

### Iceberg tables

| Table | Partitioned by | Source |
|---|---|---|
| `nessie.lakehouse.products` | `status` | `shopify.products` topic |
| `nessie.lakehouse.product_variants` | — | `shopify.products` topic (UNNEST) |
| `nessie.lakehouse.inventory_levels` | `location_id` | `shopify.inventory` topic |

---

## Services

| Service | Image | Port | Purpose |
|---|---|---|---|
| `kafka` | `confluentinc/cp-kafka:7.9.0` | `29092` (host) | Kafka broker (KRaft, no ZooKeeper) |
| `schema-registry` | `confluentinc/cp-schema-registry:7.9.0` | `8081` | Confluent Schema Registry |
| `kafka-init` | `confluentinc/cp-kafka` | — | Creates `shopify.products` and `shopify.inventory` topics once on startup |
| `minio` | `minio/minio` | `9000`, `9001` (console) | S3-compatible object storage; holds all Iceberg data files |
| `minio-init` | `minio/mc` | — | Creates the `warehouse` bucket once on startup |
| `nessie` | `ghcr.io/projectnessie/nessie:0.103.3` | `19120` | Iceberg catalog with Git-like branching (RocksDB backend) |
| `flink-jobmanager` | `shopify-lakehouse/flink:1.20.1` | `8082` (UI) | Flink cluster job manager |
| `flink-taskmanager` | `shopify-lakehouse/flink:1.20.1` | — | Flink task slots (2 slots) |
| `flink-sql-submit` | `shopify-lakehouse/flink:1.20.1` | — | One-shot container: submits `ingest.sql` then exits |
| `generator` | `shopify-lakehouse/generator` | — | Go service: produces faux Shopify events to Kafka |
| `spark-compaction` | `shopify-lakehouse/spark:3.5.4` | — | Periodic Iceberg compaction loop |
| `dremio` | `dremio/dremio-oss:25.2` | `9047` | SQL query engine (requires manual source setup — see below) |
| `dremio-bootstrap` | `curlimages/curl` | — | Attempts automated Dremio Nessie source registration |

---

## Prerequisites

- Docker with Compose V2 (`docker compose`)
- ~8 GB of available RAM (Dremio alone uses 4 GB by default)
- Ports `8081`, `8082`, `9000`, `9001`, `9047`, `19120`, `29092` free on the host

---

## Quickstart

```bash
# 1. Clone
git clone https://github.com/netlogics/shopify-lakehouse.git
cd shopify-lakehouse

# 2. Build custom images (Flink + Spark + Generator)
docker compose build

# 3. Start the stack
docker compose up -d

# 4. Watch service health
docker compose ps

# 5. Tail Flink job submission
docker compose logs -f flink-sql-submit

# 6. Confirm Flink jobs are running
open http://localhost:8082   # Flink UI

# 7. Browse raw files in MinIO
open http://localhost:9001   # MinIO console (minioadmin / minioadmin)

# 8. Query via Dremio (requires one-time source setup — see dremio/setup.md)
open http://localhost:9047
```

> **Dremio note:** the automated `dremio-bootstrap` container attempts to register the Nessie source on first run, but this is known to be unreliable. Follow [`dremio/setup.md`](dremio/setup.md) to configure the source manually if the bootstrap does not complete successfully.

### Useful commands

```bash
# Check all service logs
docker compose logs -f

# Run Flink SQL interactively
docker compose exec flink-jobmanager \
  /opt/flink/bin/sql-client.sh

# Query Iceberg table counts via Spark
docker compose exec spark-compaction \
  /opt/spark/bin/spark-sql \
  --conf spark.sql.catalog.nessie=org.apache.iceberg.spark.SparkCatalog \
  --conf spark.sql.catalog.nessie.catalog-impl=org.apache.iceberg.nessie.NessieCatalog \
  --conf spark.sql.catalog.nessie.uri=http://nessie:19120/api/v1 \
  --conf spark.sql.catalog.nessie.ref=main \
  --conf spark.sql.catalog.nessie.warehouse=s3a://warehouse/ \
  -e "SELECT COUNT(*) FROM nessie.lakehouse.products"

# Trigger a manual compaction run
docker compose restart spark-compaction

# Tear down (preserves data volumes)
docker compose down

# Tear down and wipe all data
docker compose down -v
```

---

## Configuration

All tuneable parameters live in `.env`. Key values:

| Variable | Default | Description |
|---|---|---|
| `KAFKA_TOPIC_PRODUCTS` | `shopify.products` | Products Kafka topic name |
| `KAFKA_TOPIC_INVENTORY` | `shopify.inventory` | Inventory Kafka topic name |
| `KAFKA_TOPIC_PARTITIONS` | `3` | Partition count for both topics |
| `KAFKA_TOPIC_RETENTION_MS` | `604800000` | Topic retention (7 days) |
| `WAREHOUSE_BUCKET` | `warehouse` | MinIO bucket for Iceberg files |
| `AWS_ACCESS_KEY_ID` | `minioadmin` | MinIO access key (also used by Flink + Spark) |
| `AWS_SECRET_ACCESS_KEY` | `minioadmin` | MinIO secret key |
| `NESSIE_VERSION_STORE_TYPE` | `ROCKSDB` | Nessie backend (`ROCKSDB` or `IN_MEMORY`) |
| `COMPACTION_INTERVAL` | `600` | Seconds between Spark compaction runs |
| `DREMIO_MAX_MEMORY_SIZE_MB` | `4096` | Dremio JVM heap limit |

Generator rates can be adjusted in `generator/config.yaml`:

| Key | Default | Description |
|---|---|---|
| `products.rate` | `1/s` | New product events per second |
| `products.seed_count` | `100` | Products published on startup |
| `inventory.rate` | `10/s` | Inventory update events per second |
| `inventory.locations` | `3` | Number of simulated warehouse locations |

---

## Flink image

The custom Flink image (`flink/Dockerfile`) extends the stock `flink:1.20.1` image with:

- `flink-sql-connector-kafka` 3.4.0-1.20
- `flink-sql-avro-confluent-registry` 1.20.1
- `iceberg-flink-runtime-1.20` 1.9.1
- `iceberg-nessie` 1.9.1 + transitive deps (resolved via Maven)
- `hadoop-aws` + `aws-java-sdk-bundle` 1.12.780 (S3A → MinIO via `HadoopFileIO`)

> `iceberg-aws-bundle` is intentionally excluded — it duplicates `iceberg-core` and causes `AbstractMethodError` due to Jackson interface conflicts with `iceberg-flink-runtime`.

---

## Project structure

```
.
├── compose/              # Per-service Docker Compose files (included by root)
│   ├── flink.yml
│   ├── generator.yml
│   ├── kafka.yml
│   ├── minio.yml
│   ├── nessie.yml
│   ├── spark.yml
│   └── dremio.yml
├── docker-compose.yml    # Root compose — includes all of the above
├── .env                  # All tuneable config
├── flink/
│   ├── Dockerfile        # Custom Flink image with Iceberg/Kafka/Nessie deps
│   ├── core-site.xml     # Hadoop S3A config for MinIO
│   └── sql/
│       └── ingest.sql    # Flink SQL: Kafka sources, Nessie catalog, Iceberg sinks
├── generator/            # Go service: Shopify event generator
│   ├── cmd/generator/    # main entrypoint
│   ├── internal/
│   │   ├── config/       # YAML config + env overrides
│   │   ├── gen/          # Fake product/inventory data generation
│   │   ├── model/        # Go structs matching Avro schemas
│   │   └── producer/     # Confluent Kafka producer + Schema Registry
│   └── config.yaml       # Default generator rates and topic names
├── schemas/              # Avro schemas (source of truth for Kafka topics)
│   ├── product.avsc
│   └── inventory_level.avsc
├── spark/
│   ├── Dockerfile        # Spark image with Iceberg + Nessie jars
│   ├── compact.py        # Compaction script (rewrite, manifests, expire)
│   └── entrypoint.sh     # Loop wrapper: runs compact.py every $COMPACTION_INTERVAL
├── kafka/
│   └── init-topics.sh    # Creates shopify.products and shopify.inventory
├── minio/
│   └── init-bucket.sh    # Creates the warehouse bucket
└── dremio/
    ├── bootstrap.sh      # Automated Nessie source registration (best-effort)
    └── setup.md          # Manual Dremio setup guide
```
