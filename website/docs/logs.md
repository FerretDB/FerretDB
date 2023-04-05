---
sidebar_position: 10
description: Logs
---

# Logs

## Docker logs

Logs from FerretDB running on Docker can be accessed through the container.

If Docker was launched with the [quick start](quickstart-guide/docker.md#setup-with-docker-compose),
the following command can be used to fetch the logs.

```shell
docker logs -name ferretdb-docker-ferretdb-1
```

## Binary executable logs

FerretDB generates logs to standard output `stdout` and standard error `stderr` streams
but does not retain them.
Refer to the [flags](configuration/flags.md#miscellaneous) to adjust the log level.
