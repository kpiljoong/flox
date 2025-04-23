# Flox: Fast, Programmable Log/Event Processor

**Flox** is a lightweight, high-performance log processor designed for cloud-native and containerized environments.
It supports **file-based** and **HTTP-based** inputs, real-time **filtering** via JSON rules, and pluggable outputs like **stdout**, **Loki**, and **Kafka**.

## Why Flox?

Most log processors are too heavy, inflexible, or complex. `Flox` is built for:

* **Performance-first** design
* **Programmable filters** (JSON rules, future WASM plugins)
* **Simple YAML pipeline configuration**
* **Cloud-native friendly**: DaemonSest, Sidecar, Prometheus metrics
* **Structured log ingestion and transformation**

## Features

* File-based input (tail Kubernetes pod logs)
* HTTP-based input (receive JSON log events)
* Filters: **drop**, **rename**, **add fields** (per event)
* Pluggable outputs:
  * Stdout
  * File
  * Loki
  * Kafka
* Prometheus metrics exposed at `:2112/metrics`
* DaemonSet-ready, Sidecar-ready
* Resume from saved file offsets
* Built-in graceful shutdown handling
* Hot new file detection
* Loki + Grafana local stack supported

## Running Flox Locally on kind

The fastest way to run `flox` locally is using `kind` and a simple setup script.

### 1. Clone and Setup

```bash
git clone https://github.com/kpiljoong/flox.git
cd flox
sh ./script/deploy-local-loki.sh
```

This will:
* Create a kind cluster
* Install Loki (via Helm)
* Install Grafana (via Helm)
* Deploy Flox as a DaemonSet
* Deploy a `log-writer` app to generate logs

### 2. Port-Forward Grafana

```bash
kubectl port-forward svc/grafana 3000:80 -n flox-test
```

Then access Grafana at http://localhost:3000
(default credentials: admin/admin)
Log data will appear under the pre-configured Loki datasource.

## Repository Structure

```text
flox/
├── cmd/                  # CLI entrypoint
├── internal/             # Metrics, offset, filters, inputs, outputs
│   ├── config/           # YAML loader
│   ├── filters/          # JSON field processors
│   ├── input/            # File and HTTP inputs
│   ├── metrics/          # Prometheus metrics
│   └── output/           # Output plugins (stdout, file, loki, kafka)
├── manifests/            # K8s manifests (DaemonSet, ConfigMap, example app)
├── scripts/              # Full local deployment script
│   ├── deploy-local-loki.sh
├── pipeline.yaml         # Default pipeline config
├── Dockerfile
├── Makefile              # Build/test/lint/dev flow
└── README.md
```

## Flox Deployment Models

Flox is primarily designed to run as a DaemonSet on Kubernetes nodes.
It can also be deployed as a sidecar if needed for special cases.

- **DaemonSet Mode (Recommended)**:
  - Runs once per node
  - Collects logs from multiple pods
  - Ideal for platform-wide or multi-tenant environments
  - Works seamlessly with Loki, Kafka, and other backends

- **Sidecar Mode (Optional)**:
  - Attached directly to a single application pod
  - Useful for special-purpose log filtering at the app level
  - Requires log file sharing between app and Flox containers

### Example: DaemonSet Deployment

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: flox
spec:
  selector:
    matchLabels:
      app: flox
  template:
    metadata:
      labels:
        app: flox
    spec:
      containers:
      - name: flox
        image: flox:dev-local
        args: ["--config", "/etc/flox/pipeline.yaml"]
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: config
          mountPath: /etc/flox
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: config
        configMap:
          name: flox-config
```

### Example: Sidecar Deploymnet

```yaml
containers:
  - name: app
    image: my-app:latest
    volumeMounts:
      - name: logs
        mountPath: /var/log
  - name: flox
    image: flox:dev-local
    args: ["--config", "/etc/flox/pipeline.yaml"]
    volumeMounts:
      - name: logs
        mountPath: /var/log
      - name: config
        mountPath: /etc/flox
volumes:
  - name: logs
    emptyDir: {}
  - name: config
    configMap:
      name: flox-config
```

## 📜 License

[![MIT](https://img.shields.io/badge/license-MIT-blue)](https://github.com/kpiljoong/flox/blob/master/LICENSE)
