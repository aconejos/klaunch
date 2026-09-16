# Enhanced Observability — Plan

Configurable Kafka Connect scrape target for the existing Prometheus + Grafana stack, plus a Connect-focused dashboard.

## Goal

- Point the running Grafana/Prometheus at any Kafka Connect JMX Prometheus endpoint (local by default, remote when needed).
- Ship one Connect-focused dashboard that proves the pipeline works end to end.

## Non-goals

- Not installing the JMX Prometheus javaagent on remote workers. Remote scraping requires the target Connect to already expose `/metrics`.
- Not touching broker or ZooKeeper scrape targets. Only the `kafka-connect` job is configurable.
- Not managing Grafana dashboards through the CLI. New dashboards = drop JSON into `volumes/dashboards/`.
- No new Go dependency; no template engine; no Prometheus lifecycle API.

## Current state (verified before planning)

- `docker-compose.yaml` already runs `prometheus` (host `9090`) and `grafana` (host `3000`, admin/foobar).
- JMX Prometheus javaagents already mounted on `kafka1/2/3` (container `:8091`), `zookeeper1` (container `:8091`, host `8094`), `kafka-connect` (container `:8091`, host `8095`).
- Scrape rules live in `volumes/kafka_config.yml`, `volumes/kafka_connect.yml`, `volumes/zookeeper_config.yml`.
- `volumes/prometheus.yml` has three static jobs (`kafka`, `zookeeper`, `kafka-connect`) with hardcoded container-network targets.
- Grafana auto-provisions from `volumes/provisioning/{datasources,dashboards}/` and loads everything under `volumes/dashboards/`.
- No CLI subcommand touches Prometheus or Grafana today.

## Design decisions

- **Scope:** override the `kafka-connect` job's target only.
- **CLI shape:** one new top-level command, one positional arg.
  - `klaunch observability` — prints current target + Prometheus/Grafana URLs.
  - `klaunch observability <host:port>` — rewrites the `kafka-connect` job's target in `volumes/prometheus.yml`, then `docker restart prometheus`.
- **Auth:** keep Grafana `admin`/`foobar`. Local repro only; documented.
- **Dashboards:** keep the three existing ones (`kafka-overview`, `zookeeper-overview`, `kafka-connect-cluster`), add one Connect-focused board.
- **How the rewrite works:** in-place regex on the `kafka-connect` job block. No template file, no YAML parser.
- **How the reload works:** `docker restart prometheus`. Simpler than enabling `--web.enable-lifecycle` and POST `/-/reload`.

## Files touched

New:
- `docs/observability-plan.md` (this file)
- `observability.go` — Cobra wiring + `rewriteConnectTarget(hostport string) error`. ~40 lines.
- `observability_test.go` — one table test on `rewriteConnectTarget`. ~30 lines.
- `volumes/dashboards/klaunch-connect-tasks.json` — Connect-focused dashboard.

Edited:
- `main.go` — register the new command.
- `README.md` — Observability section + remote JMX caveat.
- `AGENTS.md` — one section on the new command + "do not hand-edit `volumes/prometheus.yml` when it's the current target" note.

Not touched:
- `docker-compose.yaml` (no Prometheus flag change, no `extra_hosts` on `prometheus` — add only if a real user hits it).
- `volumes/prometheus.yml` — edited by the CLI at runtime, not converted to a template.

## Dashboard content

Driven by rules already in `volumes/kafka_connect.yml`. Panels:

- Connector state count (running/paused/failed) from `kafka_connect_connector_metrics`.
- Per-connector throughput from `sum by (connector) (kafka_connect_source_task_metrics_source_record_write_total)` and the sink equivalent.
- Errors per (connector, task) from `kafka_connect_task_error_metrics_total_errors_logged`.
- DLQ pressure from `kafka_connect_task_error_metrics_deadletterqueue_produce_requests_total`.
- Worker rebalance activity from `kafka_connect_connect_worker_rebalance_metrics_*`.

**MongoDB Sink Task row** (added after the plan was drafted, per the "justin sink connector" reference doc). Covers all 10 sink-task metrics exposed by the MongoDB Kafka Connector 3.x under the `com.mongodb.kafka.connect` JMX domain:

| Panel | Metric(s) |
|---|---|
| MongoDB records written (rate) | `com_mongodb_kafka_connect_sink_task_metrics_records_successful` |
| Total records written (since restart) | same, as a stat |
| Kafka-vs-Sink lag | `com_mongodb_kafka_connect_sink_task_metrics_latest_kafka_time_difference_ms` |
| put() calls/sec + avg latency | `in_task_put`, `in_task_put_duration_ms` |
| Kafka Connect framework overhead | `in_connect_framework`, `in_connect_framework_duration_ms` |
| SMT / processing phase time | `processing_phases`, `processing_phases_duration_ms` |
| Batch writes to MongoDB (rate + avg latency) | `batch_writes_successful`, `batch_writes_successful_duration_ms` |

