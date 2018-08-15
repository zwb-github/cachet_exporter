# cachet_exporter

Exports metrics from Cachet status page for consumption by Prometheus.

### Running

```console
./cachet_exporter --cachet.api-url=${CACHET_API_URL}
```

Or with docker:

```console
docker run -p 9470:9470 -e "CACHET_API_URL=${CACHET_API_URL}" caninjas/cachet_exporter
```

### Flags

Name               | Description
-------------------|---------------------------------------------------------
web.listen-address | Address to listen on for web interface and telemetry (default `:9470`)
web.telemetry-path | Path under which to expose metrics (default `/metrics`)
cachet.api-url     | Your Cachet instance API URL (required)
