---
sidebar_position: 1
description: How to set up FerretDB using Docker
---

# Docker

We provide three Docker images for various deployments:

- **an evaluation image**: for quick testing and experiments
- **a development image**: for debugging problems
- **a production image**: for all other cases

An evaluation image is documented [separately](../evaluation.md).
The rest are covered below.

All Docker images include a [`HEALTHCHECK` instruction](https://docs.docker.com/reference/dockerfile/#healthcheck)
that behaves like a [readiness probe](../../configuration/observability.md#probes).

## Production image

Our production image
[`ghcr.io/ferretdb/ferretdb:2.0.0`](https://ghcr.io/ferretdb/ferretdb:2.0.0)
is recommended for most deployments.
It does not include a PostgreSQL image with DocumentDB extension, so you must run this [pre-packaged PostgreSQL image with DocumentDB extension](../documentdb/docker.md) separately.

You can do that with Docker Compose, Kubernetes, or any other means.

:::tip
We strongly recommend specifying the full image tag (e.g., `2.0.0`)
to ensure consistency across deployments.
Ensure to [enable telemetry](../../telemetry.md) to receive notifications on the latest versions.

For more information on the best DocumentDB version to use, see the [corresponding release notes for the FerretDB version](https://github.com/FerretDB/FerretDB/releases/).
:::

### PostgreSQL setup with Docker Compose

The following steps describe a quick local setup:

1. Store the following in the `docker-compose.yml` file:

   <!-- TODO https://github.com/FerretDB/FerretDB/issues/4726 -->

   ```yaml
   services:
     postgres:
       image: ghcr.io/ferretdb/postgres-documentdb:17-0.102.0-ferretdb-2.0.0
       platform: linux/amd64
       restart: on-failure
       environment:
         - POSTGRES_USER=username
         - POSTGRES_PASSWORD=password
         - POSTGRES_DB=postgres
       volumes:
         - ./data:/var/lib/postgresql/data

     ferretdb:
       image: ghcr.io/ferretdb/ferretdb:2.0.0
       restart: on-failure
       ports:
         - 27017:27017
       environment:
         - FERRETDB_POSTGRESQL_URL=postgres://username:password@postgres:5432/postgres

   networks:
     default:
       name: ferretdb
   ```

   `postgres` container runs a pre-packaged PostgreSQL with DocumentDB extension and stores data in the `./data` directory on the host.
   `ferretdb` runs FerretDB.

2. Start services with `docker compose up -d`.
3. If you have `mongosh` installed, just run it to connect to FerretDB.
   It will use credentials passed in `mongosh` flags or MongoDB URI to authenticate to the PostgreSQL database.
   The example URI would look like:

   ```text
   mongodb://username:password@127.0.0.1/
   ```

   See [Authentication](../../security/authentication.md) and
   [Securing connection with TLS](../../security/tls-connections.md) for more details.

   If you don't have `mongosh`, run the following command to run it inside the temporary MongoDB container,
   attaching to the same Docker network:

   ```sh
   docker run --rm -it --network=ferretdb --entrypoint=mongosh \
     mongo mongodb://username:password@ferretdb/
   ```

You can improve that setup by:

- [securing connections with TLS](../../security/tls-connections.md);
- adding backups.

Find out more about:

- [getting logs](../../configuration/observability.md#docker-logs).

## Development image

The development image
[`ghcr.io/ferretdb/ferretdb-dev:2`](https://ghcr.io/ferretdb/ferretdb-dev:2)
contains the
[development build](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version#hdr-Development_builds)
of FerretDB, and is recommended for debugging problems.
It includes additional debugging features that make it significantly slower.
For this reason, it is not recommended for production use.
