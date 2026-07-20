"""
Spark compaction script for Iceberg/Nessie tables.

Runs three maintenance procedures for each table:
  1. rewrite_data_files  — bin-pack small files into 128 MB files
  2. rewrite_manifests   — clean up manifest metadata
  3. expire_snapshots    — drop snapshots older than 1 hour, keep at least 5
"""

import os
from datetime import datetime, timedelta, timezone

from pyspark.sql import SparkSession

# ---------------------------------------------------------------------------
# Configuration (read from environment with sensible defaults)
# ---------------------------------------------------------------------------
NESSIE_URI = os.environ.get("NESSIE_URI", "http://nessie:19120/api/v1")
S3_ENDPOINT = os.environ.get("S3_ENDPOINT", "http://minio:9000")
AWS_ACCESS_KEY_ID = os.environ.get("AWS_ACCESS_KEY_ID", "")
AWS_SECRET_ACCESS_KEY = os.environ.get("AWS_SECRET_ACCESS_KEY", "")

TABLES = [
    "nessie.lakehouse.products",
    "nessie.lakehouse.product_variants",
    "nessie.lakehouse.inventory_levels",
]

# ---------------------------------------------------------------------------
# SparkSession
# ---------------------------------------------------------------------------
spark = (
    SparkSession.builder.appName("iceberg-compaction")
    # Iceberg + Nessie extensions
    .config(
        "spark.sql.extensions",
        "org.apache.iceberg.spark.extensions.IcebergSparkSessionExtensions,"
        "org.projectnessie.spark.extensions.NessieSparkSessionExtensions",
    )
    # Nessie catalog
    .config("spark.sql.catalog.nessie", "org.apache.iceberg.spark.SparkCatalog")
    .config(
        "spark.sql.catalog.nessie.catalog-impl",
        "org.apache.iceberg.nessie.NessieCatalog",
    )
    .config("spark.sql.catalog.nessie.uri", NESSIE_URI)
    .config("spark.sql.catalog.nessie.ref", "main")
    .config("spark.sql.catalog.nessie.warehouse", "s3a://warehouse/")
    # S3A / MinIO
    .config("spark.hadoop.fs.s3a.endpoint", S3_ENDPOINT)
    .config("spark.hadoop.fs.s3a.path.style.access", "true")
    .config("spark.hadoop.fs.s3a.access.key", AWS_ACCESS_KEY_ID)
    .config("spark.hadoop.fs.s3a.secret.key", AWS_SECRET_ACCESS_KEY)
    .config(
        "spark.hadoop.fs.s3a.impl",
        "org.apache.hadoop.fs.s3a.S3AFileSystem",
    )
    .getOrCreate()
)

spark.sparkContext.setLogLevel("WARN")

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def run_sql(label: str, sql: str):
    """Execute a SQL statement, print a summary, and return the result DataFrame."""
    print(f"\n[{label}] {sql[:120]}")
    df = spark.sql(sql)
    df.show(truncate=False)
    return df


def compact_table(table: str) -> None:
    print(f"\n{'=' * 60}")
    print(f"Compacting: {table}")
    print(f"{'=' * 60}")

    # 1. Rewrite data files (bin-pack strategy, 128 MB target)
    run_sql(
        "rewrite_data_files",
        f"CALL nessie.system.rewrite_data_files("
        f"table => '{table}', "
        f"strategy => 'binpack', "
        f"options => map('target-file-size-bytes', '134217728'))",
    )

    # 2. Rewrite manifests
    run_sql(
        "rewrite_manifests",
        f"CALL nessie.system.rewrite_manifests('{table}')",
    )

    # Enable GC before expiry (NessieCatalog disables it by default)
    run_sql("enable_gc", f"ALTER TABLE {table} SET TBLPROPERTIES ('gc.enabled'='true')")

    # 3. Expire snapshots older than 1 hour, retain at least 5.
    # Flink writes ~1 snapshot/min; a 30-day window accumulates thousands of
    # snapshots and produces multi-MB metadata.json files that choke query engines.
    cutoff = (datetime.now(timezone.utc) - timedelta(hours=1)).strftime(
        "%Y-%m-%d %H:%M:%S"
    )
    run_sql(
        "expire_snapshots",
        f"CALL nessie.system.expire_snapshots("
        f"table => '{table}', "
        f"older_than => TIMESTAMP '{cutoff}', "
        f"retain_last => 5)",
    )


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
if __name__ == "__main__":
    start = datetime.now(timezone.utc)
    print(f"Compaction run started at {start.isoformat()}")

    for table in TABLES:
        try:
            compact_table(table)
        except Exception as exc:  # noqa: BLE001
            print(f"ERROR compacting {table}: {exc}")

    elapsed = (datetime.now(timezone.utc) - start).total_seconds()
    print(f"\nCompaction run finished in {elapsed:.1f}s")

    spark.stop()
