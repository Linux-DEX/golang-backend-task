# K8s Setup for EDI Processing

So I threw together some basic k8s manifests to get this thing running in a cluster. Nothing fancy, just the basics to get MongoDB, Redis, the API, and workers all talking to each other.

## Getting it running

Just run the deploy script and you're good to go:

```bash
./k8s/deploy.sh
```

It'll build the Docker images, load them into your cluster (works with Minikube and Kind), create the namespace, and spin everything up. Give it a minute to get all the pods running.

Once it's up, you can access the API like this:

```bash
kubectl port-forward -n edi svc/edi-api 8080:8080

# then in another terminal
curl http://localhost:8080/health
```

## Checking logs and stuff

```bash
# API logs
kubectl logs -n edi -l app=edi-api -f

# Worker logs
kubectl logs -n edi -l app=edi-worker -f

# or just watch everything
kubectl logs -n edi --all-containers -f

# see what's running
kubectl get pods -n edi
kubectl get all -n edi
```

## Testing it out

```bash
# forward the port first
kubectl port-forward -n edi svc/edi-api 8080:8080 &

# health check
curl http://localhost:8080/health

# upload a file
curl -X POST http://localhost:8080/jobs -F "file=@sample.edi"

# check job status (use the ID from above)
curl http://localhost:8080/jobs/<JOB_ID>

# get results
curl http://localhost:8080/jobs/<JOB_ID>/result
```

## Cleaning up

When you're done:

```bash
./k8s/cleanup.sh
```

## If you want to do it manually

```bash
# build images
docker build -t edi-api:latest -f Dockerfile .
docker build -t edi-worker:latest -f Dockerfile.worker .

# load into minikube (if that's what you're using)
minikube image load edi-api:latest
minikube image load edi-worker:latest

# apply everything
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/mongodb.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/worker.yaml
```

## What's running

Everything's in the `edi` namespace. There's persistent volumes for MongoDB (5Gi) and Redis (1Gi), then 2 replicas of the API and 2 workers. Resource limits are set to keep things from going crazy - APIs and workers get 128-256Mi mem and 100-200m CPU, MongoDB gets a bit more at 256-512Mi mem and 250-500m CPU.

## Scaling

Need more capacity? Just scale it:

```bash
kubectl scale deployment edi-api -n edi --replicas=3
kubectl scale deployment edi-worker -n edi --replicas=5
kubectl get pods -n edi
```

## When things break

If pods aren't starting, check what's up:

```bash
kubectl get pods -n edi
kubectl describe pod <pod-name> -n edi
kubectl logs <pod-name> -n edi
```

Getting `ImagePullBackOff`? Make sure you actually built the images (`docker images | grep edi`) and loaded them into your cluster. For Minikube use `minikube image load`, for Kind use `kind load docker-image`.

PVC stuck in pending? Check your storage class with `kubectl get storageclass`. Should work fine on Minikube/Kind out of the box.

Services not connecting? Test it:

```bash
kubectl exec -n edi deployment/edi-api -- nc -zv mongodb 27017
kubectl exec -n edi deployment/edi-api -- nc -zv redis 6379
```

## Changing config

Edit `k8s/configmap.yaml` then:

```bash
kubectl apply -f k8s/configmap.yaml
kubectl rollout restart deployment/edi-api -n edi
kubectl rollout restart deployment/edi-worker -n edi
```

## Production notes

Look, this is just a dev setup. If you're actually taking this to prod you'll want to:
- Push images to a proper registry (ECR, GCR, whatever)
- Add an Ingress controller for real traffic
- Move secrets out of the configmap into actual Secrets
- Set up proper storage with backups
- Add monitoring (Prometheus/Grafana)
- Set up HPA for autoscaling
- Network policies and RBAC
- Make MongoDB and Redis properly HA with StatefulSets

But for local dev and testing? This works fine.

## Files

```
k8s/
├── namespace.yaml      # the namespace
├── configmap.yaml      # env vars and config
├── storage.yaml        # PVCs for mongo and redis
├── mongodb.yaml        # mongo deployment
├── redis.yaml          # redis deployment
├── api.yaml            # API server
├── worker.yaml         # background workers
├── deploy.sh           # one-click deploy
└── cleanup.sh          # tear it all down
```

Need to setup kubectl and docker first obviously. Works with Minikube, Kind, Docker Desktop k8s, or any cluster really.
