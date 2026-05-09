# Kubernetes Deployment

This directory contains basic Kubernetes manifests for deploying the EDI Processing System.

## Prerequisites

- Kubernetes cluster (Minikube, Kind, Docker Desktop, or any other)
- kubectl installed and configured
- Docker installed

## Quick Start

### Deploy Everything

```bash
./k8s/deploy.sh
```

This script will:
1. Build Docker images for API and Worker
2. Load images into your cluster (if using Minikube or Kind)
3. Create namespace and all resources
4. Wait for everything to be ready

### Access the API

```bash
# Port-forward to access the API
kubectl port-forward -n edi svc/edi-api 8080:8080

# In another terminal, test it
curl http://localhost:8080/health
```

### View Logs

```bash
# API logs
kubectl logs -n edi -l app=edi-api -f

# Worker logs
kubectl logs -n edi -l app=edi-worker -f

# All logs
kubectl logs -n edi --all-containers -f
```

### Check Status

```bash
# View all pods
kubectl get pods -n edi

# View all resources
kubectl get all -n edi

# Describe a specific pod
kubectl describe pod <pod-name> -n edi
```

### Cleanup

```bash
./k8s/cleanup.sh
```

## Manual Deployment

If you prefer to deploy manually:

```bash
# 1. Build images
docker build -t edi-api:latest -f Dockerfile .
docker build -t edi-worker:latest -f Dockerfile.worker .

# 2. Load images (if using Minikube)
minikube image load edi-api:latest
minikube image load edi-worker:latest

# 3. Apply manifests in order
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/mongodb.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/worker.yaml
```

## Kubernetes Resources

### Namespace
- `edi` - Isolated namespace for all resources

### Storage
- `mongodb-pvc` - 5Gi persistent volume for MongoDB
- `redis-pvc` - 1Gi persistent volume for Redis

### Databases
- `mongodb` - MongoDB deployment (1 replica) and service
- `redis` - Redis deployment (1 replica) and service

### Applications
- `edi-api` - API server deployment (2 replicas) and service
- `edi-worker` - Worker deployment (2 replicas) and service

### Configuration
- `edi-config` - ConfigMap with environment variables

## Resource Limits

Each component has resource requests and limits:

**API & Worker:**
- Requests: 128Mi memory, 100m CPU
- Limits: 256Mi memory, 200m CPU

**MongoDB:**
- Requests: 256Mi memory, 250m CPU
- Limits: 512Mi memory, 500m CPU

**Redis:**
- Requests: 128Mi memory, 100m CPU
- Limits: 256Mi memory, 200m CPU

## Testing the Deployment

```bash
# 1. Port-forward the API
kubectl port-forward -n edi svc/edi-api 8080:8080 &

# 2. Test health endpoint
curl http://localhost:8080/health

# 3. Upload a sample EDI file
curl -X POST http://localhost:8080/jobs -F "file=@sample.edi"

# 4. Check job status (replace JOB_ID)
curl http://localhost:8080/jobs/<JOB_ID>

# 5. Get results
curl http://localhost:8080/jobs/<JOB_ID>/result
```

## Scaling

Scale the API or Worker:

```bash
# Scale API to 3 replicas
kubectl scale deployment edi-api -n edi --replicas=3

# Scale Worker to 5 replicas
kubectl scale deployment edi-worker -n edi --replicas=5

# Check status
kubectl get pods -n edi
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -n edi

# Describe problematic pod
kubectl describe pod <pod-name> -n edi

# Check logs
kubectl logs <pod-name> -n edi
```

### Image pull errors

If you see `ImagePullBackOff`:
- Make sure images are built: `docker images | grep edi`
- For Minikube: Run `minikube image load edi-api:latest edi-worker:latest`
- For Kind: Run `kind load docker-image edi-api:latest edi-worker:latest`

### PVC not binding

```bash
# Check PVC status
kubectl get pvc -n edi

# If pending, check storage class
kubectl get storageclass
```

For Minikube/Kind, storage should work automatically. For other clusters, you may need to configure a storage class.

### Connection refused errors

Make sure services can reach each other:

```bash
# Test from API pod to MongoDB
kubectl exec -n edi deployment/edi-api -- nc -zv mongodb 27017

# Test from API pod to Redis
kubectl exec -n edi deployment/edi-api -- nc -zv redis 6379
```

## Configuration

To modify configuration, edit `k8s/configmap.yaml` and reapply:

```bash
kubectl apply -f k8s/configmap.yaml
kubectl rollout restart deployment/edi-api -n edi
kubectl rollout restart deployment/edi-worker -n edi
```

## Production Considerations

This is a basic setup suitable for learning and development. For production:

1. **Use a registry** - Push images to a container registry instead of local images
2. **Add Ingress** - Set up proper ingress for external access
3. **Add Secrets** - Move sensitive data to Kubernetes Secrets
4. **Add persistence** - Use proper storage classes with backups
5. **Add monitoring** - Set up Prometheus and Grafana
6. **Add autoscaling** - Configure HPA (Horizontal Pod Autoscaler)
7. **Add security** - Network policies, pod security policies, RBAC
8. **Multi-replica databases** - Use StatefulSets for MongoDB and Redis

## File Structure

```
k8s/
├── namespace.yaml      # Namespace definition
├── configmap.yaml      # Configuration
├── storage.yaml        # PersistentVolumeClaims
├── mongodb.yaml        # MongoDB deployment & service
├── redis.yaml          # Redis deployment & service
├── api.yaml            # API deployment & service
├── worker.yaml         # Worker deployment & service
├── deploy.sh           # Deployment script
├── cleanup.sh          # Cleanup script
└── README.md           # This file
```
