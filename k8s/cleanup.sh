#!/bin/bash

set -e

echo "🧹 Cleaning up EDI Processing System from Kubernetes..."

# Delete all resources
kubectl delete -f k8s/worker.yaml --ignore-not-found=true
kubectl delete -f k8s/api.yaml --ignore-not-found=true
kubectl delete -f k8s/redis.yaml --ignore-not-found=true
kubectl delete -f k8s/mongodb.yaml --ignore-not-found=true
kubectl delete -f k8s/storage.yaml --ignore-not-found=true
kubectl delete -f k8s/configmap.yaml --ignore-not-found=true

# Delete namespace (this will delete everything inside it)
kubectl delete namespace edi --ignore-not-found=true

echo ""
echo "✅ Cleanup complete!"
echo ""
