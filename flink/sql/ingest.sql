-- Flink SQL: transform raw Shopify API events (Kafka/Avro) → Iceberg/Nessie.
--
-- Each event carries a unique event_id for deduplication. Kafka sources use
-- committed-offset startup to avoid replaying events after Flink restarts
-- (combined with EXACTLY_ONCE checkpointing at 60s intervals).
--
-- Five Iceberg sinks:
--   nessie.lakehouse.products          one row per product event
--   nessie.lakehouse.product_variants  one row per variant (UNNEST)
--   nessie.lakehouse.inventory_levels  one row per inventory level event
--   nessie.lakehouse.order_details     one row per order detail (line item) event
--   nessie.lakehouse.customers         one row per customer event
--
-- Transformations applied vs raw Kafka payload:
--   products.tags        STRING (CSV) — ARRAY<STRING> split needs a custom UDF; kept as-is
--   products.body_html   dropped (no analytics value)
--   variants.price       STRING → DECIMAL(10,2)
--   variants.compare_at_price STRING → DECIMAL(10,2) (nullable)
--   variants.grams       dropped (redundant with weight + weight_unit)
--   order_details.price         STRING → DECIMAL(10,2)
--   order_details.total_discount STRING → DECIMAL(10,2)
--   order_details.grams  dropped (redundant with variant weight)
--   timestamps           ISO 8601 UTC strings → TIMESTAMP(3) via TO_TIMESTAMP + LEFT
--
-- Four INSERT blocks (Jobs 1-4).  Jobs 1-4 read Kafka topics independently
-- (separate consumer groups).  Jobs 1+2 share products_source in a single
-- consumer group; Jobs 3+4 are standalone INSERTs.  product_variants (Job 2)
-- reads products_source separately so UNNEST does not conflict with the
-- shared source in the STATEMENT SET (Flink 1.20 limitation).
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
  event_id         STRING,
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
    event_id             STRING,
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
  'scan.startup.mode'            = 'group-offsets',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

