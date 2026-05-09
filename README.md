# EDI Processing System

A backend system built with Go for processing EDI (Electronic Data Interchange) files asynchronously. The system uses MongoDB for data storage, Redis for job queuing, and provides a REST API for file uploads and status tracking.

## Features

- **REST API** for uploading EDI files and checking processing status
- **Asynchronous Processing** using Redis queue
- **Automatic Retries** for failed jobs (up to 3 attempts)
- **MongoDB Storage** for persistence
- **Docker Compose** for easy local deployment
- **Kubernetes** manifests for production-like deployment
- **Prometheus Metrics** endpoint for monitoring
- **Graceful Shutdown** handling

## Deployment Options

Choose the deployment method that fits your needs:

| Method             | Best For                            | Complexity | Commands                      |
| ------------------ | ----------------------------------- | ---------- | ----------------------------- |
| **Docker Compose** | Local development, quick testing    | Low        | `docker-compose up -d`        |
| **Kubernetes**     | Learning K8s, production-like setup | Medium     | `./k8s/deploy.sh`             |
| **Local**          | Development, debugging              | Low        | `make run && make run-worker` |

## Tech Stack

- **Language**: Go 1.22
- **API Framework**: Gin
- **Database**: MongoDB 7.0
- **Queue**: Redis 7.2
- **Containerization**: Docker & Docker Compose

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Client    │─────>│  API Server │─────>│    Redis    │
│             │      │  (Port 8080)│      │   Queue     │
└─────────────┘      └─────────────┘      └──────┬──────┘
                                                  │
                                                  v
                                          ┌─────────────┐
                                          │   Worker    │
                                          │  (Port 9091)│
                                          └──────┬──────┘
                                                  │
                                                  v
                                          ┌─────────────┐
                                          │  MongoDB    │
                                          │             │
                                          └─────────────┘
```

## Project Structure

```
golang-backend-task/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Worker service entry point
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── config/       # Configuration management
│   ├── logger/       # Logging utilities
│   ├── metrics/      # Prometheus metrics
│   ├── models/       # Data models
│   ├── parser/       # EDI file parser
│   ├── queue/        # Redis queue operations
│   ├── storage/      # MongoDB operations
│   └── worker/       # Job processor
├── Dockerfile        # API server Docker image
├── Dockerfile.worker # Worker Docker image
├── docker-compose.yml
├── Makefile
└── README.md
```

## Quick Start

### Prerequisites

- **For Docker Compose**: Docker and Docker Compose
- **For Kubernetes**: kubectl, Minikube/Kind/Docker Desktop, Docker
- **For local development**: Go 1.22+, MongoDB 7.0+, Redis 7.2+

### Option 1: Docker Compose (Easiest)

1. Clone the repository:

```bash
git clone <repository-url>
cd golang-backend-task
```

2. Start all services:

```bash
docker-compose up -d
```

3. Check that services are running:

```bash
docker-compose ps
```

You should see:

- `edi-api` on port 8080
- `edi-worker` on port 9091
- `edi-mongodb` on port 27017
- `edi-redis` on port 6379

4. Test the API:

```bash
curl http://localhost:8080/health
```

### Option 2: Kubernetes (Production-like)

1. Deploy to Kubernetes:

```bash
./k8s/deploy.sh
```

2. Access the API:

```bash
# Port-forward to access locally
kubectl port-forward -n edi svc/edi-api 8080:8080

# Test it
curl http://localhost:8080/health
```

3. View logs:

```bash
kubectl logs -n edi -l app=edi-api -f
```

4. Cleanup when done:

```bash
./k8s/cleanup.sh
```

See [k8s/README.md](k8s/README.md) for detailed Kubernetes documentation.

### Option 3: Local Development

1. Start MongoDB and Redis:

```bash
docker run -d -p 27017:27017 --name mongodb mongo:7.0
docker run -d -p 6379:6379 --name redis redis:7.2-alpine
```

2. Install dependencies:

```bash
go mod download
```

3. Run the API server:

```bash
make run
# or
go run cmd/api/main.go
```

4. In another terminal, run the worker:

```bash
make run-worker
# or
go run cmd/worker/main.go
```

## API Endpoints

### Health Check

```bash
GET /health
```

Response:

```json
{
  "status": "healthy"
}
```

### Upload EDI File

```bash
POST /jobs
Content-Type: multipart/form-data

# Example
curl -X POST http://localhost:8080/jobs \
  -F "file=@sample.edi"
