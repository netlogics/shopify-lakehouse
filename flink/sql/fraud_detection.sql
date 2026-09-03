-- Independent, real-time fraud detection over order_details.
--
-- Deliberately does NOT read the generator's is_synthetic_fraud /
-- fraud_pattern ground-truth label -- that would be a pass-through, not
-- detection. Instead this is a tumbling-window velocity/volume check: how
-- many orders (and how much quantity) a single customer racks up in a
-- 30-second window, using only fields a real fraud system would have
-- (customer_id, quantity, arrival time). Flagged events are expected to
-- correlate strongly with the generator's injected "velocity_burst"
-- episodes (which concentrate abnormal order volume onto one customer for
-- ~20s), but that correlation is validated downstream in dbt, not assumed
-- here.
--
-- Windowed on PROCTIME(), not event time / created_at. This isn't a
-- shortcut -- created_at is independently randomly backdated up to 30
-- days per row (generator/internal/gen/gen.go), so a fraud episode's
-- orders, despite arriving from the generator within ~20 real seconds,
-- carry created_at values scattered uniformly across the entire 30-day
-- range in event time. The velocity signal this query needs to see
-- (many orders from one customer in a short span) only exists in
-- processing/arrival time; an event-time TUMBLE window on created_at
-- would put those same-episode orders in effectively random, mostly
-- disjoint windows and never see the burst at all. Confirmed directly:
-- an earlier event-time version consumed 400+ records over several
-- minutes without a single window ever firing above the count/quantity
-- thresholds.
--
-- Runs as its own Flink SQL client session (own consumer group, own
-- source table definition) rather than being folded into ingest.sql's
-- STATEMENT SET, so this experimental detection logic can be iterated on
-- or resubmitted independently of the core ingest pipeline.

SET 'sql-client.execution.result-mode' = 'tableau';

CREATE TABLE order_details_fraud_source (
  event_id                     STRING,
  order_id                     BIGINT,
  id                           BIGINT,
  customer_id                  BIGINT,
  quantity                     INT,
  is_synthetic_fraud           BOOLEAN,
  fraud_pattern                STRING,
  proc_time AS PROCTIME()
) WITH (
  'connector'                    = 'kafka',
  'topic'                        = 'shopify.order_details',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id'          = 'flink-fraud-detection',
  'scan.startup.mode'            = 'latest-offset',
  'format'                       = 'avro-confluent',
  'avro-confluent.url'           = 'http://schema-registry:8081'
);

-- Temporary print sink for e6i.1 -- proves the detection query works and
-- correlates with the ground truth. e6i.2 replaces/adds a real
-- shopify.fraud_alerts Kafka topic + Iceberg sink for these results.
CREATE TABLE fraud_alerts_print (
  customer_id       BIGINT,
  window_start      TIMESTAMP(3),
  window_end        TIMESTAMP(3),
  order_count       BIGINT,
  total_quantity    BIGINT,
  synthetic_fraud_order_count BIGINT
) WITH (
  'connector' = 'print'
);

INSERT INTO fraud_alerts_print
SELECT
  customer_id,
  window_start,
  window_end,
  order_count,
  total_quantity,
  synthetic_fraud_order_count
FROM (
  SELECT
    customer_id,
    TUMBLE_START(proc_time, INTERVAL '30' SECOND) AS window_start,
    TUMBLE_END(proc_time, INTERVAL '30' SECOND) AS window_end,
    COUNT(*) AS order_count,
    SUM(quantity) AS total_quantity,
    -- Kept only to make correlation with the ground truth observable in
    -- this print-sink verification step -- not used in the flagging
    -- predicate below, and dropped once e6i.2 wires the real sink.
    SUM(CASE WHEN is_synthetic_fraud THEN 1 ELSE 0 END) AS synthetic_fraud_order_count
  FROM order_details_fraud_source
  WHERE customer_id IS NOT NULL
  GROUP BY customer_id, TUMBLE(proc_time, INTERVAL '30' SECOND)
)
-- Independent velocity/volume thresholds: >=5 orders or >=150 total
-- units from one customer in 30s is well above normal single-customer
-- traffic at the generator's default order_details rate (2/s spread
-- across the whole customer registry).
WHERE order_count >= 5 OR total_quantity >= 150;
