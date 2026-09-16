# AGENTS.md

Klaunch is a single-binary Cobra CLI (`main.go`) that stands up a Kafka + MongoDB Kafka Connect stack via Docker Compose for reproduction cases.

## Layout gotchas
- All Go code is `package main` at the repo root. `test/` does NOT exist — ignore any claim in `CLAUDE.md` / `TESTS.md` about it. `test_utils.go` is a plain source file (not `_test`) used by the tests, also `package main`.
- `internal/`, `pkg/`, `old/`, `docker/` are unused/legacy. Nothing imports `internal/*`. Don't refactor into them without checking.
- Filename typo is preserved for git history: `check_mogodb_running.go` (missing "n"). Function is `check_mongodb_running`.

## Build / run
- Build: `make build` -> `build/klaunch`. Release: `make release`.
- Linux binary can NOT be cross-compiled from macOS (librdkafka needs CGO). `make build-linux` uses `CGO_ENABLED=0` and fails. Build on Linux; shipped `klaunch_linux` is Ubuntu 24.04.
- CLI subcommands: `start [version] | stop | create | delete [all|connectors|topics] | show <components|messages> | logs`.
- `start` opens Docker Desktop via `open -a Docker` (macOS-only, silently no-op on Linux), then tries `docker-compose -p klaunch up -d`, falling back to `docker compose -p klaunch up -d`. Always use `-p klaunch` so `stop`/`status-infra` match.

## Tests
- Everything runs at the root package. Use `./run_tests.sh {unit|integration|infrastructure|benchmarks|static|all}` or `make test|unit-tests|benchmarks|coverage`.
- Integration + infrastructure tests need Docker; the script auto-skips infra when Docker is absent.
- Direct: `go test -v .` (NOT `./test`).

## Runtime side effects to warn about
- `start` calls `check_mongodb_running`, which expects a local 3-node replica set on 27017/27018/27019, may edit `/etc/hosts` to add `127.0.0.1 host.docker.internal` (needs sudo to actually write), and runs `mongosh` to `rs.reconfig` members to `host.docker.internal:2701[7-9]`. Do not run against a shared MongoDB.
- `check_connector_updates` overwrites `MONGO_KAFKA_CONNECT_VERSION` in `.env` and downloads the JAR to `volumes/mongo-kafka-connect-<ver>-all.jar` (compose mounts this exact path).

## Configs
- `create` menu lists `./case_configs/*.json`. Its "default" fallback points at `./case_configs/default_topic.json`, which does not ship — the real templates live in `example_configs/`. Put new picker options in `case_configs/`.

## Hardcoded endpoints (change together, or nothing works)
- Host side: brokers `localhost:9091,9092,9093`, Connect REST `http://localhost:8083`, Schema Registry `http://localhost:8081`, CMAK `http://localhost:9000` (needs one-time UI setup: cluster `kafka-connect`, ZK `zookeeper1`).
- Container side (used via `docker exec kafka-connect ...`): brokers `kafka1:19091,kafka2:19092,kafka3:19093`.
- Consumer in `list_messages.go` seeks to `OffsetTail(10)` on assignment — only the last 10 messages per partition are shown.

## Known quirks
- `main.go` `delete topics` branch calls the same interactive connector-delete path as `delete connectors` (bug); use `delete all` for a full wipe.
- Compose file references `$PWD/volumes/...` — always invoke Compose from repo root.

## Instruction sources
- `README.md` — user-facing usage + release process (verified).
- `CLAUDE.md`, `TESTS.md` — historical notes; the `test/` directory they describe was never created. Trust code over these two.
