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
#   - iceberg-aws-bundle          1.9.1  (S3FileIO, replaces hadoop-aws)
#   - iceberg-nessie              1.9.1  (NessieCatalog; unshaded, needs deps)
#   - hadoop-client-api/-runtime  3.3.4  (Iceberg's Flink catalog loader still
#     instantiates org.apache.hadoop.conf.Configuration even when the actual
#     FileIO is S3FileIO, not HadoopFileIO)

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
RUN mvn -B dependency:copy-dependencies \
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
    curl -fL -o /opt/flink/lib/iceberg-aws-bundle-${ICEBERG_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-aws-bundle/${ICEBERG_VERSION}/iceberg-aws-bundle-${ICEBERG_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/hadoop-client-api-${HADOOP_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-client-api/${HADOOP_VERSION}/hadoop-client-api-${HADOOP_VERSION}.jar"; \
    curl -fL -o /opt/flink/lib/hadoop-client-runtime-${HADOOP_VERSION}.jar \
      "https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-client-runtime/${HADOOP_VERSION}/hadoop-client-runtime-${HADOOP_VERSION}.jar"

# iceberg-nessie itself + its resolved transitive deps (nessie-client, etc.)
COPY --from=deps /build/lib/*.jar /opt/flink/lib/

USER flink
