---
sidebar_position: 3
description: Logging
---

# Logging

## Docker logs

Logs from FerretDB running on Docker can be accessed through the container.

If Docker was launched with [our quick local setup with Docker Compose](../quickstart-guide/docker.md#setup-with-docker-compose),
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

## Binary executable logs

FerretDB writes logs to the standard error (`stderr`) stream but does not retain them.
Refer to the [flags](flags.md#miscellaneous) to adjust the log level.
