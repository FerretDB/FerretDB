---
sidebar_position: 1
---

# Docker

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/documentdb/documentdb) as a database engine.

We provide two Docker images for setting up PostgreSQL with DocumentDB extension:

- Production image for stable and optimized deployments.
- Development image for debugging problems.

## Installation

The production image for PostgreSQL with DocumentDB extension
[`ghcr.io/ferretdb/postgres-documentdb:17-0.108.0-ferretdb-2.8.0`](https://ghcr.io/ferretdb/postgres-documentdb:17-0.108.0-ferretdb-2.8.0)
is recommended for most deployments.
It does not include FerretDB, so you must run it separately.
For a complete setup that includes FerretDB, see the [FerretDB installation](../ferretdb/docker.md).

:::tip
We strongly recommend specifying the full image tag (e.g., `17-0.108.0-ferretdb-2.8.0`)
to ensure consistency across deployments.
For more information on the best FerretDB image to use, see the [DocumentDB release notes](https://github.com/FerretDB/documentdb/releases/).
:::

### Run production image

FerretDB requires a user credential (username and password) and `postgres` database to be initialized and must already exist before a connection can be set up.
See [Set up PostgreSQL connection](../../security/authentication.md#set-up-postgresql-connection) for more details.

1. You can run the image with the following command:

   ```sh
   docker run -d \
     --restart on-failure \
     -e POSTGRES_USER=<username> \
     -e POSTGRES_PASSWORD=<password> \
     -e POSTGRES_DB=postgres \
     -v ./data:/var/lib/postgresql/data \
     -p 5432:5432 \
     ghcr.io/ferretdb/postgres-documentdb:17-0.108.0-ferretdb-2.8.0
   ```

   Ensure to update `<username>` and `<password>` with your desired values.

2. If you have `psql` installed, you can connect to the PostgreSQL with DocumentDB extension with the following command:

   ```sh
   psql 'postgres://<username>:<password>@localhost:5432/postgres'
   ```

   If you don't have `psql`, you can run the following command to run it inside the temporary PostgreSQL container:

   ```sh
   docker run --rm -it ghcr.io/ferretdb/postgres-documentdb:17-0.108.0-ferretdb-2.8.0 \
     psql postgres://<username>:<password>@localhost:5432/postgres
   ```

3. See [FerretDB Docker installation](../ferretdb/docker.md) for more details on connecting to FerretDB.

### Run development image

The development image for PostgreSQL with DocumentDB extension
[`ghcr.io/ferretdb/postgres-documentdb-dev`](https://ghcr.io/ferretdb/postgres-documentdb-dev)
is recommended for debugging problems.
It includes additional debugging features that make it significantly slower.
For this reason, it is not recommended for production use.

## Updating to a new version

Before [updating to a new FerretDB release](../ferretdb/docker.md#updating-to-a-new-version), it is critical to install the matching DocumentDB image first.

The following steps are critical to ensuring a successful update.

Edit your Compose file to use the matching DocumentDB image tag.
You can find the correct tag in the DocumentDB release notes (for example: `17-0.108.0-ferretdb-2.8.0`).
Then run:

```sh
docker compose pull <documentdb-container-name>
docker compose up -d <documentdb-container-name>
```

Next, from within your `postgres` database, upgrade the DocumentDB extension by running:

```sh
docker compose exec <documentdb-container-name> \
  psql -U <username> -d postgres -c 'ALTER EXTENSION documentdb UPDATE;'
```

Replace `<documentdb-container-name>`, `<username>`, `<password>`, and `<host>` with the appropriate values for your setup.

For more details, refer to the [Docker Compose documentation](https://docs.docker.com/compose/) and the [PostgreSQL documentation](https://www.postgresql.org/docs/).
After the extension update, verify or update `postgresql.conf` settings to ensure the following configurations are present:

```sh
docker compose exec -T  <documentdb-container-name> \
  psql -U <username> -d postgres <<EOF
ALTER SYSTEM SET shared_preload_libraries = 'pg_cron','pg_documentdb_core','pg_documentdb';
ALTER SYSTEM SET cron.database_name = 'postgres';

ALTER SYSTEM SET documentdb.enableCompact = true;

ALTER SYSTEM SET documentdb.enableLetAndCollationForQueryMatch = true;
ALTER SYSTEM SET documentdb.enableNowSystemVariable = true;
ALTER SYSTEM SET documentdb.enableSortbyIdPushDownToPrimaryKey = true;

ALTER SYSTEM SET documentdb.enableSchemaValidation = true;
ALTER SYSTEM SET documentdb.enableBypassDocumentValidation = true;

ALTER SYSTEM SET documentdb.enableUserCrud = true;
ALTER SYSTEM SET documentdb.maxUserLimit = 100;
EOF
```

Restart the container to apply changes:

```sh
docker compose restart <documentdb-container-name>
```

Once the DocumentDB update is ready, follow the FerretDB update process to update FerretDB.
