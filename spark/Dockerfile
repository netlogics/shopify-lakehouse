ARG SPARK_VERSION=3.5.4
FROM apache/spark:${SPARK_VERSION}

ARG SPARK_VERSION=3.5.4

USER root

# Download Iceberg and supporting jars to the Spark jars directory
RUN curl -fsSL -o /opt/spark/jars/iceberg-spark-runtime-3.5_2.12-1.9.1.jar \
      https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-spark-runtime-3.5_2.12/1.9.1/iceberg-spark-runtime-3.5_2.12-1.9.1.jar && \
    curl -fsSL -o /opt/spark/jars/nessie-spark-extensions-3.5_2.12-0.103.3.jar \
      https://repo1.maven.org/maven2/org/projectnessie/nessie-integrations/nessie-spark-extensions-3.5_2.12/0.103.3/nessie-spark-extensions-3.5_2.12-0.103.3.jar && \
    curl -fsSL -o /opt/spark/jars/hadoop-aws-3.3.4.jar \
      https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.4/hadoop-aws-3.3.4.jar && \
    curl -fsSL -o /opt/spark/jars/aws-java-sdk-bundle-1.12.780.jar \
      https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.780/aws-java-sdk-bundle-1.12.780.jar

# Install Python dependencies
RUN pip install --no-cache-dir boto3

WORKDIR /opt/spark/work-dir

COPY compact.py /opt/spark/work-dir/compact.py
COPY entrypoint.sh /opt/spark/work-dir/entrypoint.sh
RUN chmod +x /opt/spark/work-dir/entrypoint.sh

USER spark
