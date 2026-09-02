#!/bin/sh
set -e

echo "==> dbt debug"
dbt debug

echo "==> dbt run"
dbt run

echo "==> dbt test"
# Data-quality tests (not_null, unique, relationships, accepted_values)
# across staging and mart models. Two relationships tests are configured
# at warn severity for expected eventual-consistency lag between the
# independent order_details/customers Kafka streams (see
# models/staging/_staging.yml) -- a warning there is not a failure.
dbt test

echo "==> dbt docs generate"
dbt docs generate

echo "==> publishing docs to shared volume"
mkdir -p /usr/app/docs_output
cp target/index.html target/catalog.json target/manifest.json target/run_results.json /usr/app/docs_output/
