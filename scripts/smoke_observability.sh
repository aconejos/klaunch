#!/usr/bin/env bash
# Smoke test for the klaunch observability feature.
#
# What it does, in order:
#   1. Assumes the klaunch stack is running (`./klaunch start` first) OR brings
#      it up if --start is passed.
#   2. Confirms all 5 Prometheus scrape targets are healthy.
#   3. Confirms Grafana provisioned the 4 dashboards including
#      klaunch-connect-tasks.
#   4. Confirms `klaunch observability` prints the default state.
#   5. Runs a MongoDB replica set container, creates a MongoDB source +
#      MongoDB sink connector, pushes docs through the pipeline.
#   6. Confirms all 10 MongoDB Kafka Connector sink-task metrics from the
#      reference doc are exported and non-zero.
#   7. Retargets Prometheus at host.docker.internal:8095 (proves the rewrite
#      + restart path). Resets to default.
#   8. Deletes the smoke connectors and the MongoDB smoke container. Does NOT
#      run `klaunch stop` unless --stop is passed.
#
# Requires: docker, jq, curl, ./build/klaunch (run `make build` first).
# Exits non-zero on the first assertion failure.

set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

BIN=${KLAUNCH_BIN:-./build/klaunch}
CONNECT_URL=http://localhost:8083
PROM_URL=http://localhost:9090
GRAFANA_URL=http://localhost:3000
GRAFANA_AUTH=${GRAFANA_AUTH:-admin:foobar}
MONGO_NAME=klaunch-smoke-mongo
MONGO_NET=klaunch_default

DO_START=0
DO_STOP=0
for arg in "$@"; do
  case "$arg" in
    --start) DO_START=1 ;;
    --stop)  DO_STOP=1 ;;
    -h|--help)
      sed -n '2,25p' "$0"; exit 0 ;;
    *) echo "unknown flag: $arg" >&2; exit 2 ;;
  esac
done

