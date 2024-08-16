# OpenTelemetry Collector

This folder contains very simple open telemetry setup.

```zsh
docker compose up -d
```

That will expose the following backends:

- Jaeger at http://0.0.0.0:16686
- Zipkin at http://0.0.0.0:9411
- Prometheus at http://0.0.0.0:9090
