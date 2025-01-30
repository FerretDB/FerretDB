---
sidebar_position: 1
---

# Docker

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

We provide two Docker images for setting up PostgreSQL with DocumentDB extension:

- **a development image**: for debugging problems
- **a production image**: for all other cases

## Production image for PostgreSQL with DocumentDB extension

The [production image for PostgreSQL with DocumentDB extension](https://ghcr.io/ferretdb/postgres-documentdb:16) (`ghcr.io/ferretdb/postgres-documentdb:16`) is recommended for most deployments.
It does not include FerretDB, so you must run it separately.
For a complete setup that includes FerretDB, see the [FerretDB installation](../ferretdb/docker.md).

### Running the image

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
     ghcr.io/ferretdb/postgres-documentdb:16
   ```

   Ensure to update `<username>` and `<password>` with your desired values.

2. If you have `psql` installed, you can connect to the PostgreSQL with DocumentDB extension with the following command:

   ```sh
   psql 'postgres://<username>:<password>@localhost:5432/postgres'
   ```

   If you don't have `psql`, you can run the following command to run it inside the temporary PostgreSQL container:

   ```sh
   docker run --rm -it ghcr.io/ferretdb/postgres-documentdb:16 psql 'postgres://<username>:<password>@localhost:5432/postgres'
   ```

3. See [FerretDB Docker installation](../ferretdb/docker.md) for more details on connecting to FerretDB.

## Development image for PostgreSQL with DocumentDB extension

The [development image for PostgreSQL with DocumentDB extension](https://ghcr.io/ferretdb/postgres-documentdb-dev:16) (`ghcr.io/ferretdb/postgres-documentdb-dev:16`) is recommended for debugging problems.
It includes additional debugging features and is not recommended for production use.
