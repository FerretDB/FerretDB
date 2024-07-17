---
sidebar_position: 3
description: Observability
---

# Observability

## Logging

FerretDB writes structured logs to the standard error (`stderr`) stream.
The most recent entries are also available via `getLog` command.

:::note

<!-- https://github.com/FerretDB/FerretDB/issues/3421 -->

Structured log format is not stable yet; field names and formatting of values might change in minor releases.
:::

FerretDB provides the following log formats:

<!-- https://github.com/FerretDB/FerretDB/issues/4438 -->

- `console` is a human-readable format with optional colors;
- `text` is machine-readable [logfmt](https://brandur.org/logfmt)-like format
  (powered by [Go's `slog.TextHandler`](https://pkg.go.dev/log/slog#TextHandler));
- `json` if machine-readable JSON format
  (powered by [Go's `slog.JSONHandler`](https://pkg.go.dev/log/slog#JSONHandler)).

There are four logging levels:

<!-- https://github.com/FerretDB/FerretDB/issues/4439 -->

- `error` is used for errors that can't be handled gracefully
  and typically result in client connection being closed;
- `warn` is used for errors that can be handled gracefully
  and typically result in an error being returned to the client (without closing the connection);
- `info` is used for various information messages;
- `debug` should only be used for debugging.

The default level is `info`, except for [debug builds](https://pkg.go.dev/github.com/FerretDB/FerretDB/build/version#hdr-Debug_builds) that default to `debug`.

:::caution
`debug`-level messages include complete query and response bodies, full error messages, authentication credentials,
and other sensitive information.

Since logs are often retained by the infrastructure
(and FerretDB itself makes recent entries available via the `getLog` command),
that poses a security risk.
Additionally, writing out a significantly larger number of log messages affects FerretDB performance.
For those reasons, the `debug` level should not be enabled in production environments.
:::

The format and level can be adjusted by [configuration flags](flags.md#miscellaneous).

### Docker logs

If Docker was launched with [our quick local setup with Docker Compose](../quickstart-guide/docker.md#postgresql-setup-with-docker-compose),
the following command can be used to fetch the logs.

```sh
docker compose logs ferretdb
```

Otherwise, you can check a list of running Docker containers with `docker ps`
and get logs with `docker logs`:

```sh
$ docker ps
CONTAINER ID   IMAGE                       COMMAND                  CREATED              STATUS          PORTS                                           NAMES
13db4c8800d3   postgres                    "docker-entrypoint.s…"   About a minute ago   Up 59 seconds   5432/tcp                                        my-postgres
44fe6f4c3527   ghcr.io/ferretdb/ferretdb   "/ferretdb"              About a minute ago   Up 59 seconds   8088/tcp, 27018/tcp, 0.0.0.0:27017->27017/tcp   my-ferretdb

$ docker logs my-ferretdb
```

## Debug handler

FerretDB exposes various HTTP endpoints with the debug handler on `http://127.0.0.1:8088/debug/` by default.
The host and port can be changed with [`--debug-addr` flag](flags.md#interfaces).

### Metrics

FerretDB exposes metrics in Prometheus format on the `/debug/metrics` endpoint.
There is no need to use an external exporter.

<!-- https://github.com/FerretDB/FerretDB/issues/3420 -->

:::note
The set of metrics is not stable yet; metric and label names and value formatting might change in minor releases.
:::

### Probes

FerretDB exposes the following probes that can be used for
[Kubernetes health checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
and similar use cases.
They return HTTP code 2xx if a probe is successful and 5xx otherwise.
The response body is always empty, but additional information may be present in logs.

- `/debug/livez` is a liveness probe.
  It succeeds if FerretDB is ready to accept new connections from MongoDB protocol clients.
  It does not check if the connection with the backend can be established or authenticated.
  An error response or timeout indicates (after a small initial startup delay) a serious problem.
  Generally, FerretDB should be restarted in that case.
  Additionally, the error is returned during the FerretDB shutdown while it waits for established connections to be closed.
- `/debug/readyz` is a readiness probe.
  It succeeds if the liveness probe succeeds.
  Additionally, if [new authentication](../security/authentication.md) is enabled and setup credentials are provided,
  it checks that connection with the backend can be established and authenticated
  by sending MongoDB `ping` command to FerretDB.
  An error response or timeout indicates a problem with the backend or configuration.
