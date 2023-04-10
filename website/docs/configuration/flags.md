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

<!-- Keep order in sync with the `--help` output -->

<!-- For <br /> -->
<!-- markdownlint-disable MD033 -->

## General

| Flag           | Description                          | Environment Variable | Default Value                  |
| -------------- | ------------------------------------ | -------------------- | ------------------------------ |
| `-h`, `--help` | Show context-sensitive help          |                      | false                          |
| `--version`    | Print version to stdout and exit     |                      | false                          |
| `--handler`    | Backend handler                      | `FERRETDB_HANDLER`   | `pg` (PostgreSQL)              |
| `--mode`       | [Operation mode](operation-modes.md) | `FERRETDB_MODE`      | `normal`                       |
| `--state-dir`  | Path to the FerretDB state directory | `FERRETDB_STATE_DIR` | `.`<br />(`/state` for Docker) |

## Interfaces

| Flag                     | Description                                              | Environment Variable            | Default Value                                |
| ------------------------ | -------------------------------------------------------- | ------------------------------- | -------------------------------------------- |
| `--listen-addr`          | Listen TCP address                                       | `FERRETDB_LISTEN_ADDR`          | `127.0.0.1:27017`<br />(`:27017` for Docker) |
| `--listen-unix`          | Listen Unix domain socket path                           | `FERRETDB_LISTEN_UNIX`          |                                              |
| `--listen-tls`           | Listen TLS address (see [here](../security.md))          | `FERRETDB_LISTEN_TLS`           |                                              |
| `--listen-tls-cert-file` | TLS cert file path                                       | `FERRETDB_LISTEN_TLS_CERT_FILE` |                                              |
| `--listen-tls-key-file`  | TLS key file path                                        | `FERRETDB_LISTEN_TLS_KEY_FILE`  |                                              |
| `--listen-tls-ca-file`   | TLS CA file path                                         | `FERRETDB_LISTEN_TLS_CA_FILE`   |                                              |
| `--proxy-addr`           | Proxy address                                            | `FERRETDB_PROXY_ADDR`           |                                              |
| `--debug-addr`           | Listen address for HTTP handlers for metrics, pprof, etc | `FERRETDB_DEBUG_ADDR`           | `127.0.0.1:8088`<br />(`:8088` for Docker)   |

## Backend handlers

### PostgreSQL

PostgreSQL backend can be enabled by `--handler=pg` flag or `FERRETDB_HANDLER=pg` environment variable.

| Flag               | Description                     | Environment Variable      | Default Value                        |
| ------------------ | ------------------------------- | ------------------------- | ------------------------------------ |
| `--postgresql-url` | PostgreSQL URL for 'pg' handler | `FERRETDB_POSTGRESQL_URL` | `postgres://127.0.0.1:5432/ferretdb` |

### Tigris (beta)

Tigris backend can be enabled by `--handler=tigris` flag or `FERRETDB_HANDLER=tigris` environment variable.

| Flag                     | Description                     | Environment Variable            | Default Value    |
| ------------------------ | ------------------------------- | ------------------------------- | ---------------- |
| `--tigris-url`           | Tigris URL for 'tigris' handler | `FERRETDB_TIGRIS_URL`           | `127.0.0.1:8081` |
| `--tigris-client-id`     | Tigris Client ID                | `FERRETDB_TIGRIS_CLIENT_ID`     |                  |
| `--tigris-client-secret` | Tigris Client secret            | `FERRETDB_TIGRIS_CLIENT_SECRET` |                  |

## Miscellaneous

| Flag                  | Description                                       | Environment Variable    | Default Value |
| --------------------- | ------------------------------------------------- | ----------------------- | ------------- |
| `--log-level`         | Log level: 'debug', 'info', 'warn', 'error'       | `FERRETDB_LOG_LEVEL`    | `info`        |
| `--[no-]log-uuid`     | Add instance UUID to all log messages             | `FERRETDB_LOG_UUID`     |               |
| `--[no-]metrics-uuid` | Add instance UUID to all metrics                  | `FERRETDB_METRICS_UUID` |               |
| `--telemetry`         | Enable or disable [basic telemetry](telemetry.md) | `FERRETDB_TELEMETRY`    | `undecided`   |

<!-- Do not document `--test-XXX` flags here -->

<!-- markdownlint-enable MD033 -->
