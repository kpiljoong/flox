#!/bin/bash
set -e

CLUSTER_NAME="flox-dev"
NAMESPACE="flox-test"
GRAFANA_NODEPORT=32000

echo "Setting up local kind cluster..."

if ! kind get clusters | grep -q "$CLUSTER_NAME"; then
  kind create cluster --name "$CLUSTER_NAME"
else
  echo "Cluster $CLUSTER_NAME already exists."
fi

echo "Setting up namespace..."
kubectl create namespace "$NAMESPACE" || true

echo "Building flox docker image..."
docker build -t flox:dev-local .

echo "Loading flox image into kind..."
kind load docker-image flox:dev-local --name "$CLUSTER_NAME"

echo "Installing Loki stack..."
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install loki grafana/loki-stack \
  --namespace "$NAMESPACE" \
  --set grafana.enabled=false \
  --set promtail.enabled=false

echo "Installing Grafana separately..."
helm upgrade --install grafana grafana/grafana \
  --namespace "$NAMESPACE" \
  --set adminPassword=admin \
  --set service.type=NodePort \
  --set service.nodePort="$GRAFANA_NODEPORT" \
  --set datasources."datasources\.yaml".apiVersion=1 \
  --set datasources."datasources\.yaml".datasources[0].name="Loki" \
  --set datasources."datasources\.yaml".datasources[0].type="loki" \
  --set datasources."datasources\.yaml".datasources[0].url="http://loki:3100" \
  --set datasources."datasources\.yaml".datasources[0].access="proxy" \
  --set datasources."datasources\.yaml".datasources[0].isDefault=true

echo "Deploying flox..."
kubectl apply -n "$NAMESPACE" -f manifests/flox-configmap.yaml
kubectl apply -n "$NAMESPACE" -f manifests/flox-daemonset.yaml

echo "Deploying log-writter..."
kubectl apply -n "$NAMESPACE" -f manifests/log-writer.yaml

# kubectl port-forward svc/loki 3100:3100 -n "$NAMESPACE" &
# echo "Port forwarding Loki (3100)..."
# sleep 3

echo "Waiting for Grafana pod to be ready..."
kubectl rollout status deployment grafana -n "$NAMESPACE" --timeout=30s

echo "Port forwarding Grafana (3000)..."
kubectl port-forward svc/grafana 3000:80 -n "$NAMESPACE" &

echo "All done!"
echo "Grafana URL: http://localhost:3000 (or NodePort: $GRAFANA_NODEPORT)"
echo "Login with: admin/admin"
