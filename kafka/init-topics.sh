#!/usr/bin/env bash
set -euo pipefail

BOOTSTRAP="${KAFKA_BOOTSTRAP:-kafka:9092}"
PARTITIONS="${KAFKA_TOPIC_PARTITIONS:-3}"
REPLICATION="${KAFKA_TOPIC_REPLICATION:-1}"
RETENTION_MS="${KAFKA_TOPIC_RETENTION_MS:-604800000}"
TOPICS=("${KAFKA_TOPIC_PRODUCTS:-shopify.products}" "${KAFKA_TOPIC_INVENTORY:-shopify.inventory}")

for topic in "${TOPICS[@]}"; do
  if kafka-topics --bootstrap-server "$BOOTSTRAP" --describe --topic "$topic" >/dev/null 2>&1; then
    echo "Topic '$topic' already exists, skipping."
  else
    echo "Creating topic '$topic' (partitions=$PARTITIONS, replication=$REPLICATION, retention.ms=$RETENTION_MS)..."
    kafka-topics --bootstrap-server "$BOOTSTRAP" --create \
      --topic "$topic" \
      --partitions "$PARTITIONS" \
      --replication-factor "$REPLICATION" \
      --config "retention.ms=$RETENTION_MS"
  fi
done

echo "Topics ready:"
kafka-topics --bootstrap-server "$BOOTSTRAP" --list
exit 0
