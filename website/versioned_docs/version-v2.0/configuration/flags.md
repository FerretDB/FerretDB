---
sidebar_position: 1
---

# Configuration flags

FerretDB provides numerous configuration flags you can customize to suit your needs and environment.
You can always see the complete list by using `--help` flag.
To make user experience cloud native, every flag has its environment variable equivalent.
There is no configuration file.

:::info
Some default values are overridden in [our Docker image](../installation/ferretdb/docker.md).
:::

<!-- Keep order in sync with the `--help` output -->

<!-- For <br /> -->
<!-- markdownlint-capture -->
<!-- markdownlint-disable MD033 -->

## PostgreSQL with DocumentDB extension

| Flag               | Description               | Environment Variable      | Default Value                        |
| ------------------ | ------------------------- | ------------------------- | ------------------------------------ |
| `--postgresql-url` | PostgreSQL connection URL | `FERRETDB_POSTGRESQL_URL` | `postgres://127.0.0.1:5432/postgres` |

FerretDB uses [pgx v5](https://github.com/jackc/pgx) library for connecting to PostgreSQL.
Supported URL parameters are documented there:

- https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#ParseConfig
- https://pkg.go.dev/github.com/jackc/pgx/v5#ParseConfig
- https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ParseConfig

Additionally:

- `pool_max_conns` parameter is set to 50 if it is unset in the URL;
- `application_name` is always set to "FerretDB";
- `timezone` is always set to "UTC".

## Interfaces

| Flag                     | Description                                                                                                                      | Environment Variable            | Default Value                                |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- | -------------------------------------------- |
| `--listen-addr`          | Listen TCP address for MongoDB protocol<br />(set to empty value or `-` to disable)                                              | `FERRETDB_LISTEN_ADDR`          | `127.0.0.1:27017`<br />(`:27017` for Docker) |
| `--listen-unix`          | Listen Unix domain socket path for MongoDB protocol<br />(set to empty value or `-` to disable)                                  | `FERRETDB_LISTEN_UNIX`          |                                              |
| `--listen-tls`           | Listen TLS address for MongoDB protocol (see [here](../security/tls-connections.md))<br />(set to empty value or `-` to disable) | `FERRETDB_LISTEN_TLS`           |                                              |
| `--listen-tls-cert-file` | TLS cert file path                                                                                                               | `FERRETDB_LISTEN_TLS_CERT_FILE` |                                              |
| `--listen-tls-key-file`  | TLS key file path                                                                                                                | `FERRETDB_LISTEN_TLS_KEY_FILE`  |                                              |
| `--listen-tls-ca-file`   | TLS CA file path                                                                                                                 | `FERRETDB_LISTEN_TLS_CA_FILE`   |                                              |
| `--listen-data-api-addr` | Listen TCP address for HTTP Data API<br />(set to empty value or `-` to disable)                                                 | `FERRETDB_LISTEN_DATA_API_ADDR` |                                              |
| `--proxy-addr`           | Proxy address for non-normal [operation mode](operation-modes.md)                                                                | `FERRETDB_PROXY_ADDR`           |                                              |
| `--proxy-tls-cert-file`  | Proxy TLS cert file path                                                                                                         | `FERRETDB_PROXY_TLS_CERT_FILE`  |                                              |
| `--proxy-tls-key-file`   | Proxy TLS key file path                                                                                                          | `FERRETDB_PROXY_TLS_KEY_FILE`   |                                              |
| `--proxy-tls-ca-file`    | Proxy TLS CA file path                                                                                                           | `FERRETDB_PROXY_TLS_CA_FILE`    |                                              |
| `--debug-addr`           | Listen address for HTTP handlers for metrics, pprof, etc<br />(set to empty value or `-` to disable)                             | `FERRETDB_DEBUG_ADDR`           | `127.0.0.1:8088`<br />(`:8088` for Docker)   |

## Miscellaneous

| Flag                  | Description                                                                                                                 | Environment Variable       | Default Value                  |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------- | -------------------------- | ------------------------------ |
| `-h`, `--help`        | Show context-sensitive help                                                                                                 |                            |                                |
| `--version`           | Print version to stdout and exit                                                                                            |                            |                                |
| `--mode`              | [Operation mode](operation-modes.md)                                                                                        | `FERRETDB_MODE`            | `normal`                       |
| `--state-dir`         | Path to the FerretDB state directory                                                                                        | `FERRETDB_STATE_DIR`       | `.`<br />(`/state` for Docker) |
| `--[no-]auth`         | [Enable authentication](../security/authentication.md)                                                                      | `FERRETDB_AUTH`            | enabled                        |
| `--log-level`         | Log level: 'debug', 'info', 'warn', 'error'                                                                                 | `FERRETDB_LOG_LEVEL`       | `info`                         |
| `--[no-]log-uuid`     | Add instance UUID to all log messages                                                                                       | `FERRETDB_LOG_UUID`        | disabled                       |
| `--[no-]metrics-uuid` | Add instance UUID to all metrics                                                                                            | `FERRETDB_METRICS_UUID`    | disabled                       |
| `--otel-traces-url`   | OpenTelemetry OTLP/HTTP traces endpoint URL (e.g. `http://host:4318/v1/traces`)<br />(set to empty value or `-` to disable) | `FERRETDB_OTEL_TRACES_URL` | disabled                       |
| `--telemetry`         | Enable or disable [basic telemetry](telemetry.md)                                                                           | `FERRETDB_TELEMETRY`       | `undecided`                    |

<!-- Do not document `--dev-XXX` flags -->

<!-- markdownlint-restore -->
