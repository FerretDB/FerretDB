---
sidebar_position: 3
description: Observability
---

# Observability

## Logging

The log level and format can be adjusted by [configuration flags](flags.md#miscellaneous).

Please note that the structured log format is not stable yet; field names and formatting of values might change in minor releases.

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

### Binary executable logs

FerretDB writes logs to the standard error (`stderr`) stream.

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
They return HTTP code 200 if a probe is successful and 500 otherwise.
The response body is always empty, but additional information may be present in logs.

- `/debug/livez` is a liveness probe.
  It succeeds if FerretDB is ready to accept connections from MongoDB protocol clients.
  It does not check if the connection with the backend can be established or authenticated.
  An error response or timeout indicates (after a small initial startup delay) a serious problem.
  Generally, FerretDB should be restarted in that case.
- `/debug/readyz` is a readiness probe.
  It succeeds if the liveness probe succeeds.
  Additionally, if [new authentication](../security/authentication.md) is enabled and setup credentials are provided,
  it checks that connection with the backend can be established and authenticated
  by sending MongoDB `ping` command to FerretDB.
  An error response or timeout indicates a problem with the backend or configuration.
