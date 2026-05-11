#!/bin/bash

set -e

echo "Deploying to k8s..."

# Check if kubectl can connect to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "No Kubernetes cluster found. Setting up kind cluster..."
    
    # Check if kind is installed
    if ! command -v kind &> /dev/null; then
        echo "Installing kind..."
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
    fi
    
    # Create kind cluster if it doesn't exist
    if ! kind get clusters 2>/dev/null | grep -q "kind"; then
        echo "Creating kind cluster..."
        kind create cluster --name kind
    fi
    
    echo "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=120s
fi

docker build -t edi-api:latest -f Dockerfile .
docker build -t edi-worker:latest -f Dockerfile.worker .
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    echo "Loading images to minikube..."
    minikube image load edi-api:latest
    minikube image load edi-worker:latest
elif command -v kind &> /dev/null && kind get clusters 2>/dev/null | grep -q .; then
    echo "Loading images to kind..."
    CLUSTER_NAME=$(kind get clusters | head -n 1)
    kind load docker-image edi-api:latest --name "$CLUSTER_NAME"
    kind load docker-image edi-worker:latest --name "$CLUSTER_NAME"
else
    echo "Using local k8s"
fi
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/mongodb.yaml
kubectl apply -f k8s/redis.yaml
kubectl wait --for=condition=available --timeout=120s deployment/mongodb -n edi
kubectl wait --for=condition=available --timeout=120s deployment/redis -n edi
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/worker.yaml
kubectl wait --for=condition=available --timeout=120s deployment/edi-api -n edi
kubectl wait --for=condition=available --timeout=120s deployment/edi-worker -n edi

echo ""
echo "Done!"
echo ""
kubectl get pods -n edi
echo ""
echo "Access the API:"
echo "   kubectl port-forward -n edi svc/edi-api 8080:8080"
echo ""
echo "View logs:"
echo "   kubectl logs -n edi -l app=edi-api -f"
echo "   kubectl logs -n edi -l app=edi-worker -f"
echo ""

