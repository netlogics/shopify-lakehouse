#!/bin/sh
set -e

COMPACTION_INTERVAL=${COMPACTION_INTERVAL:-600}

echo "Spark compaction loop started (interval=${COMPACTION_INTERVAL}s)"

while true; do
    echo "$(date -u '+%Y-%m-%dT%H:%M:%SZ') Starting compaction run"
    /opt/spark/bin/spark-submit /opt/spark/work-dir/compact.py
    echo "$(date -u '+%Y-%m-%dT%H:%M:%SZ') Compaction done. Sleeping ${COMPACTION_INTERVAL}s"
    sleep "$COMPACTION_INTERVAL"
done
