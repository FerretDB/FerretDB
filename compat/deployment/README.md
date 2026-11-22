# Run LibreChat and UniFi with FerretDB

This guide builds a local FerretDB image and starts a separate PostgreSQL/DocumentDB backend with Podman. Existing MongoDB data is copied and verified before changing an application's connection. Keep the application version unchanged during that move.

The application tests cover LibreChat v0.8.9, UniFi Network 10.6.109 and UniFi OS Server 5.1.42 with Network 10.6.109, using FerretDB 2.7.0, PostgreSQL 17 and DocumentDB 0.107. See [results](../TEST_RESULTS.md). The images below pin the tested backend. Validate other version combinations in an isolated stack before migration.

## Choose the source

Use the contribution branch in [this fork](https://github.com/tednv/FerretDB/tree/codex/librechat-unifi-compatibility), then record the commit you build. An upstream release can replace it after it includes the needed compatibility fixes. Check [the compatibility paths](../../UNIFI_COMPATIBILITY.md), native-statistics support and the pg_cron patch separately when selecting a release.

```sh
git clone --single-branch --branch codex/librechat-unifi-compatibility \
  https://github.com/tednv/FerretDB.git
cd FerretDB
```

Run these examples from the repository root in Bash on Linux or a Podman Linux VM. Windows users can run them inside WSL with access to their Podman engine. Use the same Podman engine and user for the application and database containers; a Docker container cannot resolve a name on a separate Podman network.

## Build the images

Podman, Git and Python 3 are required. Check out the intended commit, then build the images locally:

```sh
git rev-parse HEAD
podman build -f compat/deployment/Containerfile \
  -t localhost/ferretdb:application-compat .
podman build -t localhost/postgres-documentdb:cron-idle-fix compat/pg_cron
```

The first recipe compiles the source without embedding local Git history or host paths. The second applies [the pg_cron fix](../pg_cron/README.md). Keep the image IDs from `podman image inspect` in your private deployment record. Building the source checks compilation; the application test results above belong to the recorded tested builds.

## Create an isolated database stack

This example uses the `appdb` prefix and port 27018 on loopback. Use a different prefix and port for each independent application stack. Dedicated stacks keep application credentials and maintenance separate. Choose unused names before running the commands. Existing resources are a reason to inspect and reuse or choose different names, rather than delete them.

```sh
podman network exists appdb-network
podman container exists appdb-postgres
podman container exists appdb-gateway
podman volume exists appdb-pgdata
podman volume exists appdb-state
podman secret inspect appdb-password appdb-postgresql-url appdb-mongo-tools appdb-mongo-uri
```

These discovery commands should report absent resources (nonzero exit codes) for a fresh setup. Also check that loopback port 27018 is free. The example creates a bootstrap administrator for this dedicated backend. Do not reuse that credential across unrelated applications or expose the gateway publicly.

Create credentials directly in Podman's secret store, without printing a password or putting it in shell history:

```sh
# Set this to the application's existing MongoDB database name.
export APP_DATABASE=LibreChat
python3 - <<'PY'
import json, os, secrets, subprocess
from urllib.parse import quote

database = os.environ['APP_DATABASE']
if not database:
    raise SystemExit('Set APP_DATABASE to the existing application database name')
names = ['appdb-password', 'appdb-postgresql-url', 'appdb-mongo-tools', 'appdb-mongo-uri']
for name in names:
    if subprocess.run(['podman', 'secret', 'inspect', name],
                      stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0:
        raise SystemExit('A requested secret already exists; inspect before proceeding')
password = secrets.token_hex(32)
values = [password,
          'postgres://appdb_admin:' + password +
          '@appdb-postgres:5432/postgres?pool_min_conns=1&pool_max_conns=5',
          json.dumps({'uri': 'mongodb://appdb_admin@appdb-gateway:27017/'
                             '?authSource=admin&authMechanism=SCRAM-SHA-256',
                      'password': password}),
          'mongodb://appdb_admin:' + password + '@appdb-gateway:27017/' +
          quote(database, safe='') +
          '?authSource=admin&authMechanism=SCRAM-SHA-256&maxPoolSize=8']
for name, value in zip(names, values):
    subprocess.run(['podman', 'secret', 'create', name, '-'],
                   input=value.encode(), check=True, stdout=subprocess.DEVNULL)
PY

podman network create --internal --label io.ferretdb.setup=appdb appdb-network
podman volume create --label io.ferretdb.setup=appdb appdb-pgdata
podman volume create --label io.ferretdb.setup=appdb appdb-state

podman run -d --name appdb-postgres --network appdb-network \
  --label io.ferretdb.setup=appdb --restart unless-stopped \
  --memory 1g --cpus 1 --shm-size 256m --health-cmd=none \
  --log-opt=max-size=5mb \
  --secret appdb-password \
  -e POSTGRES_PASSWORD_FILE=/run/secrets/appdb-password \
  -e POSTGRES_USER=appdb_admin -e POSTGRES_DB=postgres \
  -v appdb-pgdata:/var/lib/postgresql/data \
  localhost/postgres-documentdb:cron-idle-fix \
  -c shared_buffers=64MB -c max_connections=20 -c work_mem=4MB \
  -c maintenance_work_mem=64MB -c max_wal_size=256MB \
  -c log_min_messages=warning -c log_checkpoints=off \
  -c cron.log_run=off -c cron.log_statement=off
```

Wait for initialization, then confirm the extension is installed:

```sh
podman exec appdb-postgres pg_isready -U appdb_admin -d postgres
podman exec appdb-postgres psql -X -v ON_ERROR_STOP=1 -U appdb_admin -d postgres \
  -c "SELECT extname, extversion FROM pg_extension WHERE extname LIKE 'documentdb%';"
```

Start the gateway after PostgreSQL is ready. Data and gateway state use named volumes. PostgreSQL has no published host port.

```sh
podman run -d --name appdb-gateway --network appdb-network \
  --label io.ferretdb.setup=appdb --restart unless-stopped \
  --memory 256m --cpus 1 --health-cmd=none --log-opt=max-size=5mb \
  -p 127.0.0.1:27018:27017 \
  --secret appdb-postgresql-url,type=env,target=FERRETDB_POSTGRESQL_URL \
  -e FERRETDB_AUTH=true -e FERRETDB_UNIFI_COMPAT=true \
  -e FERRETDB_MAX_PERSISTENT_CURSORS=8 \
  -e FERRETDB_LISTEN_ADDR=0.0.0.0:27017 -e FERRETDB_STATE_DIR=/state \
  -e FERRETDB_TELEMETRY=disable -e FERRETDB_LOG_LEVEL=WARN \
  -v appdb-state:/state localhost/ferretdb:application-compat

podman exec appdb-gateway /ferretdb ping
```

Increase connection and memory limits together when a workload needs them. Keep `fsync`, `full_page_writes` and `synchronous_commit` enabled. The quiet pg_cron settings disable run-history logging while TTL and index jobs continue running.

## Copy existing MongoDB data

1. Inventory the application's actual database names, collections, indexes and options, source MongoDB version, tool versions and persistent application files. Keep this record private.
2. Take an application backup plus a consistent database backup. Stop all application writers for the final dump, or use an explicitly verified consistent snapshot procedure. A dump of an active standalone server can span different points in time.
3. Use `mongodump` compatible with the source version. Preserve BSON and metadata. Export the application's databases, including UniFi's statistics and audit databases when present. Do not copy MongoDB's internal `admin`, `local` or `config` databases into FerretDB as application data.
4. Restore into the new empty backend using compatible `mongorestore` with `--stopOnError`. Pass credentials in a protected configuration file or mounted secret. Restore indexes and collection options, then compare them with the source. If a tool needs a separate index replay, preserve and verify the full index definitions; `--noIndexRestore` by itself leaves that work unfinished.
5. Compare document counts, BSON types and representative content, plus index keys/options and TTL definitions. Account explicitly for any expired records. Rehearse application startup and restart against the copied data before the final cutover.

The `appdb-mongo-tools` secret is a Database Tools configuration containing the target URI and password. A tools container on `appdb-network` can mount it with `--secret appdb-mongo-tools` and pass `--config=/run/secrets/appdb-mongo-tools`. Mount the dump read-only. Tool images need their own verified version/digest; a current tool is not necessarily compatible with an old MongoDB server.

Back up LibreChat's uploaded files, configuration, encryption/signing keys and any separately stored search/vector data too. Back up UniFi's full configuration directory, keys and certificates alongside every application database. Keep restores isolated from devices and external actions until the actual cutover.

## Point LibreChat at the new backend

On the shared Podman network, the connection has this shape:

```text
MONGO_URI=mongodb://appdb_admin:<password>@appdb-gateway:27017/<existing-database>?authSource=admin&authMechanism=SCRAM-SHA-256&maxPoolSize=8
```

Replace the placeholders through a private secret/configuration workflow. The generated password is hexadecimal, so it needs no URI escaping. Other passwords need percent encoding. For an application running on the host, use `127.0.0.1:27018`. For an application container, attach it to `appdb-network` and use `appdb-gateway:27017`; its own loopback is not the database host.

For a directly managed Podman LibreChat container, the generated `appdb-mongo-uri` secret supplies the complete URI. Add `--secret appdb-mongo-uri,type=env,target=MONGO_URI` and `--network appdb-network` to its saved container definition, preserving its other settings. If it already needs another network, retain that network too. A Compose deployment needs the equivalent secret/override supported by its Compose provider. Do not print the full resolved configuration in a shared log.

[LibreChat uses `MONGO_URI`](https://www.librechat.ai/docs/configuration/dotenv). Its Compose file may set this explicitly and override `.env`, so check the effective configuration privately. Update the service override or secret binding that actually supplies the value. Preserve the original override for rollback. Change/recreate only the application service once the final data copy passes verification. Keep MongoDB stopped and its storage intact after a successful cutover; remove any dependency that would automatically restart it.

Verify existing login, conversation history, streamed replies, follow-ups, edits, imports, presets/tags/bookmarks and uploaded-file access. Restart LibreChat and confirm history remains. Verify scheduled expiry against disposable fixtures with future and missing-expiry controls.

## Point UniFi Network at the new backend

For [LinuxServer's Network application](https://docs.linuxserver.io/images/docker-unifi-network-application/), preserve the existing `MONGO_DBNAME` and use these settings with the dedicated backend:

```text
MONGO_HOST=appdb-gateway
MONGO_PORT=27017
MONGO_USER=appdb_admin
MONGO_AUTHSOURCE=admin
FILE__MONGO_PASS=/run/secrets/appdb-password
MONGO_DBNAME=<existing-database>
```

Mount `appdb-password` into the application and attach it to `appdb-network`. Keep its configuration volume, application image, ports, controller address, site/device identities and inform settings unchanged. If the runtime is Docker, create the backend in that engine or configure a deliberately restricted reachable endpoint; the Podman container name is not a cross-engine DNS name.

Changing the database endpoint with the full controller data preserved normally retains adoption. Confirm existing devices inform successfully before accepting the cutover. Check sites, devices, clients, provisioning, rules, alerts, reports, backups, restore and retention. A copied controller used for rehearsal should not receive real device traffic.

## UniFi OS and native statistics

Native UniFi OS Server uses its own service/package configuration rather than LinuxServer's environment variables. Inventory that version's external MongoDB settings and service definitions before changing it. Preserve its native PostgreSQL data independently of the MongoDB migration.

[The optional bridge](../unifi/README.md) handles MODERN selection persistence, native-statistics backup/restore and the tested OS startup path. It checks the exact application build. Install it only for a matching build, following that guide, and verify MODERN and statistics survive both a Network restart and an OS restart. An encrypted `.unf` round trip should recover all four native tables. A MongoDB dump alone does not back up native statistics.

Build the bridge from a user-supplied application installation. Keep vendor software out of redistributable images. Include ASM's BSD notice with compiled bridge distributions and the JDBC driver's notice when redistributing that driver. The current bridge builder needs notice packaging before publishing a binary release.

## Verify and keep rollback ready

Check application and backend logs privately. Verify PostgreSQL durability settings, active maintenance jobs, scheduled TTL and background indexing. Check deadlocks, lock waits, idle transactions and settled CPU/disk writes. [The recorded idle-write results](../IDLE_WRITES.md) show what was measured; active application traffic and backups still produce legitimate writes.

Keep the original MongoDB image, volumes, backup and connection configuration. If rollback is required, stop application writers first. Decide how to preserve writes made after cutover before reconnecting to the old database, which is now stale. Do not run both copies as independent writable controllers.

For FerretDB backups, use a verified whole-cluster backup or [the DocumentDB-aware logical recovery procedure](../DOCUMENTDB_RECOVERY.md). A plain `pg_dump` omits catalog data needed by this tested extension. Keep native statistics and application files in the backup plan. Test recovery into a separate stack.

## Stop or clean up the example

Stopping preserves data:

```sh
podman stop appdb-gateway appdb-postgres
```

Restart PostgreSQL, wait for readiness, then start the gateway. Reconnect/start the application afterward.

To remove a discarded example, first verify the exact resources carry `io.ferretdb.setup=appdb`, stop and detach any application using them, and confirm the volumes contain no data you need. Remove only the two named containers and network. Volume and secret deletion is a separate, explicit decision after backups and rollback are no longer needed. Avoid `podman system prune`, broad name matches and Compose `down -v` as cleanup shortcuts.

## Use an agent

Give the agent this instruction from your own deployment workspace:

> Read `compat/deployment/README.md` and `compat/deployment/AGENTS.md` in my FerretDB checkout. Inspect my local infrastructure read-only, prepare an isolated target, and produce a concrete migration and rollback plan. Preserve existing data and UniFi adoption. Ask before the production cutover. Keep secrets, backups and logs local. Use the contribution build or a verified upstream release that includes the required fixes.

An `AGENTS.md` nested in this directory is not automatically discovered for an unrelated application checkout. Refer to it explicitly, or copy its applicable instructions into that workspace after reviewing its existing instructions.
