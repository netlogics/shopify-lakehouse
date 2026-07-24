# Ruby Kafka Consumer

A small standalone Ruby service that consumes Avro-encoded Shopify order-detail events from the `shopify.order_details` Kafka topic and stores them in SQLite using Karafka and Active Record.

## Structure

```text
ruby-kafka-consumer/
├── Gemfile
├── karafka.rb
├── run.rb
├── app/
│   ├── consumers/
│   │   ├── application_consumer.rb
│   │   └── order_details_consumer.rb
│   └── models/
│       ├── database.rb
│       └── order_detail.rb
└── db/
```

## How It Works

1. **`run.rb`** is the executable entry point. It loads `karafka.rb` and starts `KarafkaApp`.
2. **`karafka.rb`** configures Karafka to:
   - Connect to Kafka at `KAFKA_BOOTSTRAP` (default `127.0.0.1:9092`)
   - Use the client ID `order_details_consumer`
   - Route the `shopify.order_details` topic to `OrderDetailsConsumer`
3. **`order_details_consumer.rb`** decodes each Confluent Avro wire-format message and saves:
   - Kafka envelope: topic, partition, offset, key, consumed UTC timestamp
   - All Shopify order detail fields from `schemas/order_detail.avsc` (a line item on an order)
   - Avro field `id` is stored as `line_item_id` to avoid collision with the database primary key
4. **`database.rb`** connects to `db/development.sqlite3` and creates the `order_details` table if it does not exist.
5. **`order_detail.rb`** defines the Active Record model and validates required fields.

## Confluent Avro Decoding

Messages on `shopify.order_details` use the Confluent wire format:

```
byte 0    : 0x00 (magic byte)
bytes 1-4 : big-endian schema ID (uint32)
bytes 5+  : Avro binary payload
```

The consumer strips the envelope, fetches the schema from the Schema Registry
by ID (cached in memory), and decodes the Avro binary with the `avro` gem.

Set `SCHEMA_REGISTRY_URL` to point at the Schema Registry
(default: `http://127.0.0.1:8081`). Set `KAFKA_BOOTSTRAP` to point at the
Kafka broker (default: `127.0.0.1:9092`).

The schema source of truth is `schemas/order_detail.avsc` in the repository root.

## Data Integrity

The tuple `(topic, partition, offset)` has a unique database index and model
validation, preventing duplicate inserts. `order_id` and `line_item_id` store
the Shopify identifiers separately from the database primary key.

The consumer uses `create!`, so a duplicate message raises a validation or
database error rather than being silently skipped.

## Dependencies

The `Gemfile` declares:

- Karafka `~> 2.4`
- Active Record `~> 7.2`
- SQLite3 `~> 2.0`
- avro `~> 1.12` — Confluent Avro wire-format decoding

## Current Scope

The service is intentionally minimal. It has no lockfile, tests, or external
migration files. The `db/` directory is initially empty; the SQLite database
and `order_details` table are created at runtime on first boot.
