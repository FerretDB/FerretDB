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
44fe6f4c3527   ghcr.io/ferretdb/ferretdb   "/ferretdb"              About a minute ago   Up 59 seconds   8080/tcp, 27018/tcp, 0.0.0.0:27017->27017/tcp   my-ferretdb

$ docker logs my-ferretdb
```

### Binary executable logs

FerretDB writes logs to the standard error (`stderr`) stream.

## Metrics

FerretDB exposes metrics in Prometheus format on the debug handler on `http://127.0.0.1:8088/debug/metrics` by default.
There is no need to use an external exporter.
The host and port can be changed with [`--debug-addr` flag](flags.md#interfaces).

Please note that the set of metrics is not stable yet; metric and label names and formatting of values might change in minor releases.
