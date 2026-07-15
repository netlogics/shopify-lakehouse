-- Flink SQL streaming ingestion: Kafka (avro-confluent) -> Iceberg on
-- Nessie/MinIO. Run against a running JobManager via:
--   flink/bin/sql-client.sh -f flink/sql/ingest.sql
--
-- Connection values below are hardcoded to the shared docker network
-- service names / creds defined in the root .env contract (MINIO_ROOT_USER,
-- MINIO_ROOT_PASSWORD, AWS_REGION, NESSIE_URI, KAFKA_BOOTSTRAP,
-- SCHEMA_REGISTRY_URL). Update here if those keys ever change.

SET 'execution.checkpointing.interval' = '60s';
SET 'execution.checkpointing.mode' = 'EXACTLY_ONCE';
SET 'execution.runtime-mode' = 'STREAMING';

-- ---------------------------------------------------------------------
-- Kafka source tables (default in-memory catalog; the Nessie/Iceberg
-- catalog below only supports Iceberg tables, not Kafka connector tables).
-- ---------------------------------------------------------------------

CREATE TABLE products_source (
  id                BIGINT,
  title             STRING,
  vendor            STRING,
  product_type      STRING,
  tags              ARRAY<STRING>,
  variants          ARRAY<ROW<id BIGINT, sku STRING, price STRING, inventory_item_id BIGINT>>,
  created_at        TIMESTAMP(3),
  updated_at        TIMESTAMP(3)
) WITH (
  'connector' = 'kafka',
  'topic' = 'shopify.products',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id' = 'flink-ingest-products',
  'scan.startup.mode' = 'earliest-offset',
  'format' = 'avro-confluent',
  'avro-confluent.url' = 'http://schema-registry:8081'
);

CREATE TABLE inventory_source (
  inventory_item_id BIGINT,
  sku               STRING,
  product_id        BIGINT,
  location_id       BIGINT,
  available         INT,
  updated_at        TIMESTAMP(3)
) WITH (
  'connector' = 'kafka',
  'topic' = 'shopify.inventory',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id' = 'flink-ingest-inventory',
  'scan.startup.mode' = 'earliest-offset',
  'format' = 'avro-confluent',
  'avro-confluent.url' = 'http://schema-registry:8081'
);

-- ---------------------------------------------------------------------
-- Nessie/Iceberg catalog + sink tables
-- ---------------------------------------------------------------------

CREATE CATALOG nessie WITH (
  'type' = 'iceberg',
  'catalog-impl' = 'org.apache.iceberg.nessie.NessieCatalog',
  'uri' = 'http://nessie:19120/api/v1',
  'ref' = 'main',
  'warehouse' = 's3a://warehouse/',
  'io-impl' = 'org.apache.iceberg.aws.s3.S3FileIO',
  's3.endpoint' = 'http://minio:9000',
  's3.path-style-access' = 'true',
  's3.access-key-id' = 'minioadmin',
  's3.secret-access-key' = 'minioadmin',
  'client.region' = 'us-east-1'
);

CREATE DATABASE IF NOT EXISTS nessie.lakehouse;

CREATE TABLE IF NOT EXISTS nessie.lakehouse.products (
  id                BIGINT,
  title             STRING,
  vendor            STRING,
  product_type      STRING,
  tags              ARRAY<STRING>,
  variants          ARRAY<ROW<id BIGINT, sku STRING, price STRING, inventory_item_id BIGINT>>,
  created_at        TIMESTAMP(3),
  updated_at        TIMESTAMP(3)
) PARTITIONED BY (product_type) WITH (
  'format-version' = '2'
);

CREATE TABLE IF NOT EXISTS nessie.lakehouse.inventory_levels (
  inventory_item_id BIGINT,
  sku               STRING,
  product_id        BIGINT,
  location_id       BIGINT,
  available         INT,
  updated_at        TIMESTAMP(3)
) PARTITIONED BY (location_id) WITH (
  'format-version' = '2'
);

-- ---------------------------------------------------------------------
-- Continuous ingestion jobs (single Flink job, two sinks)
-- ---------------------------------------------------------------------

EXECUTE STATEMENT SET
BEGIN
  INSERT INTO nessie.lakehouse.products
  SELECT id, title, vendor, product_type, tags, variants, created_at, updated_at
  FROM default_catalog.default_database.products_source;

  INSERT INTO nessie.lakehouse.inventory_levels
  SELECT inventory_item_id, sku, product_id, location_id, available, updated_at
  FROM default_catalog.default_database.inventory_source;
END;
