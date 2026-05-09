# Quick Start

So you want to get this thing running? Cool, here's how.

## Easiest way - Docker Compose

Just run this:

```bash
docker-compose up -d
```

Wait like 30 seconds for everything to start up, then test it:

```bash
curl http://localhost:8080/health
```

If that works, try uploading a file:

```bash
curl -X POST http://localhost:8080/jobs -F "file=@sample.edi"
```

You'll get back a job ID. Use that to check the status:

```bash
curl http://localhost:8080/jobs/YOUR_JOB_ID
```

And grab the results when it's done:

```bash
curl http://localhost:8080/jobs/YOUR_JOB_ID/result
```

There's also a test script if you want to run through everything automatically:

```bash
chmod +x test-api.sh
./test-api.sh
```

When you're done, shut it down:

```bash
docker-compose down
```

## Using Kubernetes

If you want to try the k8s setup:

```bash
chmod +x k8s/deploy.sh
./k8s/deploy.sh
```

Then forward the port so you can hit the API:

```bash
kubectl port-forward -n edi svc/edi-api 8080:8080
```

Now you can use the same curl commands as above. Check logs with:

```bash
kubectl logs -n edi -l app=edi-api -f
```

Clean up when done:

```bash
./k8s/cleanup.sh
```

## Running locally (for development)

Start MongoDB and Redis first:

```bash
docker run -d -p 27017:27017 --name mongodb mongo:7.0
docker run -d -p 6379:6379 --name redis redis:7.2-alpine
```

Then run the API server:

```bash
go run cmd/api/main.go
```

And in another terminal, start the worker:

```bash
go run cmd/worker/main.go
```

That's it. Use the same curl commands to test.

Stop everything with `killall main` and then clean up the containers:

```bash
docker stop mongodb redis && docker rm mongodb redis
```

## What you'll find running

- API is at http://localhost:8080
- Health check: http://localhost:8080/health
- Metrics: http://localhost:8080/metrics and http://localhost:9091/metrics (worker)
- MongoDB on 27017, Redis on 6379

## If something breaks

Check the logs:

```bash
docker-compose logs
```

Or for a specific service:

```bash
docker-compose logs worker
```

If the API won't start, maybe port 8080 is already taken:

```bash
lsof -i :8080
```

Jobs stuck? Check the queue:

```bash
docker exec -it edi-redis redis-cli LLEN edi:jobs:queue
```

Restart stuff if needed:

```bash
docker-compose restart
```

## Running tests

```bash
make test
```

Or with coverage:

```bash
make test-coverage
```

## More info

Check out README.md for the full documentation. TESTING.md has more details on testing, and FEATURES.md lists everything this thing can do.
