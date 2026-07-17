#!/bin/sh
set -e

DREMIO_HOST="${DREMIO_HOST:-dremio}"
DREMIO_PORT="${DREMIO_PORT:-9047}"
DREMIO_ADMIN_USER="${DREMIO_ADMIN_USER:-admin}"
DREMIO_ADMIN_PASSWORD="${DREMIO_ADMIN_PASSWORD:-dremio123}"
NESSIE_ENDPOINT="${NESSIE_ENDPOINT:-http://nessie:19120/api/v2}"
AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-minioadmin}"
AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-minioadmin}"
S3_ENDPOINT="${S3_ENDPOINT:-http://minio:9000}"

BASE_URL="http://${DREMIO_HOST}:${DREMIO_PORT}"

# ---------------------------------------------------------------------------
# Step 1: Wait for Dremio to be ready
# ---------------------------------------------------------------------------
printf 'Waiting for Dremio at %s ...\n' "${BASE_URL}"
i=0
while true; do
  status=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}/apiv2/server_status" || true)
  if [ "${status}" = "200" ]; then
    printf 'Dremio is ready.\n'
    break
  fi
  i=$((i + 1))
  if [ "${i}" -ge 60 ]; then
    printf 'ERROR: Dremio did not become ready after 60 attempts.\n' >&2
    exit 1
  fi
  printf 'Dremio not ready (HTTP %s), retrying in 5s ...\n' "${status}"
  sleep 5
done

# ---------------------------------------------------------------------------
# Step 2: Create first admin user (idempotent — 409 means already exists)
# ---------------------------------------------------------------------------
printf 'Creating admin user "%s" ...\n' "${DREMIO_ADMIN_USER}"
first_user_status=$(curl -s -o /dev/null -w '%{http_code}' \
  -X PUT \
  -H 'Content-Type: application/json' \
  -d "{\"userName\":\"${DREMIO_ADMIN_USER}\",\"firstName\":\"Admin\",\"lastName\":\"User\",\"email\":\"admin@localhost\",\"createdAt\":0,\"password\":\"${DREMIO_ADMIN_PASSWORD}\"}" \
  "${BASE_URL}/apiv2/bootstrap/firstuser" || true)

if [ "${first_user_status}" = "200" ] || [ "${first_user_status}" = "204" ]; then
  printf 'Admin user created.\n'
elif [ "${first_user_status}" = "400" ] || [ "${first_user_status}" = "409" ]; then
  printf 'Admin user already exists (idempotent — OK).\n'
else
  printf 'ERROR: Unexpected status %s when creating admin user.\n' "${first_user_status}" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Step 3: Authenticate and extract token
# ---------------------------------------------------------------------------
printf 'Authenticating as "%s" ...\n' "${DREMIO_ADMIN_USER}"
login_response=$(curl -s \
  -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"userName\":\"${DREMIO_ADMIN_USER}\",\"password\":\"${DREMIO_ADMIN_PASSWORD}\"}" \
  "${BASE_URL}/apiv2/login" || true)

TOKEN=$(printf '%s' "${login_response}" | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')

if [ -z "${TOKEN}" ]; then
  printf 'ERROR: Failed to extract auth token. Login response: %s\n' "${login_response}" >&2
  exit 1
fi
printf 'Authenticated successfully.\n'

# ---------------------------------------------------------------------------
# Step 4: Check whether Nessie source already exists
# ---------------------------------------------------------------------------
printf 'Checking if Nessie source already exists ...\n'
catalog_response=$(curl -s \
  -H "Authorization: _dremio${TOKEN}" \
  -H 'Content-Type: application/json' \
  "${BASE_URL}/api/v3/catalog" || true)

nessie_exists=$(printf '%s' "${catalog_response}" | grep -o '"name":"nessie"' | head -1)

if [ -n "${nessie_exists}" ]; then
  printf 'Nessie source already exists — skipping creation.\n'
  printf 'Bootstrap complete.\n'
  exit 0
fi

# ---------------------------------------------------------------------------
# Step 5: Create Nessie source
# ---------------------------------------------------------------------------
printf 'Creating Nessie source ...\n'
create_status=$(curl -s -o /tmp/create_response.txt -w '%{http_code}' \
  -X POST \
  -H "Authorization: _dremio${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d "{
    \"entityType\": \"source\",
    \"name\": \"nessie\",
    \"type\": \"NESSIE\",
    \"config\": {
      \"nessieEndpoint\": \"${NESSIE_ENDPOINT}\",
      \"nessieAuthType\": \"NONE\",
      \"awsAccessKey\": \"${AWS_ACCESS_KEY_ID}\",
      \"awsAccessSecret\": \"${AWS_SECRET_ACCESS_KEY}\",
      \"awsRootPath\": \"warehouse\",
      \"propertyList\": [
        {\"name\": \"fs.s3a.endpoint\", \"value\": \"${S3_ENDPOINT}\"},
        {\"name\": \"fs.s3a.path.style.access\", \"value\": \"true\"},
        {\"name\": \"dremio.s3.compat\", \"value\": \"true\"},
        {\"name\": \"fs.s3a.impl\", \"value\": \"org.apache.hadoop.fs.s3a.S3AFileSystem\"}
      ],
      \"credentialType\": \"ACCESS_KEY\",
      \"secure\": false
    }
  }" \
  "${BASE_URL}/api/v3/catalog" || true)

if [ "${create_status}" = "200" ] || [ "${create_status}" = "201" ]; then
  printf 'Nessie source created successfully.\n'
elif [ "${create_status}" = "409" ]; then
  printf 'Nessie source already exists (concurrent creation — OK).\n'
else
  printf 'ERROR: Unexpected status %s when creating Nessie source.\n' "${create_status}" >&2
  cat /tmp/create_response.txt >&2
  exit 1
fi

printf 'Bootstrap complete.\n'
exit 0
