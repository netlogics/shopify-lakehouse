#!/usr/bin/env bash
# `airflow standalone` auto-generates a random admin password on first run
# (printed once to stdout / standalone_admin_password.txt) unless an admin
# user already exists -- so this creates one explicitly, with a fixed,
# reproducible password from _AIRFLOW_WWW_USER_USERNAME/_PASSWORD (the same
# env var names Airflow's own official docker-compose.yaml uses for this
# purpose), before handing off to `airflow standalone`.
set -euo pipefail

airflow db migrate

# Pre-populate FAB's default roles/permissions here, as a single isolated
# process, before `standalone` launches its webserver+scheduler+triggerer
# as three concurrent processes sharing one SQLite file. Doing this
# one-time, several-thousand-row bootstrap for the first time only after
# those three are already racing each other for the SQLite write lock
# reproducibly stalls for 15+ minutes (observed directly: scheduler,
# triggerer, and webserver all go completely silent at the same instant).
# Once already populated, this is a fast no-op on every subsequent start.
airflow sync-perm

if ! airflow users list | grep -q "^${_AIRFLOW_WWW_USER_USERNAME}"; then
  airflow users create \
    --username "${_AIRFLOW_WWW_USER_USERNAME}" \
    --password "${_AIRFLOW_WWW_USER_PASSWORD}" \
    --firstname Admin \
    --lastname User \
    --role Admin \
    --email admin@example.com
fi

exec airflow standalone
