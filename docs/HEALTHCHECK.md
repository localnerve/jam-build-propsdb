# PropsDB - Health Check Usage

## Running Health Check

The health check utility can be run in several ways:

### 1. Standalone Binary

```bash
# Build the healthcheck binary
go build -o healthcheck ./cmd/healthcheck

# Run with environment variables
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_APP_DATABASE=jam_build
export DB_APP_USER=jbadmin
export DB_APP_PASSWORD=password
export AUTHZ_URL=http://localhost:8080
export AUTHZ_CLIENT_ID=your_client_id

./healthcheck
```

### 2. Docker Run Command

```bash
# Run health check in a running container
docker exec propsdb /app/healthcheck

# Run health check as a one-off command
docker run --rm \
  -e DB_TYPE=mysql \
  -e DB_HOST=mariadb \
  -e DB_PORT=3306 \
  -e DB_APP_DATABASE=jam_build \
  -e DB_APP_USER=jbadmin \
  -e DB_APP_PASSWORD=password \
  -e AUTHZ_URL=http://authorizer:8080 \
  -e AUTHZ_CLIENT_ID=your_client_id \
  --network propsdb-network \
  propsdb:latest \
  /app/healthcheck
```

### 3. Docker Compose

```bash
# Check health of running service
docker-compose exec propsdb /app/healthcheck

# Or check from host
docker exec propsdb-propsdb-1 /app/healthcheck
```

## Health Check Output

### Healthy System

```json
{
  "status": "healthy",
  "database": "ok",
  "authorizer": "ok",
  "details": {
    "authorizer_url": "http://localhost:8080",
    "database_name": "jam_build",
    "database_type": "mysql"
  }
}
```

Exit code: `0`

### Unhealthy System

```json
{
  "status": "unhealthy",
  "database": "unreachable",
  "authorizer": "ok",
  "details": {
    "authorizer_url": "http://localhost:8080",
    "database_ping_error": "dial tcp: connection refused"
  },
  "error": "Database ping failed: dial tcp: connection refused"
}
```

Exit code: `1`

## Automated Health Checks

### Docker Healthcheck

The Dockerfile includes an automated healthcheck that runs every 30 seconds:

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/app/healthcheck"]
```

View health status:
```bash
docker ps
# Look for (healthy) or (unhealthy) in STATUS column

docker inspect propsdb | jq '.[0].State.Health'
```

### Kubernetes Liveness/Readiness Probes

```yaml
livenessProbe:
  exec:
    command:
    - /app/healthcheck
  initialDelaySeconds: 5
  periodSeconds: 30
  timeoutSeconds: 10
  failureThreshold: 3

readinessProbe:
  exec:
    command:
    - /app/healthcheck
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

## Monitoring Integration

### Prometheus

Export health check metrics by parsing the JSON output:

```bash
#!/bin/bash
HEALTH_OUTPUT=$(/app/healthcheck)
STATUS=$(echo $HEALTH_OUTPUT | jq -r '.status')

if [ "$STATUS" == "healthy" ]; then
  echo "propsdb_health 1"
else
  echo "propsdb_health 0"
fi
```

### Nagios/Icinga

```bash
#!/bin/bash
/app/healthcheck > /dev/null 2>&1
exit $?
```

## Troubleshooting

### Database Connection Issues

```bash
# Check database connectivity
docker exec propsdb /app/healthcheck | jq '.database'

# View detailed error
docker exec propsdb /app/healthcheck | jq '.details.database_ping_error'
```

### Authorizer Connection Issues

```bash
# Check Authorizer connectivity
docker exec propsdb /app/healthcheck | jq '.authorizer'

# View detailed error
docker exec propsdb /app/healthcheck | jq '.details.authorizer_error'
```

### Debug Mode

For more verbose output, check the container logs:

```bash
docker logs propsdb
```

The health check logs its results to stdout/stderr with timestamps.