```

Response:

```json
{
  "job_id": "65a5b1234567890abcdef123",
  "message": "Job created successfully and queued for processing"
}
```

### Get Job Status

```bash
GET /jobs/:job_id

# Example
curl http://localhost:8080/jobs/65a5b1234567890abcdef123
```

Response:

```json
{
  "job_id": "65a5b1234567890abcdef123",
  "status": "completed",
  "retry_count": 0,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:10Z"
}
```

Status values: `pending`, `processing`, `completed`, `failed`

### Get Processing Result

```bash
GET /jobs/:job_id/result

# Example
curl http://localhost:8080/jobs/65a5b1234567890abcdef123/result
```

Response:

```json
{
  "status": "completed",
  "claims": [
    {
      "claim_id": "CLM001",
      "member_id": "MEM123",
      "amount": 2500
    },
    {
      "claim_id": "CLM002",
      "member_id": "MEM456",
      "amount": 3000
    }
  ],
  "summary": {
    "total_claims": 2,
    "total_amount": 5500
  }
}
```

## EDI File Format

The system expects EDI files with the following format:

```
CLAIM*CLM001*MEM123*2500
CLAIM*CLM002*MEM456*3000
CLAIM*CLM003*MEM789*1500
```

Each line represents a claim with pipe-separated values:

- `CLAIM`: Record type (must be "CLAIM")
- `claim_id`: Unique claim identifier
- `member_id`: Member identifier
- `amount`: Claim amount (numeric)

## Configuration

Configuration is managed through environment variables:

| Variable               | Description                   | Default                     |
| ---------------------- | ----------------------------- | --------------------------- |
| `MONGODB_URI`          | MongoDB connection string     | `mongodb://localhost:27017` |
| `MONGODB_DATABASE`     | Database name                 | `edi_processor`             |
| `REDIS_HOST`           | Redis host                    | `localhost`                 |
| `REDIS_PORT`           | Redis port                    | `6379`                      |
| `REDIS_PASSWORD`       | Redis password                | ``                          |
| `REDIS_DB`             | Redis database number         | `0`                         |
| `WORKER_MAX_RETRIES`   | Maximum retry attempts        | `3`                         |
| `WORKER_POLL_INTERVAL` | Queue poll interval (seconds) | `1`                         |
| `LOG_LEVEL`            | Logging level                 | `info`                      |

## Development

### Build

```bash
# Build API server
make build

# Build worker
make build-worker
```

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### View Logs

```bash
# Docker logs
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f worker
```

### Stop Services

```bash
docker-compose down

# Remove volumes too
docker-compose down -v
```

## Monitoring

The system exposes Prometheus metrics:

- **API Metrics**: http://localhost:8080/metrics
- **Worker Metrics**: http://localhost:9091/metrics

Available metrics:

- `edi_jobs_total` - Total number of jobs by status
- Standard Go runtime metrics

## How It Works

1. **Job Creation**
   - User uploads an EDI file via POST /jobs
   - API server saves the job to MongoDB with status "pending"
   - Job ID is added to Redis queue
   - Job ID is returned to the user

2. **Job Processing**
   - Worker polls Redis queue for job IDs
   - Worker retrieves job details from MongoDB
   - Worker updates status to "processing"
   - Worker parses the EDI file
   - Worker saves results and updates status to "completed" or "failed"

3. **Retry Logic**
   - If parsing fails, job is marked as "pending" again
   - Retry count is incremented
   - Job is re-queued for processing
   - After 3 failed attempts, job is marked as "failed"

## Troubleshooting

### API not starting

- Check if port 8080 is available
- Verify MongoDB and Redis are running
- Check logs: `docker-compose logs api`

### Worker not processing jobs

- Check Redis connection: `docker exec -it edi-redis redis-cli PING`
- Check MongoDB connection: `docker exec -it edi-mongodb mongosh --eval "db.adminCommand('ping')"`
- Verify queue has jobs: `docker exec -it edi-redis redis-cli LLEN edi:jobs:queue`
- Check worker logs: `docker-compose logs worker`

### Jobs stuck in processing

- Restart worker: `docker-compose restart worker`
- Check for parsing errors in worker logs

## Future Improvements

- Add authentication and authorization
- Implement rate limiting
- Add job expiration/cleanup
- Support multiple EDI file formats
- Add webhooks for job completion notifications
- Implement horizontal scaling for workers
