# Load Testing Guide for Pulse Control Plane

This directory contains load tests for the Pulse Control Plane API using k6.

## Prerequisites

1. Install k6:
```bash
# macOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6
```

## Test Files

- **load_test.js**: Standard load test with gradual ramp-up
- **spike_test.js**: Tests system behavior during sudden traffic spikes
- **soak_test.js**: Long-duration test to check for memory leaks and performance degradation

## Running Tests

### Local Testing

```bash
# Basic load test
k6 run load_test.js

# Load test with custom base URL
k6 run -e BASE_URL=http://localhost:8081 load_test.js

# Load test with API key
k6 run -e BASE_URL=http://localhost:8081 \
       -e API_KEY=pulse_key_xxx \
       -e API_SECRET=pulse_secret_xxx \
       load_test.js

# Spike test
k6 run spike_test.js

# Soak test (runs for ~14 minutes)
k6 run soak_test.js
```

### Production Testing

```bash
# Test against production (use with caution)
k6 run -e BASE_URL=https://your-domain.com load_test.js

# Test with specific VUs and duration
k6 run --vus 100 --duration 30s load_test.js
```

### Docker

```bash
# Run load test in Docker
docker run --rm -i grafana/k6 run - < load_test.js

# With environment variables
docker run --rm -i \
  -e BASE_URL=http://host.docker.internal:8081 \
  grafana/k6 run - < load_test.js
```

## Test Scenarios

### 1. Load Test (load_test.js)

**Purpose**: Test system performance under expected load

**Profile**:
- Ramp up from 0 to 200 users over 4.5 minutes
- Sustain peak load
- Ramp down gradually

**Endpoints Tested**:
- `/health` - Health check
- `/api/v1/status` - System status
- `/api/v1/tokens/create` - Token generation (with auth)
- `/api/v1/organizations` - Organization listing

**Expected Results**:
- 95% of requests < 500ms
- Error rate < 1%
- No timeouts

### 2. Spike Test (spike_test.js)

**Purpose**: Test system resilience during sudden traffic surges

**Profile**:
- Sudden spike from 10 to 500 users
- Sustain for 30 seconds
- Drop back to baseline

**Expected Results**:
- System remains stable
- 95% of requests < 1s
- Error rate < 5%
- No crashes

### 3. Soak Test (soak_test.js)

**Purpose**: Test for memory leaks and performance degradation over time

**Profile**:
- Sustain 50 users for 10 minutes
- Realistic user behavior with think time

**Expected Results**:
- Consistent response times throughout
- No memory leaks
- Error rate < 1%
- 99% of requests < 800ms

## Interpreting Results

### Key Metrics

```
http_reqs..................: Total number of requests made
http_req_duration..........: Request duration
  - avg...................: Average duration
  - min...................: Minimum duration
  - med...................: Median duration  
  - max...................: Maximum duration
  - p(90).................: 90th percentile
  - p(95).................: 95th percentile
http_req_failed...........: Percentage of failed requests
errors....................: Custom error rate
```

### Success Criteria

✅ **Pass**:
- p(95) < 500ms for load test
- p(95) < 1000ms for spike test
- Error rate < 1% for all tests
- No timeouts or connection errors

⚠️ **Warning**:
- p(95) between 500-1000ms
- Error rate 1-5%
- Occasional timeouts

❌ **Fail**:
- p(95) > 1000ms
- Error rate > 5%
- Frequent timeouts or crashes

## Load Testing Best Practices

### Before Testing

1. **Notify team**: Inform team about load testing schedule
2. **Use test data**: Ensure test database is used, not production
3. **Monitor resources**: Set up monitoring for CPU, memory, database
4. **Set baseline**: Run tests on empty system first

### During Testing

1. **Monitor in real-time**: Watch logs and metrics
2. **Check database**: Monitor MongoDB connection pool
3. **Watch for errors**: Look for rate limiting, timeouts
4. **Resource usage**: CPU, memory, network should be within limits

### After Testing

1. **Analyze results**: Compare against thresholds
2. **Identify bottlenecks**: Find slow endpoints
3. **Check logs**: Look for errors or warnings
4. **Document findings**: Record results for future comparison

## Simulating Large Scale

### Testing with Lakhs of Requests

To simulate lakhs (100,000+) of requests:

```bash
# 1 lakh requests with 1000 concurrent users
k6 run --vus 1000 --iterations 100000 load_test.js

# 5 lakh requests over 10 minutes
k6 run --vus 500 --duration 10m load_test.js

# Monitor MongoDB during test
watch -n 1 'mongo --eval "db.serverStatus().connections"'
```

### Expected Performance

**Target Metrics for Production**:
- Handle 1000 requests/second
- Support 10,000 concurrent users
- 99% of requests under 1 second
- 0.1% error rate
- 99.9% uptime

## Troubleshooting

### Common Issues

1. **High error rates**:
   - Check rate limiting settings
   - Verify API keys are valid
   - Check database connection pool

2. **Slow response times**:
   - Check database indexes
   - Review query performance
   - Monitor CPU and memory

3. **Connection timeouts**:
   - Increase timeout settings
   - Check network capacity
   - Review connection pool size

### MongoDB Optimization

```javascript
// Check connection pool status
db.serverStatus().connections

// Check slow queries
db.setProfilingLevel(1, { slowms: 100 })
db.system.profile.find().sort({ ts: -1 }).limit(10)

// Create indexes
db.projects.createIndex({ "pulse_api_key": 1 })
db.usage_metrics.createIndex({ "project_id": 1, "timestamp": -1 })
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Load Testing

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6
      - name: Run load test
        run: k6 run -e BASE_URL=${{ secrets.API_URL }} tests/load_test.js
```

## References

- [k6 Documentation](https://k6.io/docs/)
- [k6 Test Types](https://k6.io/docs/test-types/introduction/)
- [k6 Thresholds](https://k6.io/docs/using-k6/thresholds/)
- [k6 Cloud](https://k6.io/cloud/)
