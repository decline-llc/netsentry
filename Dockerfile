# syntax=docker/dockerfile:1

FROM ubuntu:24.04 AS build

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    gcc \
    git \
    golang-go \
    libpcap-dev \
    make \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY engine/go.mod engine/go.sum ./engine/
RUN cd engine && go mod download

COPY . .
RUN make build

FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /opt/netsentry

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libpcap0.8 \
    python3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/bin/netsentry-capture /usr/local/bin/netsentry-capture
COPY --from=build /src/bin/netsentry-engine /usr/local/bin/netsentry-engine
COPY configs ./configs
COPY docs ./docs
COPY README.md LICENSE CHANGELOG.md SECURITY.md ./

RUN mkdir -p data logs && chmod 0750 data logs

EXPOSE 8080
VOLUME ["/opt/netsentry/data", "/opt/netsentry/logs"]

ENTRYPOINT ["netsentry-engine"]
CMD ["-config", "configs/config.yaml"]
