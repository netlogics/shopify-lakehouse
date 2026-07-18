#!/bin/sh
set -eu

BUCKET="${WAREHOUSE_BUCKET:-warehouse}"

mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

if mc ls "local/$BUCKET" >/dev/null 2>&1; then
  echo "Bucket '$BUCKET' already exists, skipping."
else
  echo "Creating bucket '$BUCKET'..."
  mc mb "local/$BUCKET"
fi

mc ls local
exit 0
