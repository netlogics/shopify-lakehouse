# Build context is the repository root (see compose/generator.yml), since
# this image also needs ../schemas/*.avsc alongside the generator/ module.
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY generator/go.mod generator/go.sum ./
RUN go mod download

COPY generator/ ./

RUN CGO_ENABLED=1 go build -o /out/generator ./cmd/generator

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/generator ./generator
COPY generator/config.yaml ./config.yaml
COPY schemas ./schemas

ENV GENERATOR_CONFIG_PATH=/app/config.yaml
ENV GENERATOR_SCHEMAS_DIR=/app/schemas

ENTRYPOINT ["/app/generator"]
