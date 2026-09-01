#!/bin/sh
set -e

echo "==> dbt debug"
dbt debug

echo "==> dbt run"
dbt run
