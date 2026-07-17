-- Flink SQL: transform raw Shopify API events (Kafka/Avro) → Iceberg/Nessie.
--
-- Three Iceberg sinks:
--   nessie.lakehouse.products          one row per product event
--   nessie.lakehouse.product_variants  one row per variant (UNNEST)
--   nessie.lakehouse.inventory_levels  one row per inventory level event
--
-- Timestamps arrive as ISO 8601 UTC strings ("2024-01-15T10:30:00Z").
-- Parsed with: TO_TIMESTAMP(LEFT(ts, 19), 'yyyy-MM-dd''T''HH:mm:ss')
--
-- Two EXECUTE blocks → two Flink jobs.  Jobs 1+2 read the same Kafka topics
-- independently (separate consumer groups).  Avoids the shared-source +
-- UNNEST interaction inside a single STATEMENT SET which is unreliable in
-- Flink 1.20.
--
-- KNOWN RISK: CROSS JOIN UNNEST on ARRAY<ROW<...>> — if Flink cannot resolve
-- named ROW fields via v.field_name, fall back to positional access v.f0,
-- v.f1, ... matching the order in the variants DDL below.

SET 'execution.checkpointing.interval' = '60s';
SET 'execution.checkpointing.mode' = 'EXACTLY_ONCE';
SET 'execution.runtime-mode' = 'STREAMING';

-- ---------------------------------------------------------------------------
-- Kafka sources
-- ---------------------------------------------------------------------------