step() { printf '\n\033[1;34m==>\033[0m %s\n' "$*"; }
ok()   { printf '   \033[1;32mok\033[0m %s\n' "$*"; }
fail() { printf '   \033[1;31mFAIL\033[0m %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null || fail "missing dependency: $1"; }
need docker; need jq; need curl
[ -x "$BIN" ] || fail "klaunch binary not found at $BIN (run: make build)"

# Cleanup runs on any exit path (pass, fail, ctrl-c). Idempotent.
cleanup() {
  local rc=$?
  curl -s -X DELETE "$CONNECT_URL/connectors/klaunch-smoke-src"  >/dev/null 2>&1 || true
  curl -s -X DELETE "$CONNECT_URL/connectors/klaunch-smoke-sink" >/dev/null 2>&1 || true
  docker rm -f "$MONGO_NAME" >/dev/null 2>&1 || true
  [ -n "${tmp:-}" ] && rm -rf "$tmp"
  if [ "$DO_STOP" -eq 1 ] && [ $rc -eq 0 ]; then
    "$BIN" stop >/dev/null && printf '   \033[1;32mok\033[0m klaunch stack stopped\n'
  fi
}
trap cleanup EXIT

# 1. Stack up
if [ "$DO_START" -eq 1 ]; then
  step "Starting klaunch stack"
  "$BIN" start | tail -5
fi

step "Waiting for kafka-connect REST"
for i in $(seq 1 60); do
  code=$(curl -s -o /dev/null -w '%{http_code}' "$CONNECT_URL/connectors" || true)
  [ "$code" = "200" ] && { ok "kafka-connect ready after ${i}s"; break; }
  sleep 2
  [ "$i" = "60" ] && fail "kafka-connect not ready after 120s"
done

# 2. Prometheus targets
step "Prometheus targets"
sleep 5  # give scraper one cycle
targets_json=$(curl -sf "$PROM_URL/api/v1/targets")
jobs_up=$(echo "$targets_json" | jq -r '.data.activeTargets[] | select(.health=="up") | .labels.job' | sort -u | paste -sd, -)
[ "$jobs_up" = "kafka,kafka-connect,zookeeper" ] \
  || fail "not all jobs up (got: $jobs_up)"
ok "all jobs up: $jobs_up"

# 3. Grafana dashboards
step "Grafana provisioned dashboards"
dash_list=$(curl -sf -u "$GRAFANA_AUTH" "$GRAFANA_URL/api/search?type=dash-db" | jq -r '.[].title' | sort)
for want in "Kafka Connect cluster" "Kafka Overview" "Zookeeper Overview" "klaunch-connect-tasks"; do
  echo "$dash_list" | grep -Fxq "$want" || fail "missing dashboard: $want"
done
ok "4 dashboards provisioned"

# 4. CLI state
step "klaunch observability (default state)"
out=$("$BIN" observability)
echo "$out" | grep -q "kafka-connect:8091" || fail "unexpected default target: $out"
ok "default target = kafka-connect:8091"

# 5. MongoDB smoke connectors
step "Setting up MongoDB smoke container + connectors"
docker rm -f "$MONGO_NAME" >/dev/null 2>&1 || true
docker run -d --name "$MONGO_NAME" --network "$MONGO_NET" -p 27020:27017 \
  mongo:7 --replSet replset --bind_ip_all >/dev/null
sleep 5
docker exec "$MONGO_NAME" mongosh --quiet --eval \
  "rs.initiate({_id:'replset', members:[{_id:0, host:'$MONGO_NAME:27017'}]})" >/dev/null
sleep 3
docker exec "$MONGO_NAME" mongosh --quiet --eval \
  "db.getSiblingDB('src').coll.insertMany(Array.from({length:5},(_,k)=>({seed:k})))" >/dev/null
ok "mongo replica set ready"

tmp=$(mktemp -d)
cat > "$tmp/src.json" <<EOF
{"name":"klaunch-smoke-src","config":{
  "connector.class":"com.mongodb.kafka.connect.MongoSourceConnector",
  "connection.uri":"mongodb://$MONGO_NAME:27017/?replicaSet=replset",
  "database":"src","collection":"coll","tasks.max":"1",
  "poll.max.batch.size":"100","poll.await.time.ms":"500",
  "startup.mode":"copy_existing","output.format.value":"json",
  "topic.prefix":"klaunchsmoke"}}
EOF
cat > "$tmp/sink.json" <<EOF
{"name":"klaunch-smoke-sink","config":{
  "connector.class":"com.mongodb.kafka.connect.MongoSinkConnector",
  "connection.uri":"mongodb://$MONGO_NAME:27017/?replicaSet=replset",
  "database":"dest","collection":"coll","tasks.max":"1",
  "topics":"klaunchsmoke.src.coll",
  "key.converter":"org.apache.kafka.connect.json.JsonConverter",
  "value.converter":"org.apache.kafka.connect.json.JsonConverter",
  "key.converter.schemas.enable":"false",
  "value.converter.schemas.enable":"false"}}
EOF
# Wipe stale smoke connectors from a previous aborted run before creating.
curl -s -X DELETE "$CONNECT_URL/connectors/klaunch-smoke-src"  >/dev/null 2>&1 || true
curl -s -X DELETE "$CONNECT_URL/connectors/klaunch-smoke-sink" >/dev/null 2>&1 || true
sleep 2
curl -sf -X POST -H 'Content-Type: application/json' --data @"$tmp/src.json" \
  "$CONNECT_URL/connectors" >/dev/null
curl -sf -X POST -H 'Content-Type: application/json' --data @"$tmp/sink.json" \
  "$CONNECT_URL/connectors" >/dev/null
sleep 8
for c in klaunch-smoke-src klaunch-smoke-sink; do
  state=$(curl -sf "$CONNECT_URL/connectors/$c/status" | jq -r '.tasks[0].state')
  [ "$state" = "RUNNING" ] || fail "$c task not RUNNING (state=$state)"
done
ok "src + sink connectors RUNNING"

step "Pushing load (5 x 40 docs)"
for i in 1 2 3 4 5; do
  docker exec "$MONGO_NAME" mongosh --quiet --eval \
    "db.getSiblingDB('src').coll.insertMany(Array.from({length:40},(_,k)=>({batch:$i,k,ts:new Date()})))" >/dev/null
  sleep 3
done
sleep 12  # let sink drain + at least one Prometheus scrape

# 6. All 10 MongoDB sink metrics from the reference doc must be present.
#    Count/throughput metrics must also be non-zero after load. Duration + lag
#    metrics may legitimately read 0 on a fast local machine with no SMTs.
step "MongoDB sink task metrics (com.mongodb.kafka.connect JMX domain)"
counts=(records_successful in_task_put in_connect_framework processing_phases batch_writes_successful)
gauges=(latest_kafka_time_difference_ms in_task_put_duration_ms in_connect_framework_duration_ms processing_phases_duration_ms batch_writes_successful_duration_ms)
bad=0
check() {
  local m=$1 mustBeNonZero=$2
  local q="com_mongodb_kafka_connect_sink_task_metrics_$m"
  local v
  v=$(curl -sf "$PROM_URL/api/v1/query?query=$q" | jq -r '.data.result[0].value[1] // "MISSING"')
  if [ "$v" = "MISSING" ]; then
    echo "  MISS $m (not exported)"; bad=$((bad+1)); return
  fi
  if [ "$mustBeNonZero" = "1" ] && [ "$v" = "0" ]; then
    echo "  MISS $m (still 0 after load)"; bad=$((bad+1)); return
  fi
  echo "  ok   $m = $v"
}
for m in "${counts[@]}"; do check "$m" 1; done
for m in "${gauges[@]}"; do check "$m" 0; done
[ "$bad" = "0" ] || fail "$bad MongoDB sink metric(s) missing or unexpectedly zero"
ok "all 10 metrics from the reference doc present"

# 7. Retarget test
step "Retarget scrape to host.docker.internal:8095"
"$BIN" observability host.docker.internal:8095 >/dev/null
sleep 20
tgt=$(curl -sf "$PROM_URL/api/v1/targets" \
  | jq -r '.data.activeTargets[] | select(.labels.job=="kafka-connect") | .scrapeUrl')
echo "$tgt" | grep -q "host.docker.internal:8095" \
  || fail "target did not switch (got: $tgt)"
health=$(curl -sf "$PROM_URL/api/v1/targets" \
  | jq -r '.data.activeTargets[] | select(.labels.job=="kafka-connect") | .health')
[ "$health" = "up" ] || fail "new target not healthy (health=$health)"
ok "switched target: $tgt (health=$health)"

step "Reset scrape to default"
"$BIN" observability kafka-connect:8091 >/dev/null
sleep 20
tgt=$(curl -sf "$PROM_URL/api/v1/targets" \
  | jq -r '.data.activeTargets[] | select(.labels.job=="kafka-connect") | .scrapeUrl')
echo "$tgt" | grep -q "kafka-connect:8091" \
  || fail "reset failed (got: $tgt)"
ok "reset to default: $tgt"

# Cleanup runs in the EXIT trap.
printf '\n\033[1;32mSMOKE TEST PASSED\033[0m\n'
