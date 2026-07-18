# Dremio Manual Setup Guide

Use this guide when the automated `dremio-bootstrap` container is not used or when you need to reconfigure the Nessie source manually.

## Prerequisites

The full lakehouse stack must be running before accessing Dremio:

```bash
make up
# or
docker compose up -d
```

Wait until all services are healthy. You can check with:

```bash
docker compose ps
```

## Access the Dremio UI

Open your browser and navigate to:

```
http://localhost:9047
```

## First-Run: Create Admin Account

On first launch, Dremio presents a setup screen asking you to create an administrator account.

1. Enter your desired **Username** and **Password**.
2. Fill in the required **First Name**, **Last Name**, and **Email** fields.
3. Click **Create Account**.

> Note: The defaults expected by other stack components are `admin` / `dremio123`. If you choose different credentials, update `DREMIO_ADMIN_USER` and `DREMIO_ADMIN_PASSWORD` in `.env`.

## Add the Nessie Catalog Source

After logging in, you need to register the Nessie catalog so Dremio can query Iceberg tables.

### Steps

1. In the left sidebar, click **Add Source** (the `+` icon next to "Sources").
2. In the source type list, select **Nessie**.
3. Configure the source with the following values:

**General tab**

| Field | Value |
|---|---|
| Name | `nessie` |
| Nessie endpoint URL | `http://nessie:19120/api/v1` |
| Authentication | None |

**Storage tab**

| Field | Value |
|---|---|
| AWS root path | `warehouse` |
| Credential type | AWS Access Key |
| AWS access key | `minioadmin` |
| AWS access secret | `minioadmin` |

**Connection properties** (click **Add property** for each):

| Name | Value |
|---|---|
| `fs.s3a.endpoint` | `http://minio:9000` |
| `fs.s3a.path.style.access` | `true` |
| `dremio.s3.compat` | `true` |

4. Click **Save**.
5. Dremio will begin a metadata refresh. Wait a few seconds, then the `nessie` source should appear in the left sidebar.

## Querying Iceberg Tables

Once the Nessie source is configured and the metadata refresh has completed, you can query tables using the SQL Runner (click the **SQL Runner** icon in the left sidebar).

Example queries:

```sql
-- List all tables under the lakehouse namespace
SHOW TABLES IN nessie.lakehouse;

-- Query the products table
SELECT * FROM nessie.lakehouse.products LIMIT 10;

-- Query the inventory table
SELECT * FROM nessie.lakehouse.inventory LIMIT 10;
```

## Troubleshooting

- **Nessie source shows no tables**: trigger a manual metadata refresh by right-clicking the source in the sidebar and selecting **Refresh Metadata**.
- **Cannot connect to MinIO**: verify MinIO is running (`docker compose ps minio`) and that the `fs.s3a.endpoint` property points to `http://minio:9000` (the internal Docker network name, not `localhost`).
- **Login fails after restart**: Dremio persists credentials in the `dremio-data` volume. If the volume was removed, you will be prompted to create a new admin account.
