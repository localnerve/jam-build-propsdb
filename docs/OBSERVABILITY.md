# PropsDB - Observability Guide

## Overview

PropsDB includes comprehensive observability with Prometheus metrics and Grafana dashboards for monitoring application performance, database health, and API usage.

## Components

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics` endpoint:

- **HTTP Metrics**: Request count, duration, status codes
- **Custom Metrics**: Database operations, cache hits/misses
- **Go Runtime Metrics**: Goroutines, memory usage, GC stats

### Grafana Dashboards

Pre-configured dashboards for visualizing:
- API request rates and latencies
- Error rates by endpoint
- Database connection pool stats
- System resource usage

## Quick Start

### Using Docker Compose

```bash
# Start all services including observability stack
docker-compose -f docker-compose.yml -f docker-compose.observability.yml up -d

# Access services
# PropsDB API: http://localhost:3000
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3001 (admin/admin)
```

### Standalone Setup

1. **Start PropsDB**:
```bash
make run
```

2. **View Metrics**:
```bash
curl http://localhost:3000/metrics
```

## Prometheus Configuration

The Prometheus configuration (`monitoring/prometheus.yml`) scrapes metrics from PropsDB every 5 seconds:

```yaml
scrape_configs:
  - job_name: 'propsdb'
    static_configs:
      - targets: ['propsdb:3000']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

## Grafana Setup

### Initial Login

- URL: http://localhost:3001
- Username: `admin`
- Password: `admin` (change on first login)

### Data Source

Prometheus is automatically configured as the default data source.

### Creating Dashboards

1. Navigate to Dashboards → New Dashboard
2. Add panels with PromQL queries
3. Save dashboard to `/monitoring/grafana/dashboards/`

### Example Queries

**Request Rate**:
```promql
rate(propsdb_http_requests_total[5m])
```

**Response Time (95th percentile)**:
```promql
histogram_quantile(0.95, rate(propsdb_http_request_duration_seconds_bucket[5m]))
```

**Error Rate**:
```promql
rate(propsdb_http_requests_total{status=~"5.."}[5m])
```

## Available Metrics

### HTTP Metrics

- `propsdb_http_requests_total` - Total HTTP requests
- `propsdb_http_request_duration_seconds` - Request duration histogram
- `propsdb_http_requests_in_progress` - Current in-flight requests

### Application Metrics

- `propsdb_database_connections` - Database connection pool stats
- `propsdb_cache_hits_total` - Cache hit count
- `propsdb_cache_misses_total` - Cache miss count

### Go Runtime Metrics

- `go_goroutines` - Number of goroutines
- `go_memstats_alloc_bytes` - Allocated memory
- `go_gc_duration_seconds` - GC pause duration

## Alerting

### Prometheus Alerts

Create alert rules in `monitoring/prometheus-alerts.yml`:

```yaml
groups:
  - name: propsdb
    rules:
      - alert: HighErrorRate
        expr: rate(propsdb_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
```

### Grafana Alerts

1. Open a dashboard panel
2. Click Alert tab
3. Configure alert conditions
4. Set notification channels

## Monitoring Best Practices

### Key Metrics to Watch

1. **Request Rate**: Sudden spikes or drops
2. **Error Rate**: Should be < 1%
3. **Response Time**: P95 should be < 500ms
4. **Database Connections**: Should not hit pool limit
5. **Memory Usage**: Watch for memory leaks

### Dashboard Organization

- **Overview**: High-level metrics (requests, errors, latency)
- **Database**: Connection pools, query performance
- **System**: CPU, memory, goroutines
- **Business**: Custom application metrics

## Troubleshooting

### Metrics Not Appearing

```bash
# Check if metrics endpoint is accessible
curl http://localhost:3000/metrics

# Check Prometheus targets
# Visit http://localhost:9090/targets
```

### Grafana Connection Issues

```bash
# Check Grafana logs
docker logs propsdb-grafana

# Verify Prometheus datasource
# Grafana → Configuration → Data Sources
```

### High Memory Usage

```bash
# Check Go runtime metrics
curl http://localhost:3000/metrics | grep go_memstats

# Force garbage collection (development only)
# Add endpoint: /debug/gc
```

## Production Recommendations

1. **Retention**: Configure Prometheus retention (default: 15 days)
2. **Storage**: Use persistent volumes for Prometheus data
3. **Backup**: Regular backups of Grafana dashboards
4. **Security**: Enable authentication for Prometheus/Grafana
5. **Scaling**: Consider Prometheus federation for large deployments

## Integration with External Systems

### Datadog

```bash
# Use Datadog Prometheus integration
# Configure in Datadog agent
```

### New Relic

```bash
# Use New Relic Prometheus integration
# Or use New Relic Go agent
```

### CloudWatch

```bash
# Use CloudWatch Prometheus exporter
# Or AWS Distro for OpenTelemetry
```

## Custom Metrics

Add custom metrics in your code:

```go
import "github.com/prometheus/client_golang/prometheus"

var customCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "propsdb_custom_operations_total",
        Help: "Total custom operations",
    },
    []string{"operation", "status"},
)

func init() {
    prometheus.MustRegister(customCounter)
}

// In your code
customCounter.WithLabelValues("upsert", "success").Inc()
```

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)
