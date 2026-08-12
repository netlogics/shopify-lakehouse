# Standalone JMX-to-Prometheus bridge for the Kafka broker's JVM metrics
# (kafka.server / kafka.network / kafka.controller / kafka.log MBeans).
# Runs jmx_exporter's httpserver mode, which connects out to the broker's
# JMX RMI port (see KAFKA_JMX_PORT/KAFKA_JMX_HOSTNAME in compose/kafka.yml)
# rather than being injected as a javaagent into Kafka's own JVM.
FROM eclipse-temurin:17-jre-jammy

ARG JMX_EXPORTER_VERSION=1.0.1

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN curl -fL -o jmx_prometheus_httpserver.jar \
    "https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_httpserver/${JMX_EXPORTER_VERSION}/jmx_prometheus_httpserver-${JMX_EXPORTER_VERSION}.jar"

COPY config.yml ./config.yml

EXPOSE 5556

ENTRYPOINT ["java", "-jar", "jmx_prometheus_httpserver.jar", "5556", "config.yml"]
