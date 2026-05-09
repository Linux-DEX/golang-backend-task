#!/bin/bash

set -e

echo "Cleaning up k8s resources..."
kubectl delete -f k8s/worker.yaml --ignore-not-found=true
kubectl delete -f k8s/api.yaml --ignore-not-found=true
kubectl delete -f k8s/redis.yaml --ignore-not-found=true
kubectl delete -f k8s/mongodb.yaml --ignore-not-found=true
kubectl delete -f k8s/storage.yaml --ignore-not-found=true
kubectl delete -f k8s/configmap.yaml --ignore-not-found=true
kubectl delete namespace edi --ignore-not-found=true

echo ""
echo "Done!"
echo ""
