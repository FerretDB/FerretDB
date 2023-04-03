---
sidebar_position: 1
---

# Configuration flags

FerretDB provides numerous configuration flags you can customize to suit your needs and environment.
You can always see the complete list by using `--help` flag.
To make user experience cloud native, every flag has its environment variable equivalent.
There is no configuration file.

:::info
Some default values are overridden in [our Docker image](quickstart-guide/docker.md).
:::

| Flag                    | Description                                                               | Environment Variable             | Default Value                                |
| ----------------------- | ------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------- |
| `version`               | Print version to stdout and exit                                          | `FERRETDB_VERSION`               |                                              |
| `log-level`             | Log level: `debug`, `info`, `warn`, `error`                               | `FERRETDB_LOG_LEVEL`             | `info`                                       |
| `log-uuid`              | Add instance UUID to all log messages                                     | `FERRETDB_LOG_UUID`              |                                              |
| `metrics-uuid`          | Add instance UUID to all metrics                                          | `FERRETDB_METRICS_UUID`          |                                              |
| `state-dir`             | Path to the FerretDB state directory                                      | `FERRETDB_STATE_DIR`             | `.`<br />(`/state` for Docker)               |
| `debug-addr`            | Debug address for /debug/metrics, /debug/pprof, and similar HTTP handlers | `FERRETDB_DEBUG_ADDR`            | `127.0.0.1:8088`<br />(`:8088` for Docker)   |
| **Listeners**           |                                                                           |                                  |                                              |
| `listen-addr`           | FerretDB address for incoming TCP connections                             | `FERRETDB_LISTEN_ADDR`           | `127.0.0.1:27017`<br />(`:27017` for Docker) |
| `listen-unix`           | FerretDB Unix domain socket path. If empty - Unix socket is disabled      | `FERRETDB_LISTEN_UNIX`           |                                              |
| **Handlers**            |                                                                           |                                  |                                              |
| `handler`               | FerretDB backend handler: 'dummy', 'pg', 'tigris'                         | `FERRETDB_HANDLER`               | `pg`                                         |
| `postgresql-url`        | PostgreSQL URL for `pg` handler                                           | `FERRETDB_POSTGRESQL_URL`        | `postgres://127.0.0.1:5432/ferretdb`         |
| `tigris-url`            | Tigris URL for 'tigris' handler                                           | `FERRETDB_TIGRIS_URL`            | `127.0.0.1:8081`                             |
| `tigris-client-id`      | [Tigris Client ID][tigris-docs-auth]                                      | `FERRETDB_TIGRIS_CLIENT_ID`      |                                              |
| `tigris-client-secret`  | [Tigris Client secret][tigris-docs-auth]                                  | `FERRETDB_TIGRIS_CLIENT_SECRET`  |                                              |
| **TLS**                 |                                                                           |                                  |                                              |
| `listen-tls`            | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS`            |                                              |
| `listen-tls-cert-file`  | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_CERT_FILE`  |                                              |
| `listen-tls-key-file`   | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_KEY_FILE`   |                                              |
| `listen-tls-ca-file`    | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_CA_FILE`    |                                              |
| **Operation Modes**     |                                                                           |                                  |                                              |
| `mode`                  | See [Operation modes](operation-modes.md)                                 | `FERRETDB_MODE`                  | `normal`                                     |
| `proxy-addr`            | See [Operation modes/Proxy](operation-modes.md#proxy)                     | `FERRETDB_PROXY_ADDR`            |                                              |
| **Telemetry**           |                                                                           |                                  |                                              |
| `telemetry`             | See [Configure telemetry](telemetry.md#configure-telemetry)               | `FERRETDB_TELEMETRY`             | `undecided`                                  |
| **Testing**             |                                                                           |                                  |                                              |
| `test-disable-pushdown` | Disable pushing down queries to the backend (to only filter on FerretDB)  | `FERRETDB_TEST_DISABLE_PUSHDOWN` | `false`                                      |

[tigris-docs-auth]: https://www.tigrisdata.com/docs/sdkstools/golang/getting-started/
[securing-with-tls]: /security#securing-connections-with-tls
