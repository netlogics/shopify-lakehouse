# Ruby Kafka Consumer

A small standalone Ruby service that consumes Kafka order-detail events and stores them in SQLite using Karafka and Active Record.

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
   - Connect to Kafka at `127.0.0.1:9092`
   - Use the client ID `order_details_consumer`
   - Route the `order-details` topic to `OrderDetailsConsumer`
3. **`order_details_consumer.rb`** processes each Kafka message and saves its:
   - Topic
   - Partition
   - Offset
   - Key
   - Payload
   - UTC creation time
4. **`database.rb`** connects to `db/development.sqlite3` and creates the `order_details` table if it does not exist.
5. **`order_detail.rb`** defines the Active Record model and validates message identity.

## Data Integrity

The tuple `(topic, partition, offset)` has a unique database index and model validation. This tuple represents a Kafka message's unique position and prevents the same message from being inserted twice.

One caveat is that the consumer uses `create!`, so delivery of a duplicate message raises a validation or database error rather than being silently treated as already processed.

## Dependencies

The `Gemfile` declares:

- Karafka `~> 2.4`
- Active Record `~> 7.2`
- SQLite3 `~> 2.0`

## Current Scope

The service is intentionally minimal. It currently has no lockfile, tests, external migration files, or environment-based configuration. The `db/` directory is initially empty; the SQLite database is created at runtime.
