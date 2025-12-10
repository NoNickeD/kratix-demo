# Datadog Stack Promise

This Promise provides self-service Datadog observability stacks with tiered configurations.

## Promise Generation

This Promise was generated with:

```bash
kratix init helm-promise datadog-stack \
  --chart-url https://helm.datadoghq.com \
  --chart-name datadog \
  --group platform.srekubecraft.io \
  --kind DatadogStack
```

## API Properties

The following properties were added to the Promise API:

```bash
kratix update api \
  --property tier:string \
  --property environment:string \
  --property clusterName:string
```

### Tiers

| Tier | Description | Features |
|------|-------------|----------|
| `minimal` | Basic metrics only | Metrics, Cluster Agent |
| `standard` | Metrics + APM + Logs | Metrics, APM, Logs, Service Monitoring |
| `full` | All features enabled | Metrics, APM, Logs, Process, NPM, Security |

## Pipeline Container

The pipeline is built as a Go binary and pushed to GitHub Container Registry:

```bash
# Build locally
cd pipelines/datadog-configure
go build -o pipeline .

# Build Docker image
docker build -t ghcr.io/nonicked/kratix-datadog-pipeline:latest .
```

## Adding Pipeline to Promise

```bash
kratix add container resource/configure/datadog \
  --image ghcr.io/nonicked/kratix-datadog-pipeline:latest
```

## Usage

Create a DatadogStack resource:

```yaml
apiVersion: platform.srekubecraft.io/v1alpha1
kind: DatadogStack
metadata:
  name: dev-observability
  namespace: default
spec:
  tier: standard
  environment: dev
  clusterName: dev-cluster
```

Apply the Promise and resource:

```bash
# Install the Promise
kubectl apply -f promise.yaml

# Create a DatadogStack
kubectl apply -f example-resource.yaml
```

## Pipeline Structure

```
pipelines/datadog-configure/
├── main.go              # Go pipeline logic
├── go.mod               # Go module
├── Dockerfile           # Multi-stage build
└── values/
    ├── values-minimal.yaml   # Minimal tier Helm values
    ├── values-standard.yaml  # Standard tier Helm values
    └── values-full.yaml      # Full tier Helm values
```
