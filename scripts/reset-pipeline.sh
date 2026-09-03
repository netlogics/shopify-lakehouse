#!/usr/bin/env bash
# Resets pipeline data stores after a schema change to the generator's Kafka
# topics / Iceberg tables (e.g. adding a column to order_details). Run from
# the repo root, or anywhere -- the script cd's to its own repo root.
#
# Two modes:
#
#   ./scripts/reset-pipeline.sh
#     Default (Flink-only) reset. Safe for ADDITIVE, backward-compatible
#     schema changes -- a new nullable field with a default, added to the
#     Avro schema (schemas/*.avsc), the generator, and flink/sql/ingest.sql.
#     Verified sufficient in practice for exactly this case (see
#     shopify-lakehouse-92x.1 and .92x.2): Avro's schema evolution handles
#     old messages missing the new field by filling in the default, and
#     ingest.sql's own `DROP TABLE IF EXISTS` + `CREATE TABLE IF NOT EXISTS`
#     already resets the Iceberg catalog pointer on every resubmit. Kafka
#     topics, Iceberg/Nessie/MinIO data, and webhook-service's DB are left
#     untouched -- old and new rows simply coexist with the new field only
#     populated going forward.
#
#   ./scripts/reset-pipeline.sh --full
#     Full reset, for BREAKING schema changes (a field renamed, retyped, or
#     removed -- anything Avro can't reconcile via defaults). Deletes and
#     recreates every Kafka topic, wipes the Iceberg/Nessie/MinIO data
#     volumes, and wipes + re-migrates webhook-service's SQLite DB. This is
#     destructive to all pipeline data (not your git history or code) --
#     confirm before running against anything you care about.
#
# Either mode also works around a real, observed issue: a long-running
# Flink TaskManager JVM (rebuilt from the classloader corruption in
# shopify-lakehouse-92x.1's PR description) can develop
# "Trying to access closed classloader" errors when jobs are cancelled and
# resubmitted in place. Recreating flink-jobmanager/flink-taskmanager
# containers (not just cancelling jobs) gives a clean JVM and avoids this.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

FULL=false
for arg in "$@"; do
  case "$arg" in
    --full) FULL=true ;;
    -h|--help)
      sed -n '2,29p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "Unknown argument: $arg (use --full or --help)" >&2
      exit 1
      ;;
  esac
done

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

TOPICS=(
  "${KAFKA_TOPIC_PRODUCTS:-shopify.products}"
  "${KAFKA_TOPIC_INVENTORY:-shopify.inventory}"
  "${KAFKA_TOPIC_ORDER_DETAILS:-shopify.order_details}"
  "${KAFKA_TOPIC_CUSTOMERS:-shopify.customers}"
)

echo "==> Cancelling any running Flink jobs"
running_jobs="$(curl -s http://localhost:8082/jobs 2>/dev/null | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
for j in d.get("jobs", []):
    if j["status"] == "RUNNING":
        print(j["id"])
' || true)"
for job in $running_jobs; do
  curl -s -X PATCH "http://localhost:8082/jobs/${job}?mode=cancel" >/dev/null || true
done
[ -n "$running_jobs" ] && sleep 3

echo "==> Recreating flink-jobmanager/flink-taskmanager (clean JVM; also wipes ephemeral checkpoint state under file:///tmp/flink-checkpoints)"
docker compose up -d --force-recreate flink-jobmanager flink-taskmanager

if $FULL; then
  echo "==> [--full] Stopping the generator and webhook-service so nothing writes during the reset"
  docker compose stop generator webhook-service 2>/dev/null || true

  echo "==> [--full] Deleting Kafka topics (kafka-init will recreate them)"
  for topic in "${TOPICS[@]}"; do
    docker compose exec -T kafka kafka-topics --bootstrap-server kafka:9092 \
      --delete --topic "$topic" 2>&1 | grep -v "^$" || true
  done

  echo "==> [--full] Deleting the ruby-kafka-consumer's Kafka consumer group (its old offsets are meaningless once the topic above is gone)"
  docker compose exec -T kafka kafka-consumer-groups --bootstrap-server kafka:9092 \
    --delete --group order_details_consumer 2>&1 | grep -v "^$" || true

  echo "==> [--full] Wiping Iceberg/Nessie/MinIO data (stop containers, remove containers + named volumes)"
  docker compose stop minio nessie
  # Also remove the one-shot minio-init/nessie-init sidecars: they've
  # already exited, but Docker still holds their volume references open,
  # which makes the volume rm below fail with "volume is in use" even
  # though minio/nessie themselves were just removed.
  docker compose rm -f minio nessie minio-init nessie-init
  docker volume rm -f shopify-lakehouse_minio-data shopify-lakehouse_nessie-data

  echo "==> [--full] Wiping webhook-service's SQLite DB (stop container, remove container + named volume)"
  docker compose stop webhook-service 2>/dev/null || true
  docker compose rm -f webhook-service 2>/dev/null || true
  docker volume rm -f shopify-lakehouse_webhook-data 2>/dev/null || true

  echo "==> [--full] Recreating minio/nessie and their one-shot init containers"
  docker compose up -d minio nessie
  docker compose up -d --force-recreate minio-init nessie-init 2>/dev/null || true
fi

echo "==> Re-running kafka-init, resubmitting Flink SQL, and restarting the generator together in one command"
echo "    (must be simultaneous: customers_source/order_details_source use scan.startup.mode=latest-offset,"
echo "    so if the generator's fresh Registry starts producing before Flink's jobs finish subscribing,"
echo "    those early events are silently skipped -- not a bug, just how latest-offset consumption works)"
docker compose up -d --force-recreate kafka-init flink-sql-submit generator

if $FULL; then
  echo "==> [--full] Recreating webhook-service and re-running its Prisma migrations (schema.prisma has no auto-migrate-on-start step)"
  docker compose up -d webhook-service
  sleep 3
  docker compose exec -T webhook-service npx prisma migrate deploy || \
    echo "    (migration step failed or webhook-service isn't part of this compose profile -- check manually)"
fi

echo
echo "Done. Watch for a clean restart:"
echo "  docker compose logs -f generator flink-sql-submit"
echo
echo "If you ran the default (Flink-only) mode and later find you actually needed"
echo "a breaking-change reset, re-run with --full."
