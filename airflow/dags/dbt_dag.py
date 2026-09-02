"""Runs the shopify_lakehouse dbt project (see /dbt) hourly, one Airflow
task per dbt model/source rather than a single opaque `dbt build` task.
"""

from datetime import datetime

from cosmos import DbtDag, ExecutionConfig, ExecutionMode, ProfileConfig, ProjectConfig

DBT_PROJECT_PATH = "/opt/airflow/dbt"

profile_config = ProfileConfig(
    profile_name="shopify_lakehouse",
    target_name="dev",
    profiles_yml_filepath=f"{DBT_PROJECT_PATH}/profiles.yml",
)

execution_config = ExecutionConfig(
    execution_mode=ExecutionMode.LOCAL,
)

dbt_shopify_lakehouse_dag = DbtDag(
    dag_id="dbt_shopify_lakehouse",
    project_config=ProjectConfig(DBT_PROJECT_PATH),
    profile_config=profile_config,
    execution_config=execution_config,
    schedule="@hourly",
    start_date=datetime(2024, 1, 1),
    catchup=False,
    default_args={"retries": 1},
)
