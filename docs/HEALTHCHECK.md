# PropsDB - Health Check Usage

## Running Health Check

The health check utility can be run in the container.

### Docker Run Command

```bash
# Run health check in a running container
docker exec propsdb-api /app/healthcheck
```

## Health Check Output

### Healthy System

```
Connected to mariadb database: jam_build
Health check passed - all systems operational
{
  "status": "healthy",
  "database": "ok",
  "authorizer": "ok",
  "details": {
    "authorizer_url": "http://host.docker.internal:8080",
    "database_name": "jam_build",
    "database_type": "mariadb"
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

docker inspect propsdb-api | jq '.[0].State.Health'
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
docker exec propsdb-api /app/healthcheck | jq '.database'

# View detailed error
docker exec propsdb-api /app/healthcheck | jq '.details.database_ping_error'
```

### Authorizer Connection Issues

```bash
# Check Authorizer connectivity
docker exec propsdb-api /app/healthcheck | jq '.authorizer'

# View detailed error
docker exec propsdb-api /app/healthcheck | jq '.details.authorizer_error'
```

### Debug Mode

For more verbose output, check the container logs:

```bash
docker logs propsdb-api
```

The health check logs its results to stdout/stderr with timestamps.
