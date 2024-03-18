---
slug: run-ferretdb-securely-using-openziti
title: 'Run FerretDB Securely Using OpenZiti'
authors: [alex]
description: >
  This guide will walk you through setting up FerretDB securely using OpenZiti.
image: /img/blog/ferretdb-openziti.jpg
tags: [tutorial, community, postgresql tools, open source]
---

![Run FerretDB ](/img/blog/ferretdb-openziti.jpg)

Securing your database is more critical than ever.
[FerretDB](https://www.ferretdb.com/), a truly open source document database with [PostgreSQL](https://www.postgresql.org/) as the backend, offers an excellent solution for developers looking for MongoDB-like experiences with the robustness of PostgreSQL.

<!--truncate-->

[OpenZiti](https://openziti.io/) is an open-source networking solution that offers zero-trust networking priciples directly to your application.

With OpenZiti, you can secure your FerretDB instances and connections, providing a zero-trust networking layer on top of it.

This guide will walk you through setting up FerretDB securely using OpenZiti.

## Prerequisites

Before we dive into the setup, ensure you have [Docker](https://www.docker.com/) installed on your system.

## Guide on setting up FerretDB with OpenZiti

### Setting Up the Environment

The [cheatsheet guide provided by OpenZiti](https://github.com/openziti/ziti-sdk-jvm/blob/main/samples/jdbc-postgres/cheatsheet.md) outlines steps for creating Ziti services, enrolling identities, and configuring policies for secure, zero-trust access to your database.

Run the following commands in your terminal:

```sh
curl -s https://get.openziti.io/dock/simplified-docker-compose.yml > docker-compose.yml
curl -s https://get.openziti.io/dock/.env > .env
```

These commands will fetch the Docker Compose file and the default environment file required for OpenZiti setup.

Next, modify the docker compose file and add Postgres with a known user/password and FerretDB should connect to the Postgres database using the `FERRETDB_POSTGRES_URI`, as shown below.

```yaml
postgres-db:
  image: postgres
  #ports:
  #  - 5432:5432
  networks:
    - ziti
  volumes:
    - ./data/db:/var/lib/postgresql/data
  environment:
    - POSTGRES_DB=ferretdb
    - POSTGRES_USER=postgres
    - POSTGRES_PASSWORD=postgres

ferretdb:
  image: ghcr.io/ferretdb/ferretdb
  restart: on-failure
  networks:
    - ziti
  environment:
    - FERRETDB_POSTGRESQL_URL=postgres://postgres-db/ferretdb
```

This setup provides a secure, isolated network for your FerretDB instance and Postgres database, ensuring that your database is not exposed to the internet.

### Docker Compose Configuration

The Docker Compose file outlines the setup for running FerretDB with OpenZiti, including services for the Ziti Controller, Edge Router, Ziti Console, PostgreSQL database, and FerretDB itself.

Let's break down the key components:

- Ziti Controller: Manages the Ziti network, identities, and policies.
- Ziti Edge Router: Handles encrypted traffic between Ziti clients and services.
- Ziti Console: Provides a web interface for managing the Ziti network.

FerretDB connects to PostgreSQL and is configured to run within the same Ziti network.

### Initialize Docker Environment

Using the project name 'pg' for Postgres, start the Docker Compose environment:

```sh
docker compose -p pg up
```

This command sets up several services, including `ziti-controller`, `ziti-edge-router`, `ziti-console`, `postgres-db`, and `ferretdb`, as specified in the Docker Compose YAML.

Run `docker ps` to show that Postgres and FerretDB are not exposed (`5432`/`27017`):

### Testing your network connection

To verify the security and functionality of your setup, follow these steps to test network connectivity between the Ziti components and ensure they're correctly configured.

To begin, access the running Ziti Controller container using the following command:

```sh
docker exec -it pg-ziti-controller-1 bash
```

This should take you into the Ziti CLI.
Authenticate using the `zitiLogin` alias:

```sh
zitiLogin
```

Test edge routers online:

Verify that all edge routers are online and properly registered with the Ziti Controller:

```sh
ziti edge list edge-routers
```

You should see your ziti-edge-router listed and marked as ONLINE.

Test edge router identities:

Each edge router should have an associated identity within the Ziti network.
Check these identities:

```sh
ziti edge list identities
```

This command lists all registered identities, including those for your routers.

Test network connectivity:

Let's ensure the Ziti Controller and a Ziti Edge Router can communicate over the network.

Use ping from within the controller container to verify connectivity to the `ziti-edge-router`:

```sh
$ ping ziti-edge-router -c 1
PING ziti-edge-router (172.26.0.6): 56 data bytes
64 bytes from 172.26.0.6: icmp_seq=0 ttl=64 time=0.562 ms
--- ziti-edge-router ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.562/0.562/0.562/0.000 ms
```

These tests ensure that your Docker network settings allow for proper communication paths between the Ziti Controller and the Ziti Edge Router, ensuring that your FerretDB setup with OpenZiti operates more efficiently and securely.

### Connecting to FerretDB

Once your services are up and running, you can connect to FerretDB using the MongoDB shell (`mongosh`) over the secure network established by OpenZiti.

```sh
docker run -it --rm --network pg_ziti mongo mongosh "mongodb://postgres:postgres@pg-ferretdb-1:27017/ferretdb?authMechanism=PLAIN"
```

This should spin up a temporary MongoDB container to use `mongosh` to connect to your FerretDB instance.

## Securing Your FerretDB Setup with OpenZiti

OpenZiti secures FerretDB by establishing a zero-trust network, and minimizing the attack surface.
It encrypts data end-to-end, prevents eavesdropping and tampering, closes all inbound firewall ports, and ensures seamless connectivity without exposing FerretDB to the internet.

Now that you understand how to secure FerretDB with OpenZiti, be sure to try it out in your project and let us know how it goes.
And if you have any questions, please [reach out to us on any of our channels](https://docs.ferretdb.io/#community).

[Learn more about FerretDB](https://docs.ferretdb.io/).

[Learn more about OpenZiti](https://openziti.io/docs/learn/introduction/).
