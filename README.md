# go-redis cache-miss metrics

A `GET` for a missing key returns `redis.Nil`. That is a normal cache miss, not
a failed Redis operation. With go-redis `v9.22.0` native OpenTelemetry metrics,
each miss is nevertheless recorded in `redis.client.errors` as an unknown
internal error.

This makes an expected cache-miss rate look like a client failure rate.

## Reproduce

Run:

```bash
make
```

This starts Redis, the Go reproducer, OpenTelemetry Collector, Prometheus, and
Grafana in Docker Compose. The reproducer performs ten `GET` requests per second
for a missing key and sends its metrics to the Collector.

Open <http://localhost:3000> and sign in with `admin` / `admin`. The provisioned
dashboard is **Redis / Redis Client Observability Dashboard**.

The `redis.client.errors` value keeps increasing by approximately ten every
second with these attributes:

```text
error.type = UNKNOWN
redis.client.errors.internal = true
```

Stop everything with:

```bash
make stop
```
