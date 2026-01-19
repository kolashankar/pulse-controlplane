# Scaling Guide - Pulse Control Plane

Guide to scaling your Pulse deployment for production workloads.

---

## Table of Contents

1. [Overview](#overview)
2. [Infrastructure Setup](#infrastructure-setup)
3. [Database Optimization](#database-optimization)
4. [Application Scaling](#application-scaling)
5. [Caching Strategy](#caching-strategy)
6. [Load Balancing](#load-balancing)
7. [Monitoring & Alerts](#monitoring--alerts)
8. [Performance Tuning](#performance-tuning)
9. [Cost Optimization](#cost-optimization)

---

## Overview

### Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| API Response Time | < 200ms (p95) | For token generation |
| Throughput | 10,000 req/s | Per instance |
| Concurrent Users | 100,000+ | System-wide |
| Database Queries | < 10ms (p95) | With indexes |
| Uptime | 99.9%+ | 43 minutes downtime/month |

### Scaling Dimensions

1. **Horizontal Scaling**: Add more instances
2. **Vertical Scaling**: Increase instance resources
3. **Database Scaling**: Optimize queries and indexes
4. **Caching**: Reduce database load
5. **CDN**: Serve static assets

---

## Infrastructure Setup

### Recommended Architecture

```
                    Internet
                       |
                   [CDN/Cloudflare]
                       |
                [Load Balancer]
                   /       \
        [API Server 1]  [API Server 2]  ... [API Server N]
                   \       /
                [MongoDB Cluster]
                   /       \
            [Primary]   [Replicas]
```

### Cloud Providers

#### AWS Setup

```yaml
# EC2 Instances
Instance Type: t3.large (2 vCPU, 8 GB RAM)
Auto Scaling: 2-10 instances
Region: us-east-1 (multiple AZs)

# MongoDB Atlas
Cluster: M30 (8 GB RAM, 2 vCPU)
Replication: 3 nodes
Backups: Continuous

# Load Balancer
Type: Application Load Balancer
Health Check: /health endpoint
```

#### Google Cloud Setup

```yaml
# Compute Engine
Machine Type: n1-standard-2
Auto Scaling: 2-10 instances
Region: us-central1

# MongoDB Atlas
Same as AWS

# Load Balancer
Type: HTTP(S) Load Balancer
```

#### DigitalOcean Setup

```yaml
# Droplets
Size: 4 GB RAM, 2 vCPUs
Load Balancer: Yes
Managed MongoDB: 4 GB RAM cluster
```

---

## Database Optimization

### Essential MongoDB Indexes

```javascript
// Organizations collection
db.organizations.createIndex({ "admin_email": 1 });
db.organizations.createIndex({ "plan": 1 });
db.organizations.createIndex({ "is_deleted": 1 });

// Projects collection
db.projects.createIndex({ "pulse_api_key": 1 }, { unique: true });
db.projects.createIndex({ "org_id": 1 });
db.projects.createIndex({ "region": 1 });
db.projects.createIndex({ "is_deleted": 1 });

// Usage metrics collection
db.usage_metrics.createIndex({ "project_id": 1, "timestamp": -1 });
db.usage_metrics.createIndex({ "timestamp": -1 });

// Audit logs collection
db.audit_logs.createIndex({ "user_email": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "action": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "timestamp": -1 }, { expireAfterSeconds: 31536000 }); // 1 year TTL

// Webhooks collection
db.webhooks.createIndex({ "project_id": 1, "timestamp": -1 });
db.webhooks.createIndex({ "status": 1 });

// Team members collection
db.team_members.createIndex({ "org_id": 1, "email": 1 }, { unique: true });
db.team_members.createIndex({ "email": 1 });
```

### Connection Pooling

```go
// config/mongodb.go
clientOptions := options.Client().
    ApplyURI(mongoURL).
    SetMaxPoolSize(100).          // Max connections
    SetMinPoolSize(10).            // Min connections
    SetMaxConnIdleTime(30 * time.Minute).
    SetServerSelectionTimeout(5 * time.Second).
    SetConnectTimeout(10 * time.Second)
```

### Query Optimization

```go
// BAD: No index, slow query
db.collection("projects").Find(context.Background(), bson.M{
    "name": bson.M{"$regex": "test", "$options": "i"},
})

// GOOD: Use indexed field
db.collection("projects").Find(context.Background(), bson.M{
    "pulse_api_key": apiKey,
    "is_deleted": false,
})

// GOOD: Limit results
opts := options.Find().SetLimit(20).SetProjection(bson.M{
    "name": 1,
    "org_id": 1,
    "created_at": 1,
})
```

### Aggregation Pipelines

```go
// Efficient aggregation for usage metrics
pipeline := mongo.Pipeline{
    {{Key: "$match", Value: bson.M{
        "project_id": projectID,
        "timestamp": bson.M{
            "$gte": startDate,
            "$lte": endDate,
        },
    }}},
    {{Key: "$group", Value: bson.M{
        "_id": "$date",
        "participant_minutes": bson.M{"$sum": "$participant_minutes"},
        "egress_minutes": bson.M{"$sum": "$egress_minutes"},
    }}},
    {{Key: "$sort", Value: bson.M{"_id": 1}}},
}
```

---

## Application Scaling

### Horizontal Scaling

#### Docker Compose (Development)

```yaml
version: '3.8'

services:
  api:
    image: pulse-api:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 1G
    environment:
      - MONGO_URL=mongodb://mongo:27017/pulse
      - GO_ENV=production
    ports:
      - "8081-8083:8081"
```

#### Kubernetes (Production)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pulse-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pulse-api
  template:
    metadata:
      labels:
        app: pulse-api
    spec:
      containers:
      - name: pulse-api
        image: pulse-api:latest
        ports:
        - containerPort: 8081
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        env:
        - name: MONGO_URL
          valueFrom:
            secretKeyRef:
              name: pulse-secrets
              key: mongo-url
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: pulse-api
spec:
  selector:
    app: pulse-api
  ports:
  - port: 80
    targetPort: 8081
  type: LoadBalancer
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: pulse-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: pulse-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Vertical Scaling

```yaml
# Increase instance resources
resources:
  requests:
    memory: "2Gi"
    cpu: "2000m"
  limits:
    memory: "4Gi"
    cpu: "4000m"
```

---

## Caching Strategy

### Redis Cache Layer

```go
import "github.com/go-redis/redis/v8"

// Initialize Redis client
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
})

// Cache API keys (30 minute TTL)
func getProject(apiKey string) (*Project, error) {
    // Try cache first
    cached, err := rdb.Get(ctx, "project:"+apiKey).Result()
    if err == nil {
        var project Project
        json.Unmarshal([]byte(cached), &project)
        return &project, nil
    }
    
    // Cache miss, query database
    project, err := db.FindProject(apiKey)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    data, _ := json.Marshal(project)
    rdb.Set(ctx, "project:"+apiKey, data, 30*time.Minute)
    
    return project, nil
}

// Cache usage metrics (5 minute TTL)
func getUsageMetrics(projectID string) (*UsageMetrics, error) {
    cacheKey := fmt.Sprintf("usage:%s:%s", projectID, time.Now().Format("2006-01-02"))
    
    cached, err := rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var metrics UsageMetrics
        json.Unmarshal([]byte(cached), &metrics)
        return &metrics, nil
    }
    
    // Query database
    metrics, err := db.GetUsageMetrics(projectID)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    data, _ := json.Marshal(metrics)
    rdb.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return metrics, nil
}
```

### In-Memory Cache (For High-Frequency Reads)

```go
import "github.com/patrickmn/go-cache"

// Create cache with 5 minute default expiration
c := cache.New(5*time.Minute, 10*time.Minute)

// Cache project data
c.Set("project:"+apiKey, project, cache.DefaultExpiration)

// Retrieve from cache
if x, found := c.Get("project:" + apiKey); found {
    project := x.(*Project)
    return project
}
```

---

## Load Balancing

### Nginx Configuration

```nginx
upstream pulse_api {
    least_conn;
    server api1.pulse.io:8081 weight=1 max_fails=3 fail_timeout=30s;
    server api2.pulse.io:8081 weight=1 max_fails=3 fail_timeout=30s;
    server api3.pulse.io:8081 weight=1 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name api.pulse.io;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.pulse.io;
    
    ssl_certificate /etc/ssl/certs/pulse.crt;
    ssl_certificate_key /etc/ssl/private/pulse.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
    limit_req zone=api_limit burst=20 nodelay;
    
    location /api/ {
        proxy_pass http://pulse_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    location /health {
        proxy_pass http://pulse_api/health;
        access_log off;
    }
}
```

### Health Checks

```go
// handlers/health_handler.go
func HealthCheck(c *gin.Context) {
    // Check database connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    err := database.GetDB().Ping(ctx, nil)
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error": "database unavailable",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
        "version": "1.0.0",
    })
}
```

---

## Monitoring & Alerts

### Prometheus Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pulse_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "pulse_http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "pulse_active_connections",
            Help: "Number of active connections",
        },
    )
)

// Middleware to track metrics
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())
        
        requestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
        requestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
    }
}
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Pulse Control Plane",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(pulse_http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, pulse_http_request_duration_seconds_bucket)"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(pulse_http_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      }
    ]
  }
}
```

### Alert Rules

```yaml
groups:
- name: pulse_alerts
  rules:
  - alert: HighErrorRate
    expr: rate(pulse_http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors/sec"
  
  - alert: SlowResponses
    expr: histogram_quantile(0.95, pulse_http_request_duration_seconds_bucket) > 1
    for: 10m
    annotations:
      summary: "API responses are slow"
      description: "P95 latency is {{ $value }} seconds"
  
  - alert: DatabaseDown
    expr: up{job="mongodb"} == 0
    for: 1m
    annotations:
      summary: "MongoDB is down"
```

---

## Performance Tuning

### Go Application Tuning

```go
// main.go
func main() {
    // Set GOMAXPROCS to number of CPUs
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Configure Gin
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    
    // Use faster JSON library
    router.Use(gin.Recovery())
    
    // Connection pooling
    db.SetMaxOpenConns(100)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(time.Hour)
}
```

### MongoDB Tuning

```javascript
// Increase connection pool size
db.adminCommand({
    setParameter: 1,
    maxConns: 10000
});

// Enable query profiling
db.setProfilingLevel(1, { slowms: 100 });

// Check slow queries
db.system.profile.find().sort({ ts: -1 }).limit(10);
```

---

## Cost Optimization

### Resource Right-Sizing

| Load | Instance Type | Monthly Cost (AWS) |
|------|---------------|-------------------|
| Light (< 1K users) | t3.small | $15 |
| Medium (1K-10K users) | t3.medium | $30 |
| Heavy (10K-100K users) | t3.large | $60 |
| Very Heavy (100K+ users) | t3.xlarge | $120 |

### Database Optimization

- Use Atlas M0 (Free) for development
- Use M10 ($57/month) for staging
- Use M30+ ($240/month) for production

### CDN Usage

- Store recordings on S3 + CloudFront
- Serve static assets via CDN
- Estimated savings: 60-80% on bandwidth

---

## Capacity Planning

### Formulas

```
RPS (Requests Per Second) = (Daily Active Users × Avg Requests Per User) / 86400
Required Instances = RPS / (Target RPS Per Instance)
Database IOPS = RPS × Avg Queries Per Request
```

### Example Calculation

```
100,000 DAU × 50 requests/day = 5,000,000 requests/day
5,000,000 / 86400 = 58 RPS

At 1000 RPS per instance:
58 / 1000 = 0.058 instances (minimum 2 for HA)

Database:
58 RPS × 2 queries = 116 IOPS (easily handled by M30)
```

---

## Resources

- [MongoDB Performance Best Practices](https://docs.mongodb.com/manual/administration/analyzing-mongodb-performance/)
- [Go Performance Tips](https://github.com/dgryski/go-perfbook)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)

---

**Last Updated**: 2025-01-19