The whitelist entry in `volumes/kafka_connect.yml` was updated from `com.mongodb:*` to `com.mongodb.kafka.connect:*` — the former matched no MBeans because the connector uses a different JMX domain.

Metric expressions validated against actual `/metrics` output during the smoke test.

## Rollout order

1. Land the plan (this file) on `enhanced-observability`, open PR to `main`.
2. Add `observability.go` + test, wire into `main.go`.
3. Add the Connect dashboard JSON.
4. Update `README.md` + `AGENTS.md`.
5. Run the smoke test below; capture the four evidence artifacts.
6. Merge.

## Testing phase

### Automated (no Docker)

- `go test -run TestRewriteConnectTarget .`
- `./run_tests.sh static`

### End-to-end smoke — "working as advertised"

Scripted in `scripts/smoke_observability.sh`. Runs steps 3–9 below with hard
assertions (exits non-zero on the first failure) and cleans up via an EXIT
trap so a failed run doesn't leave stale connectors or a smoke Mongo behind.

    make build
    ./klaunch start   # or pass --start to the script
    ./scripts/smoke_observability.sh
    # optional: --start (bring stack up first) / --stop (klaunch stop at end)

Manual walk-through (what the script does):

1. **Fresh state.** `./klaunch stop && docker volume rm $(docker volume ls -q -f name=klaunch) 2>/dev/null || true` → no `klaunch_*` containers, no leftover volumes.
2. **Start full infra.** `./klaunch start` → `docker ps` shows `zookeeper1`, `kafka1/2/3`, `kafka-connect`, `schema-registry`, `cmak`, `prometheus`, `grafana`. `.env` shows the pulled `MONGO_KAFKA_CONNECT_VERSION`.
3. **Prometheus targets healthy.** `curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job:.labels.job, health:.health}'` → `kafka`, `zookeeper`, `kafka-connect` all `up`.
4. **Grafana provisioning wired.**
   - `curl -s -u admin:foobar http://localhost:3000/api/datasources | jq '.[].name'` → contains `Prometheus`.
   - `curl -s -u admin:foobar 'http://localhost:3000/api/search?type=dash-db' | jq -r '.[].title'` → lists four dashboards including `klaunch-connect-tasks`.
5. **CLI reports state.** `./klaunch observability` → prints current target `kafka-connect:8091`, `http://localhost:9090`, `http://localhost:3000`.
6. **Create load.** `./klaunch create` → pick `default_source_task.json`. Insert docs via `npm run insert:multiple` so the source connector emits records.
7. **Connect dashboard populated.** In Grafana `klaunch-connect-tasks`:
   - Connector state panel shows one `running` connector.
   - Throughput panel shows non-zero source record writes.
   - Errors panel shows zero on the happy path.
   - Optional negative check: `./klaunch create` → `4_mask_field_sink_task_failure.json`; confirm the errors/DLQ panels light up.
8. **Retarget test.** `./klaunch observability host.docker.internal:8095` (host-exposed port of the local Connect JMX exporter; stays scrapable without spinning up a second Connect). Wait ~15s. Re-check `/api/v1/targets` → `kafka-connect` target now `host.docker.internal:8095`, still `up`. Panels still populate. Proves the rewrite + restart path without a real remote worker.
9. **Reset.** `./klaunch observability kafka-connect:8091` → target back to default, still `up`.
10. **Teardown.** `./klaunch stop` → all `klaunch` containers gone. `volumes/prometheus.yml` retains the last written target by design; users who want a clean slate re-run step 9 first.

### Evidence to capture

Not committed (screenshots bloat history). Saved locally for the PR description:

- `01_targets.json` — output of the `/api/v1/targets` call after step 3.
- `02_dashboards.txt` — Grafana dashboard list after step 4.
- `03_connect_running.png` — screenshot of the Connect dashboard with load (step 7).
- `04_target_switched.json` — `/api/v1/targets` after step 8.
- `05_target_reset.json` — after step 9.

## Caveats to document

- Remote Kafka Connect must already run `-javaagent:/path/jmx_prometheus_javaagent.jar=PORT:kafka_connect.yml` and expose PORT to the network Prometheus can reach. `volumes/kafka_connect.yml` is the mapping config we ship; users can reuse it.
- From inside the `prometheus` container, `host.docker.internal` requires the `extra_hosts: ["host.docker.internal:host-gateway"]` entry on the `prometheus` service. Currently only Kafka services declare it. If a user reports "host.docker.internal doesn't resolve", add that entry. Deferred until then.
- `./klaunch stop` does not reset the scrape target. Intentional: keeps the last configuration across sessions.

## `ponytail:` markers to leave in code when implemented

- `// ponytail: docker restart instead of prom lifecycle API; swap if reloads become frequent.`
- `// ponytail: line-rewrite instead of YAML parse; template it if a second field ever needs configuring.`
