# Build context is the repo root (see compose/airflow.yml) so this image
# can COPY both airflow/ and dbt/ -- cosmos needs the dbt project baked in
# alongside dbt-core/dbt-dremio, installed into the same Python env Airflow
# runs in (execution_mode="local" in airflow/dags/dbt_dag.py).
FROM apache/airflow:2.11.2-python3.11

COPY airflow/requirements.txt /requirements.txt
RUN pip install --no-cache-dir -r /requirements.txt

COPY --chown=airflow:root airflow/dags/ ${AIRFLOW_HOME}/dags/
COPY --chown=airflow:root airflow/entrypoint.sh /entrypoint.sh
COPY --chown=airflow:root dbt/ ${AIRFLOW_HOME}/dbt/

USER root
RUN chmod +x /entrypoint.sh && \
    mkdir -p ${AIRFLOW_HOME}/state/logs && \
    chown -R airflow:root ${AIRFLOW_HOME}/state
USER airflow

ENTRYPOINT ["/entrypoint.sh"]