CREATE TABLE products_source (
  id               BIGINT,
  title            STRING,
  body_html        STRING,
  vendor           STRING,
  product_type     STRING,
  handle           STRING,
  status           STRING,
  tags             STRING,
  created_at       STRING,
  updated_at       STRING,
  published_at     STRING,
  variants         ARRAY<ROW<
    id                   BIGINT,
    product_id           BIGINT,
    title                STRING,
    price                STRING,
    sku                  STRING,
    `position`           INT,
    inventory_policy     STRING,
    compare_at_price     STRING,
    fulfillment_service  STRING,
    inventory_management STRING,
    option1              STRING,
    option2              STRING,
    option3              STRING,
    taxable              BOOLEAN,
    barcode              STRING,
    grams                INT,
    weight               DOUBLE,
    weight_unit          STRING,
    inventory_item_id    BIGINT,
    inventory_quantity   INT,
    requires_shipping    BOOLEAN,
    created_at           STRING,
    updated_at           STRING
  >>
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.products',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-products',
  'scan.startup.mode'            = 'earliest-offset',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

CREATE TABLE products_source_variants (
  id               BIGINT,
  variants         ARRAY<ROW<
    id                   BIGINT,
    product_id           BIGINT,
    title                STRING,
    price                STRING,
    sku                  STRING,
    `position`           INT,
    inventory_policy     STRING,
    compare_at_price     STRING,
    fulfillment_service  STRING,
    inventory_management STRING,
    option1              STRING,
    option2              STRING,
    option3              STRING,
    taxable              BOOLEAN,
    barcode              STRING,
    grams                INT,
    weight               DOUBLE,
    weight_unit          STRING,
    inventory_item_id    BIGINT,
    inventory_quantity   INT,
    requires_shipping    BOOLEAN,
    created_at           STRING,
    updated_at           STRING
  >>
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.products',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-variants',
  'scan.startup.mode'            = 'earliest-offset',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

CREATE TABLE inventory_source (
  inventory_item_id  BIGINT,
  location_id        BIGINT,
  available          INT,
  updated_at         STRING
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.inventory',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-inventory',
  'scan.startup.mode'            = 'earliest-offset',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

-- ---------------------------------------------------------------------------
-- Nessie/Iceberg catalog
-- ---------------------------------------------------------------------------

CREATE CATALOG nessie WITH (
  'type'         = 'iceberg',
  'catalog-impl' = 'org.apache.iceberg.nessie.NessieCatalog',
  'uri'          = 'http://nessie:19120/api/v1',
  'ref'          = 'main',
  'warehouse'    = 's3a://warehouse/',
  'io-impl'      = 'org.apache.iceberg.hadoop.HadoopFileIO'
);

CREATE DATABASE IF NOT EXISTS nessie.lakehouse;

-- ---------------------------------------------------------------------------
-- Iceberg sink tables
-- DROP first so schema changes apply on re-run.
-- ---------------------------------------------------------------------------

DROP TABLE IF EXISTS nessie.lakehouse.products;
DROP TABLE IF EXISTS nessie.lakehouse.product_variants;
DROP TABLE IF EXISTS nessie.lakehouse.inventory_levels;

CREATE TABLE nessie.lakehouse.products (
  id            BIGINT,
  title         STRING,
  body_html     STRING,
  vendor        STRING,
  product_type  STRING,
  handle        STRING,
  status        STRING,
  tags          STRING,
  created_at    TIMESTAMP(3),
  updated_at    TIMESTAMP(3),
  published_at  TIMESTAMP(3)
) PARTITIONED BY (status) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

CREATE TABLE nessie.lakehouse.product_variants (
  product_id           BIGINT,
  variant_id           BIGINT,
  title                STRING,
  price                STRING,
  sku                  STRING,
  `position`           INT,
  inventory_policy     STRING,
  compare_at_price     STRING,
  fulfillment_service  STRING,
  inventory_management STRING,
  option1              STRING,
  option2              STRING,
  option3              STRING,
  taxable              BOOLEAN,
  barcode              STRING,
  grams                INT,
  weight               DOUBLE,
  weight_unit          STRING,
  inventory_item_id    BIGINT,
  inventory_quantity   INT,
  requires_shipping    BOOLEAN,
  created_at           TIMESTAMP(3),
  updated_at           TIMESTAMP(3)
) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

CREATE TABLE nessie.lakehouse.inventory_levels (
  inventory_item_id  BIGINT,
  location_id        BIGINT,
  available          INT,
  updated_at         TIMESTAMP(3)
) PARTITIONED BY (location_id) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

-- ---------------------------------------------------------------------------
-- Job 1: products + inventory_levels
-- ---------------------------------------------------------------------------

EXECUTE STATEMENT SET
BEGIN

  INSERT INTO nessie.lakehouse.products
  SELECT
    id,
    title,
    body_html,
    vendor,
    product_type,
    handle,
    status,
    tags,
    TO_TIMESTAMP(LEFT(created_at,  19), 'yyyy-MM-dd''T''HH:mm:ss'),
    TO_TIMESTAMP(LEFT(updated_at,  19), 'yyyy-MM-dd''T''HH:mm:ss'),
    CASE WHEN published_at IS NULL THEN NULL
         ELSE TO_TIMESTAMP(LEFT(published_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
    END
  FROM products_source;

  INSERT INTO nessie.lakehouse.inventory_levels
  SELECT
    inventory_item_id,
    location_id,
    available,
    TO_TIMESTAMP(LEFT(updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
  FROM inventory_source;

END;

-- ---------------------------------------------------------------------------
-- Job 2: product_variants via UNNEST
-- Separate job so UNNEST does not interact with the shared source above.
-- ---------------------------------------------------------------------------

INSERT INTO nessie.lakehouse.product_variants
SELECT
  p.id                    AS product_id,
  v.id                    AS variant_id,
  v.title                 AS title,
  v.price                 AS price,
  v.sku                   AS sku,
  v.`position`            AS `position`,
  v.inventory_policy      AS inventory_policy,
  v.compare_at_price      AS compare_at_price,
  v.fulfillment_service   AS fulfillment_service,
  v.inventory_management  AS inventory_management,
  v.option1               AS option1,
  v.option2               AS option2,
  v.option3               AS option3,
  v.taxable               AS taxable,
  v.barcode               AS barcode,
  v.grams                 AS grams,
  v.weight                AS weight,
  v.weight_unit           AS weight_unit,
  v.inventory_item_id     AS inventory_item_id,
  v.inventory_quantity    AS inventory_quantity,
  v.requires_shipping     AS requires_shipping,
  TO_TIMESTAMP(LEFT(v.created_at, 19), 'yyyy-MM-dd''T''HH:mm:ss'),
  TO_TIMESTAMP(LEFT(v.updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
FROM products_source_variants AS p
CROSS JOIN UNNEST(p.variants) AS v;
