# Flink SQL image with Kafka(avro-confluent) source connectors and an
# Iceberg/Nessie sink baked in. Two stages:
#   1. `deps` resolves the iceberg-nessie module's transitive jars (nessie
#      client, jackson, etc.) via Maven, since that module isn't shaded.
#   2. Final stage is stock Flink with all jars copied into /opt/flink/lib.
#
# Versions verified against Maven Central for the Flink 1.20 / Iceberg 1.9
# line at the time of writing (2026-07-14):
#   - flink-sql-connector-kafka   3.4.0-1.20
#   - flink-sql-avro-confluent-registry 1.20.1
#   - iceberg-flink-runtime-1.20  1.9.1
#   - iceberg-flink-runtime-1.20  1.9.1  (Flink SQL connector; bundles iceberg-core)
#     NOTE: iceberg-aws-bundle NOT added — it duplicates iceberg-core and causes
#     AbstractMethodError (Jackson interface mismatch). Use HadoopFileIO instead.
#   - hadoop-aws                  3.3.4  (S3AFileSystem for HadoopFileIO)
#   - aws-java-sdk-bundle         1.12.780 (AWS SDK v1 used by hadoop-aws 3.3.4)
#   - iceberg-nessie              1.9.1  (NessieCatalog; unshaded, needs deps)
#   - hadoop-client-api/-runtime  3.3.4  (Hadoop Configuration used by NessieCatalog)
#   - core-site.xml               baked into /opt/flink/conf/ for S3A → MinIO config

ARG FLINK_VERSION=1.20.1
ARG SCALA_VERSION=2.12
ARG JAVA_VERSION=17
ARG ICEBERG_VERSION=1.9.1
ARG KAFKA_CONNECTOR_VERSION=3.4.0-1.20
ARG AVRO_CONFLUENT_VERSION=1.20.1
ARG HADOOP_VERSION=3.3.4

FROM maven:3.9-eclipse-temurin-17 AS deps
ARG ICEBERG_VERSION
WORKDIR /build
RUN cat <<EOF > pom.xml
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>local</groupId>
  <artifactId>flink-nessie-deps</artifactId>
  <version>1.0</version>
  <dependencies>
    <dependency>
      <groupId>org.apache.iceberg</groupId>
      <artifactId>iceberg-nessie</artifactId>
      <version>${ICEBERG_VERSION}</version>
    </dependency>
  </dependencies>
</project>
EOF
# Resolve iceberg-nessie's runtime transitive closure (nessie-client,
# nessie-model, jackson, immutables, protobuf, etc.) into flat jars.
# Exclude org.apache.iceberg.* — iceberg-flink-runtime already bundles them
# and having duplicates causes AbstractMethodError via Jackson class conflict.
RUN mvn -B dependency:copy-dependencies \
      -DincludeScope=runtime \
      -DexcludeGroupIds=org.apache.iceberg \
      -DoutputDirectory=/build/lib
# Download iceberg-nessie itself (excluded above by group, but needed)
RUN mvn -B dependency:copy-dependencies \
      -DincludeGroupIds=org.apache.iceberg \
      -DincludeArtifactIds=iceberg-nessie \
      -DincludeScope=runtime \
      -DoutputDirectory=/build/lib

FROM flink:${FLINK_VERSION}-scala_${SCALA_VERSION}-java${JAVA_VERSION}
ARG ICEBERG_VERSION
ARG KAFKA_CONNECTOR_VERSION
ARG AVRO_CONFLUENT_VERSION
ARG HADOOP_VERSION

USER root

RUN set -eux; \
    curl -fL -o /opt/flink/lib/flink-sql-connector-kafka-${KAFKA_CONNECTOR_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/flink/flink-sql-connector-kafka/${KAFKA_CONNECTOR_VERSION}/flink-sql-connector-kafka-${KAFKA_CONNECTOR_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/flink-sql-avro-confluent-registry-${AVRO_CONFLUENT_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/flink/flink-sql-avro-confluent-registry/${AVRO_CONFLUENT_VERSION}/flink-sql-avro-confluent-registry-${AVRO_CONFLUENT_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/iceberg-flink-runtime-1.20-${ICEBERG_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-flink-runtime-1.20/${ICEBERG_VERSION}/iceberg-flink-runtime-1.20-${ICEBERG_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/hadoop-client-api-${HADOOP_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-client-api/${HADOOP_VERSION}/hadoop-client-api-${HADOOP_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/hadoop-client-runtime-${HADOOP_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-client-runtime/${HADOOP_VERSION}/hadoop-client-runtime-${HADOOP_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/hadoop-aws-${HADOOP_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/${HADOOP_VERSION}/hadoop-aws-${HADOOP_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/aws-java-sdk-bundle-1.12.780.jar \
      "https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.780/aws-java-sdk-bundle-1.12.780.jar"; \
    curl -fL -o /opt/flink/lib/commons-logging-1.2.jar \
      "https://repo1.maven.org/maven2/commons-logging/commons-logging/1.2/commons-logging-1.2.jar"

# iceberg-nessie itself + its resolved transitive deps (nessie-client, etc.)
COPY --from=deps /build/lib/*.jar /opt/flink/lib/

# S3A → MinIO config: set HADOOP_CONF_DIR so Hadoop picks up core-site.xml
COPY core-site.xml /opt/flink/conf/core-site.xml
ENV HADOOP_CONF_DIR=/opt/flink/conf

USER flink
