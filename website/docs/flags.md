---
---

# Configuration flags and variables

FerretDB provides numerous configuration flags that you can customize to suit your needs and environment.
You can always access them by using the `--help` flag.
To make user experience cloud native, every flag has its environment variable equivalent.

| Flag                   | Description                                                               | Env Variable                    | Default Value                          |
| ---------------------- | ------------------------------------------------------------------------- | ------------------------------- | -------------------------------------- |
| `version`              | Print version to stdout and exit                                          | `FERRETDB_VERSION`              |                                        |
| `log-level`            | Log level: `debug`, `info`, `warn`, `error`                               | `FERRETDB_LOG_LEVEL`            | `"info"`                               |
| `log-uuid`             | Add instance UUID to all log messages                                     | `FERRETDB_LOG_UUID`             |                                        |
| `metrics-uuid`         | Add instance UUID to all metrics                                          | `FERRETDB_METRICS_UUID`         |                                        |
| `state-dir`            | Path to the FerretDB state directory                                      | `FERRETDB_STATE_DIR`            | `"."`                                  |
| `debug-addr`           | Debug address for /debug/metrics, /debug/pprof, and similar HTTP handlers | `FERRETDB_DEBUG_ADDR`           | `"127.0.0.1:8088"`                     |
| **Listeners**          |                                                                           |                                 |                                        |
| `listen-addr`          | FerretDB address for incoming TCP connections                             | `FERRETDB_LISTEN_ADDR`          | `"127.0.0.1:27017"`                    |
| `listen-unix`          | FerretDB Unix domain socket path. If empty - Unix socket is disabled      | `FERRETDB_LISTEN_UNIX`          |                                        |
| **Handlers**           |                                                                           |                                 |                                        |
| `handler`              | FerretDB backend handler: 'dummy', 'pg', 'tigris'                         | `FERRETDB_HANDLER`              | `"pg"`                                 |
| `postgresql-url`       | PostgreSQL URL for `pg` handler                                           | `FERRETDB_POSTGRESQL_URL`       | `"postgres://127.0.0.1:5432/ferretdb"` |
| `tigris-url`           | Tigris URL for 'tigris' handler                                           | `FERRETDB_TIGRIS_URL`           | `"127.0.0.1:8081"`                     |
| `tigris-client-id`     | [Tigris Client ID][tigris-docs-auth]                                      | `FERRETDB_TIGRIS_CLIENT_ID`     |                                        |
| `tigris-client-secret` | [Tigris Client secret][tigris-docs-auth]                                  | `ferretdb_tigris_client_secret` |                                        |
| `tigris-token`         | [Tigris token][tigris-docs-auth]                                          | `FERRETDB_TIGRIS_TOKEN`         |                                        |
| **TLS**                |                                                                           |                                 |                                        |
| `listen-tls`           | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS`           |                                        |
| `listen-tls-cert-file` | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_CERT_FILE` |                                        |
| `listen-tls-key-file`  | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_KEY_FILE`  |                                        |
| `listen-tls-ca-file`   | See [Securing connections with TLS][securing-with-tls]                    | `FERRETDB_LISTEN_TLS_CA_FILE`   |                                        |
| **Operation Modes**    |                                                                           |                                 |                                        |
| `mode`                 | See [Operation modes](/operation_modes.md)                                | `FERRETDB_MODE`                 | `"normal"`                             |
| `proxy-addr`           | See [Operation modes/Proxy](/operation_modes.md#proxy)                    | `FERRETDB_PROXY_ADDR`           |                                        |
| **Telemetry**          |                                                                           |                                 |                                        |
| `telemetry`            | See [Configure telemetry](/telemetry.md#configure-telemetry)              | `FERRETDB_TELEMETRY`            | `undecided`                            |

[tigris-docs-auth]: https://docs.tigrisdata.com/apidocs/#tag/Authentication/operation/Auth_GetAccessToken
[securing-with-tls]: /security#securing-connections-with-tls
