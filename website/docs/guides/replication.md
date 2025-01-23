---
sidebar_position: 1
description: Learn about replication and how to configure it in FerretDB using PostgreSQL with DocumentDB extension.
---

# Replication

Database replication is the continuous process of synchronizing and distributing data across multiple servers, often located in different geographical regions.
This involves consistently replicating data from a primary server to one or more replica servers.

The goal of replication is to create fault-tolerant, highly available, and reliable database environments.
That means:

- Enabling disaster recovery by preventing downtime or data loss through promoting a replica when the primary server is down.
- Distributing load across different databases or regions.

## Understanding replication in FerretDB using PostgreSQL with DocumentDB extension

FerretDB uses PostgreSQL + DocumentDB extension as the database storage and currently supports replication using the Write-Ahead Logging (WAL) streaming method.
WAL is a logging method that writes every change in the database to a log (WAL) before applying it to the primary PostgreSQL data.
Other replication methods, such as logical replication, may not work because the replica requires permissions to create tables, schemas, and perform other operations.

There are two main components in a replication setup:

- **Primary server**: The main instance that handles all write operations.
- **Replica server**: A read-only secondary instance that receives real-time WAL data from the primary.

See the [Postgres documentation to understand more on WAL (Write-Ahead Logging)](https://www.postgresql.org/docs/current/wal-intro.html).

### Primary server

The Primary Server is the main PostgreSQL instance that handles all write operations.

The primary server streams WALs to one or more replicas, which then apply them.
This is an asynchronous replication process – the primary does not wait for a response from the replica before confirming the transaction to the client.
It only sends WAL to the replica _after_ committing the transactions locally.

### Replica server

The Replica Server is a read-only secondary instance that receives real-time WAL data from the primary.
There can be one or more read-only replicas connected to the primary server.
In normal operation, it handles read-only queries and can be used to offload/decrease read operations from the primary server.
During failover, a replica can also be configured to be promoted to primary to ensure high availability.

## Setting up replication for FerretDB

This guide walks through setting up streaming replication between a primary and replica PostgreSQL instance for use with FerretDB.
This configuration allows your data in the primary PostgreSQL instance to be replicated in real-time to a replica.

A typical view of the setup:

```mermaid
flowchart LR
    subgraph Application
        G[Client]
    end

    subgraph FerretDB
        G -->|Connects to| A[FerretDB Primary]
        G -->|Connects to| F[FerretDB Read-only]
    end

    subgraph PostgreSQL
        A -->|Write/Read| B[Postgres Primary]
        F -->|Read-only| C[Postgres Replica]
        B -->|Replication - WAL Streaming| C[Postgres Replica]
    end
```

:::note

1. Use of the term `PostgreSQL` or `Postgres` in this context refers to PostgreSQL with DocumentDB extension.
2. The setup and configuration in this guide are only for demonstration purposes and may need to be adjusted based on your production requirements and environment.

:::

### Step 1: Docker Compose setup

We'll start by setting up the following services with Docker Compose:

- Primary PostgreSQL instance (`postgres_primary`)
- Replica PostgreSQL instance (`postgres_replica`)
- FerretDB instance (`ferretdb`)
- FerretDB read-only (`ferretdb_readonly`)

Each of these services should be initialized one after the other to ensure the primary PostgreSQL instance is set up before the replica.
That way, the replica can be set up using the base backup of the primary PostgreSQL instance.

Here's the `docker-compose.yml` file:

```yaml
services:
  postgres_primary:
    image: ghcr.io/ferretdb/postgres-documentdb:16
    container_name: postgres_primary
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=postgres
    volumes:
      - ./primary_data:/var/lib/postgresql/data
      - replica_backup:/tmp/replica_backup
    ports:
      - '5432:5432'

  postgres_replica:
    image: ghcr.io/ferretdb/postgres-documentdb:16
    container_name: postgres_replica
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=postgres
    volumes:
      - replica_backup:/var/lib/postgresql/data
    ports:
      - '5433:5432'

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2
    container_name: ferretdb
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://user:password@postgres_primary:5432/postgres
    ports:
      - '27017:27017'

  ferretdb_readonly:
    image: ghcr.io/ferretdb/ferretdb:2
    container_name: ferretdb_readonly
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://user:password@postgres_replica:5432/postgres
    ports:
      - '27018:27017'
    depends_on:
      - postgres_replica

networks:
  default:
    name: ferretdb

volumes:
  replica_backup:
```

To initialize the setup, start the primary PostgreSQL instance (`postgres_primary`) first:

```sh
docker compose up -d postgres_primary
```

### Step 2: Create a replication role

Connect to the `postgres_primary` to configure the replication role and permissions.

```sh
docker exec -it postgres_primary psql -U user -d postgres -c "CREATE ROLE replicator WITH REPLICATION PASSWORD 'replicatorpassword' LOGIN;"
```

This role will allow the replica to connect and receive data from the primary.

### Step 3: Configure primary PostgreSQL instance for replication

Next, update `postgresql.conf` and `pg_hba.conf` to allow replication.

Inside the primary PostgreSQL container, configure the WAL settings necessary for replication.
You can access the file by opening the container's shell:

```sh
docker exec -it postgres_primary bash
```

Then access `/var/lib/postgresql/data/postgresql.conf` to add the following settings:

```sh
cat <<EOF >> /var/lib/postgresql/data/postgresql.conf
wal_level = replica
max_wal_senders = 10
wal_keep_size = 64MB
listen_addresses = '*'
EOF
```

- **`wal_level = replica`**: Enables WAL logging for replication.
- **`max_wal_senders = 10`**: Allows up to 10 replicas to connect simultaneously.
- **`wal_keep_size = 64MB`**: Retains 64MB of WAL data to ensure replicas can catch up.
- **`listen_addresses = '*'`**: Allows PostgreSQL to accept connections from any IP, enabling access by replicas.

Run the following command to add the replication user to `pg_hba.conf`:

```sh
echo 'host replication replicator 0.0.0.0/0 md5' >> /var/lib/postgresql/data/pg_hba.conf
```

This setting permits the `replicator` user to connect for replication purposes from any IP using MD5 password authentication.

In PostgreSQL, reload the configuration to apply changes without a full restart.

```sh
docker exec -it postgres_primary psql -U user -d postgres -c "SELECT pg_reload_conf();"
```

### Step 4: Prepare the replica

Use `pg_basebackup` to initialize the replica with data from the primary.

```sh
docker exec -it postgres_primary pg_basebackup -D /tmp/replica_backup -P -R -X stream -c fast -U replicator
```

Launch the replica.

```sh
docker compose up -d postgres_replica
```

You may need to give the PostgreSQL user permission to the shared directory depending on your setup.

```sh
docker exec -it postgres_replica chown -R postgres:postgres /var/lib/postgresql/data
```

### Step 5: Configure replica to follow primary

Connect to the replica `postgres_replica` and configure it to follow `postgres_primary` so it knows where to receive the WAL data.

```sh
docker exec -it postgres_replica psql -U user -d postgres -c "ALTER SYSTEM SET primary_conninfo TO 'host=postgres_primary port=5432 user=replicator password=replicatorpassword';"
```

Create a `standby.signal` file to tell PostgreSQL this instance should act as a replica.
[See the standby documentation for more](https://www.postgresql.org/docs/current/warm-standby.html).

```sh
docker exec -it postgres_replica bash -c "touch /var/lib/postgresql/data/standby.signal"
```

Restart the replica to apply the changes.

```sh
docker restart postgres_replica
```

### Step 6: Test the replication

To confirm that the replica is receiving data, inspect the replication status on `postgres_primary`.

```sh
docker exec -it postgres_primary psql -U user -d postgres -c "SELECT * FROM pg_stat_replication;"
```

The output should show the replica's connection details like below:

```text
 pid  | usesysid |  usename   | application_name |  client_addr  | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn | write_lag | flush_lag | replay_lag | sync_priority | sync_state |          reply_time
------+----------+------------+------------------+---------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------+-----------+------------+---------------+------------+------------------------------
 6359 |    18469 | replicator | walreceiver      | 192.168.160.3 |                 |       35858 | 2024-11-01 16:42:23.561418+00 |              | streaming | 0/4132270 | 0/4132270 | 0/4132270 | 0/4132270  |           |           |            |             0 | async      | 2024-11-01 17:39:25.65636+00
(1 row)
```

On the replica, confirm that it's in recovery mode, which indicates it's in sync with the primary.

```sh
docker exec -it postgres_replica psql -U user -d postgres -c "SELECT pg_is_in_recovery();"
```

The output should show `t` (true) to indicate that the replica is in recovery mode.

```text
 pg_is_in_recovery
-------------------
 t
(1 row)
```

### Step 7: Connect FerretDB

With the primary and replica configured, launch the FerretDB container to connect to the primary.

```sh
docker compose up -d ferretdb
```

If you have `mongosh` installed, connect to FerretDB using the MongoDB URI `mongodb://user:password@localhost:27017/`.
Otherwise, run the following command to connect to FerretDB using `mongosh` inside a temporary MongoDB container:

```sh
docker run --rm -it --network="host" mongo mongosh "mongodb://user:password@localhost:27017/"
```

Add the following document in FerretDB to test if data written to the primary also appears in the replica.

```js
db.record.insertOne({
  username: 'Ada Lovelace',
  content: 'Enjoying the beautiful weather today! 🌞 #sunnyday',
  likes: 120,
  timestamp: new Date()
})
```

### Step 8: Verify replication with read-only FerretDB

Let's set up another FerretDB instance to connect to the replica and verify that the data written to the `postgres_primary` is also available on the `postgres_replica`.
The FerretDB read-only instance can only perform read operations and cannot write data.

```sh
docker compose up -d ferretdb_readonly
```

Connect to the container via `mongosh` and verify that the data written to the primary is also available in the replica.

```sh
docker run --rm -it --network="host" mongo mongosh "mongodb://user:password@localhost:27018/" --eval "db.record.find()"
```

The output should show the document you inserted in the primary.

```json5
[
 {
   _id: ObjectId('672159e5a6caf9a94afa26b9'),
   username: 'Ada Lovelace',
   content: 'Enjoying the beautiful weather today! 🌞 #sunnyday',
   likes: 120,
   timestamp: ISODate('2024-10-29T21:55:49.795Z')
 }
]
```
