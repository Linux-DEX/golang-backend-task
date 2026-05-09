# Kubernetes Setup - Quick Reference

## ✅ What's Included

The `k8s/` directory contains a **basic Kubernetes deployment** suitable for a 2-year experience engineer:

```
k8s/
├── namespace.yaml      # Creates 'edi' namespace
├── configmap.yaml      # Environment configuration
├── storage.yaml        # PersistentVolumeClaims (5Gi MongoDB, 1Gi Redis)
├── mongodb.yaml        # MongoDB deployment + service
├── redis.yaml          # Redis deployment + service
├── api.yaml            # API deployment (2 replicas) + service
├── worker.yaml         # Worker deployment (2 replicas) + service
├── deploy.sh           # One-command deployment script
├── cleanup.sh          # Cleanup script
└── README.md           # Detailed documentation
```

## 🚀 Quick Commands

### Deploy
```bash
./k8s/deploy.sh
```

### Access API
```bash
kubectl port-forward -n edi svc/edi-api 8080:8080
curl http://localhost:8080/health
```

### View Status
```bash
kubectl get pods -n edi
kubectl get all -n edi
```

### View Logs
```bash
kubectl logs -n edi -l app=edi-api -f
kubectl logs -n edi -l app=edi-worker -f
```

### Scale
```bash
kubectl scale deployment edi-api -n edi --replicas=3
kubectl scale deployment edi-worker -n edi --replicas=5
```

### Cleanup
```bash
./k8s/cleanup.sh
```

## 📝 Key Features

### Resource Limits
- **API/Worker**: 128Mi-256Mi RAM, 100m-200m CPU
- **MongoDB**: 256Mi-512Mi RAM, 250m-500m CPU  
- **Redis**: 128Mi-256Mi RAM, 100m-200m CPU

### High Availability
- API: 2 replicas (can handle pod failures)
- Worker: 2 replicas (parallel job processing)
- Databases: 1 replica (sufficient for dev/testing)

### Health Checks
- API has liveness and readiness probes on `/health`
- Automatic restart if unhealthy

### Persistent Storage
- MongoDB: 5Gi persistent volume
- Redis: 1Gi persistent volume
- Data survives pod restarts

## 🎯 Supported Platforms

✅ **Minikube** - Auto-detected and images loaded automatically
✅ **Kind** - Auto-detected and images loaded automatically  
✅ **Docker Desktop** - Works with local images
✅ **Any Kubernetes cluster** - Modify `imagePullPolicy` if using a registry

## 💡 Tips

1. **First time?** Start with Docker Compose first (`docker-compose up -d`)
2. **Learning K8s?** This is perfect for understanding basic concepts
3. **Production?** See [k8s/README.md](README.md#production-considerations) for enhancements
4. **Debugging?** Use `kubectl describe pod <pod-name> -n edi` for details

## 📚 Learn More

See [k8s/README.md](README.md) for:
- Manual deployment steps
- Troubleshooting guide
- Testing procedures
- Scaling strategies
- Production considerations

## 🔄 Differences from Docker Compose

| Aspect | Docker Compose | Kubernetes |
|--------|----------------|------------|
| **Deployment** | Single host | Multi-node cluster |
| **Scaling** | Manual restart | Live scaling |
| **High Availability** | Limited | Built-in |
| **Storage** | Docker volumes | PersistentVolumes |
| **Networking** | Bridge network | Service discovery |
| **Health Checks** | Basic | Advanced probes |
| **Use Case** | Local dev | Dev + Production |

## ❓ Common Questions

**Q: Do I need Kubernetes for this project?**  
A: No! Docker Compose is perfect for local development. K8s is for learning or production-like deployments.

**Q: Which should I learn first?**  
A: Start with Docker Compose, then move to Kubernetes when comfortable.

**Q: Will this work on cloud platforms (AWS/GCP/Azure)?**  
A: Yes! Just push images to a registry and change `imagePullPolicy: Never` to `imagePullPolicy: Always`.

**Q: How do I persist data across restarts?**  
A: PersistentVolumeClaims are already configured. Data survives pod restarts automatically.

**Q: Can I use this in production?**  
A: This is a basic setup. For production, add Ingress, Secrets, monitoring, backups, and multi-replica databases. See the README for details.
