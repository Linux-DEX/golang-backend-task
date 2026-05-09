#!/bin/bash

set -e

echo "🚀 Deploying EDI Processing System to Kubernetes..."

# Build Docker images
echo "📦 Building Docker images..."
docker build -t edi-api:latest -f Dockerfile .
docker build -t edi-worker:latest -f Dockerfile.worker .

# Check if we're using Minikube or Kind
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    echo "🔧 Detected Minikube - Loading images..."
    minikube image load edi-api:latest
    minikube image load edi-worker:latest
elif command -v kind &> /dev/null && kind get clusters 2>/dev/null | grep -q .; then
    echo "🔧 Detected Kind - Loading images..."
    CLUSTER_NAME=$(kind get clusters | head -n 1)
    kind load docker-image edi-api:latest --name "$CLUSTER_NAME"
    kind load docker-image edi-worker:latest --name "$CLUSTER_NAME"
else
    echo "ℹ️  Using local Kubernetes (Docker Desktop or similar)"
fi

# Apply Kubernetes manifests
echo "📝 Applying Kubernetes manifests..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/mongodb.yaml
kubectl apply -f k8s/redis.yaml

# Wait for databases to be ready
echo "⏳ Waiting for databases to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/mongodb -n edi
kubectl wait --for=condition=available --timeout=120s deployment/redis -n edi

# Deploy applications
echo "🚀 Deploying API and Worker..."
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/worker.yaml

# Wait for applications to be ready
echo "⏳ Waiting for applications to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/edi-api -n edi
kubectl wait --for=condition=available --timeout=120s deployment/edi-worker -n edi

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Current status:"
kubectl get pods -n edi
echo ""
echo "🔗 To access the API:"
echo "   kubectl port-forward -n edi svc/edi-api 8080:8080"
echo ""
echo "   Then visit: http://localhost:8080/health"
echo ""
echo "📝 View logs:"
echo "   kubectl logs -n edi -l app=edi-api -f"
echo "   kubectl logs -n edi -l app=edi-worker -f"
echo ""