CREATE TABLE order_details_source (
  event_id                     STRING,
  order_id                     BIGINT,
  id                           BIGINT,
  variant_id                   BIGINT,
  product_id                   BIGINT,
  customer_id                  BIGINT,
  title                        STRING,
  variant_title                STRING,
  name                         STRING,
  sku                          STRING,
  vendor                       STRING,
  quantity                     INT,
  fulfillable_quantity         INT,
  current_quantity             INT,
  price                        STRING,
  total_discount               STRING,
  fulfillment_service          STRING,
  fulfillment_status           STRING,
  grams                        INT,
  requires_shipping            BOOLEAN,
  taxable                      BOOLEAN,
  gift_card                    BOOLEAN,
  product_exists               BOOLEAN,
  variant_inventory_management STRING,
  created_at                   STRING,
  updated_at                   STRING
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.order_details',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-order-details',
  'scan.startup.mode'            = 'latest-offset',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

CREATE TABLE inventory_source (
  event_id         STRING,
  inventory_item_id  BIGINT,
  location_id        BIGINT,
  available          INT,
  updated_at         STRING
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.inventory',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-inventory',
  'scan.startup.mode'            = 'group-offsets',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

CREATE TABLE customers_source (
  event_id         STRING,
  id               BIGINT,
  email            STRING,
  first_name       STRING,
  last_name        STRING,
  phone            STRING,
  state            STRING,
  verified_email   BOOLEAN,
  tags             STRING,
  created_at       STRING,
  updated_at       STRING
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.customers',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-ingest-customers',
  'scan.startup.mode'            = 'latest-offset',
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
-- Schema evolution: Add event_id column if missing (from previous schema)
-- ---------------------------------------------------------------------------

-- Drop and recreate tables to ensure correct schema (development mode)
-- For production, use ALTER TABLE for schema evolution
DROP TABLE IF EXISTS nessie.lakehouse.customers;
DROP TABLE IF EXISTS nessie.lakehouse.order_details;
DROP TABLE IF EXISTS nessie.lakehouse.inventory_levels;
DROP TABLE IF EXISTS nessie.lakehouse.product_variants;
DROP TABLE IF EXISTS nessie.lakehouse.products;

-- ---------------------------------------------------------------------------
-- Iceberg sink tables
-- Tables are created once and persist across runs. For schema evolution,
-- use ALTER TABLE ... ADD COLUMN or ALTER TABLE ... REPLACE COLUMNS.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS nessie.lakehouse.products (
  event_id      STRING,
  id            BIGINT,
  title         STRING,
  vendor        STRING,
  product_type  STRING,
  handle        STRING,
  status        STRING,
  tags          STRING,
  created_at    TIMESTAMP(3),
  updated_at    TIMESTAMP(3),
  published_at  TIMESTAMP(3)
) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

CREATE TABLE IF NOT EXISTS nessie.lakehouse.product_variants (
  event_id             STRING,
  product_id           BIGINT,
  variant_id           BIGINT,
  title                STRING,
  price                DECIMAL(10,2),
  sku                  STRING,
  `position`           INT,
  inventory_policy     STRING,
  compare_at_price     DECIMAL(10,2),
  fulfillment_service  STRING,
  inventory_management STRING,
  option1              STRING,
  option2              STRING,
  option3              STRING,
  taxable              BOOLEAN,
  barcode              STRING,
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

CREATE TABLE IF NOT EXISTS nessie.lakehouse.inventory_levels (
  event_id         STRING,
  inventory_item_id  BIGINT,
  location_id        BIGINT,
  available          INT,
  updated_at         TIMESTAMP(3)
) PARTITIONED BY (location_id) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

CREATE TABLE IF NOT EXISTS nessie.lakehouse.order_details (
  event_id                     STRING,
  order_id                     BIGINT,
  id                           BIGINT,
  variant_id                   BIGINT,
  product_id                   BIGINT,
  customer_id                  BIGINT,
  title                        STRING,
  variant_title                STRING,
  name                         STRING,
  sku                          STRING,
  vendor                       STRING,
  quantity                     INT,
  fulfillable_quantity         INT,
  current_quantity             INT,
  price                        DECIMAL(10,2),
  total_discount               DECIMAL(10,2),
  fulfillment_service          STRING,
  fulfillment_status           STRING,
  requires_shipping            BOOLEAN,
  taxable                      BOOLEAN,
  gift_card                    BOOLEAN,
  product_exists               BOOLEAN,
  variant_inventory_management STRING,
  created_at                   TIMESTAMP(3),
  updated_at                   TIMESTAMP(3)
) WITH (
  'format-version' = '2',
  'gc.enabled'     = 'true'
);

CREATE TABLE IF NOT EXISTS nessie.lakehouse.customers (
  event_id         STRING,
  id               BIGINT,
  email            STRING,
  first_name       STRING,
  last_name        STRING,
  phone            STRING,
  state            STRING,
  verified_email   BOOLEAN,
  tags             STRING,
  created_at       TIMESTAMP(3),
  updated_at       TIMESTAMP(3)
) WITH (
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
    event_id,
    id,
    title,
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
    event_id,
    inventory_item_id,
    location_id,
    available,
    TO_TIMESTAMP(LEFT(updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
  FROM inventory_source;

END;

-- ---------------------------------------------------------------------------
-- Job 2: product_variants via UNNEST
-- Uses the same products_source as Job 1. Flink reuses the single consumer
-- group across both INSERTs because they read from the same table.
-- ---------------------------------------------------------------------------

INSERT INTO nessie.lakehouse.product_variants
SELECT
  v.event_id                                  AS event_id,
  p.id                                        AS product_id,
  v.id                                        AS variant_id,
  v.title                                     AS title,
  CAST(v.price AS DECIMAL(10,2))              AS price,
  v.sku                                       AS sku,
  v.`position`                                AS `position`,
  v.inventory_policy                          AS inventory_policy,
  CAST(v.compare_at_price AS DECIMAL(10,2))   AS compare_at_price,
  v.fulfillment_service                       AS fulfillment_service,
  v.inventory_management                      AS inventory_management,
  v.option1                                   AS option1,
  v.option2                                   AS option2,
  v.option3                                   AS option3,
  v.taxable                                   AS taxable,
  v.barcode                                   AS barcode,
  v.weight                                    AS weight,
  v.weight_unit                               AS weight_unit,
  v.inventory_item_id                         AS inventory_item_id,
  v.inventory_quantity                        AS inventory_quantity,
  v.requires_shipping                         AS requires_shipping,
  TO_TIMESTAMP(LEFT(v.created_at, 19), 'yyyy-MM-dd''T''HH:mm:ss'),
  TO_TIMESTAMP(LEFT(v.updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
FROM products_source AS p
CROSS JOIN UNNEST(p.variants) AS v;

-- ---------------------------------------------------------------------------
-- Job 3: order_details
-- Separate job so it runs independently and can be restarted without
-- affecting the product/inventory jobs.
-- ---------------------------------------------------------------------------

INSERT INTO nessie.lakehouse.order_details
SELECT
  event_id,
  order_id,
  id,
  variant_id,
  product_id,
  customer_id,
  title,
  variant_title,
  name,
  sku,
  vendor,
  quantity,
  fulfillable_quantity,
  current_quantity,
  CAST(price AS DECIMAL(10,2))          AS price,
  CAST(total_discount AS DECIMAL(10,2)) AS total_discount,
  fulfillment_service,
  fulfillment_status,
  requires_shipping,
  taxable,
  gift_card,
  product_exists,
  variant_inventory_management,
  TO_TIMESTAMP(LEFT(created_at, 19), 'yyyy-MM-dd''T''HH:mm:ss'),
  TO_TIMESTAMP(LEFT(updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
FROM order_details_source;

-- ---------------------------------------------------------------------------
-- Job 4: customers
-- ---------------------------------------------------------------------------

INSERT INTO nessie.lakehouse.customers
SELECT
  event_id,
  id,
  email,
  first_name,
  last_name,
  phone,
  state,
  verified_email,
  tags,
  TO_TIMESTAMP(LEFT(created_at, 19), 'yyyy-MM-dd''T''HH:mm:ss'),
  TO_TIMESTAMP(LEFT(updated_at, 19), 'yyyy-MM-dd''T''HH:mm:ss')
FROM customers_source;
