# EDI Processing System

Backend service for processing EDI files asynchronously. Built with Go, uses MongoDB for storage and Redis for job queuing.

## What it does

Basically you upload an EDI file via REST API, it gets queued in Redis, a worker picks it up and processes it. If something fails, it retries up to 3 times. Results get stored in MongoDB and you can check the status anytime.

There's also a Prometheus metrics endpoint if you need monitoring.

## Tech stack

- Go 1.22
- Gin for the API
- MongoDB 7.0
- Redis 7.2
- Docker & Docker Compose

## Getting started

You've got three options here:

**Docker Compose** - easiest way, just run `docker-compose up -d` and everything works

**Kubernetes** - if you want to test it in a production-like setup, there's K8s manifests in the `k8s/` folder. Run `./k8s/deploy.sh` to deploy everything

**Local dev** - spin up MongoDB and Redis yourself, then run the API server and worker separately. Good for debugging.

### Using Docker Compose (recommended)

```/dev/null/bash.sh#L1-10
git clone <repository-url>
cd golang-backend-task
docker-compose up -d

# Check everything is running
docker-compose ps

# Test it
curl http://localhost:8080/health
```

You'll get the API on port 8080, worker on 9091, MongoDB on 27017, and Redis on 6379.

### Using Kubernetes

```/dev/null/bash.sh#L1-10
./k8s/deploy.sh

# Forward the port to access locally
kubectl port-forward -n edi svc/edi-api 8080:8080

# Check it works
curl http://localhost:8080/health

# When you're done
./k8s/cleanup.sh
```

Check `k8s/README.md` for more details on the K8s setup.

### Local development

```/dev/null/bash.sh#L1-15
# Start MongoDB and Redis
docker run -d -p 27017:27017 --name mongodb mongo:7.0
docker run -d -p 6379:6379 --name redis redis:7.2-alpine

# Install deps
go mod download

# Run API server
make run
# or just: go run cmd/api/main.go

# In another terminal, run the worker
make run-worker
# or: go run cmd/worker/main.go
```

## API usage

**Health check:**

```/dev/null/bash.sh#L1-2
GET /health
```

**Upload a file:**

```/dev/null/bash.sh#L1-3
curl -X POST http://localhost:8080/jobs \
  -F "file=@sample.edi"
```

Returns a job_id you can use to check status.

**Check job status:**

```/dev/null/bash.sh#L1-2
curl http://localhost:8080/jobs/{job_id}
```

Status will be one of: `pending`, `processing`, `completed`, `failed`

**Get the results:**

```/dev/null/bash.sh#L1-2
curl http://localhost:8080/jobs/{job_id}/result
```

This gives you the parsed claims with a summary.

## EDI file format

Pretty simple format - just pipe-separated values:

```/dev/null/edi.txt#L1-3
CLAIM*CLM001*MEM123*2500
CLAIM*CLM002*MEM456*3000
CLAIM*CLM003*MEM789*1500
```

Format is: `CLAIM*{claim_id}*{member_id}*{amount}`

## Project structure

```/dev/null/tree.txt#L1-15
golang-backend-task/
├── cmd/
│   ├── api/          # API server
│   └── worker/       # Worker service
├── internal/
│   ├── api/          # HTTP handlers
│   ├── config/       # Config stuff
│   ├── logger/       # Logging
│   ├── metrics/      # Prometheus metrics
│   ├── models/       # Data models
│   ├── parser/       # EDI parser
│   ├── queue/        # Redis queue
│   ├── storage/      # MongoDB ops
│   └── worker/       # Job processing
├── Dockerfile & Dockerfile.worker
└── docker-compose.yml
```

## Configuration

Set these via environment variables:

- `MONGODB_URI` - defaults to `mongodb://localhost:27017`
- `MONGODB_DATABASE` - defaults to `edi_processor`
- `REDIS_HOST` - defaults to `localhost`
- `REDIS_PORT` - defaults to `6379`
- `REDIS_PASSWORD` - leave empty if no password
- `REDIS_DB` - defaults to `0`
- `WORKER_MAX_RETRIES` - defaults to `3`
- `WORKER_POLL_INTERVAL` - how often worker checks queue (seconds), defaults to `1`
- `LOG_LEVEL` - defaults to `info`

## How it works

1. You POST a file to `/jobs`
2. API creates a job record in MongoDB (status: pending) and adds the job ID to Redis queue
3. Worker polls the Redis queue, grabs the job ID
4. Worker fetches job from MongoDB, updates status to processing, parses the file
5. If successful, saves results and marks as completed. If it fails, retries up to 3 times.

Pretty straightforward.

## Development stuff

Build:

```/dev/null/bash.sh#L1-3
make build          # API server
make build-worker   # Worker
```

Tests:

```/dev/null/bash.sh#L1-2
make test
```

View logs:

```/dev/null/bash.sh#L1-5
docker-compose logs -f

# Or specific service
docker-compose logs -f api
docker-compose logs -f worker
```

Stop everything:

```/dev/null/bash.sh#L1-4
docker-compose down

# Remove volumes too
docker-compose down -v
```

## Monitoring

Metrics are exposed at:

- API: http://localhost:8080/metrics
- Worker: http://localhost:9091/metrics

Main metric is `edi_jobs_total` which tracks jobs by status. Also includes standard Go runtime metrics.
