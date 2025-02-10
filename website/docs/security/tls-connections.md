---
sidebar_position: 2
description: Learn to secure connections using TLS
---

# TLS Connections

It is possible to encrypt connections between FerretDB and clients by using TLS.
All you need to do is to start the server with the following flags or environment variables:

- `--listen-tls` / `FERRETDB_LISTEN_TLS` specifies the TCP hostname and port
  that will be used for listening for incoming TLS connections.
  If empty, TLS listener is disabled;
- `--listen-tls-cert-file` / `FERRETDB_LISTEN_TLS_CERT_FILE` specifies the PEM encoded, TLS certificate file
  that will be presented to clients;
- `--listen-tls-key-file` / `FERRETDB_LISTEN_TLS_KEY_FILE` specifies the TLS private key file
  that will be used to decrypt communications;
- `--listen-tls-ca-file` / `FERRETDB_LISTEN_TLS_CA_FILE` specifies the root CA certificate file
  that will be used to verify client certificates.

Then use `tls` query parameters in MongoDB URI for the client.
You may also need to set `tlsCAFile` parameter if the system-wide certificate authority did not issue the server's certificate.
See documentation for your client or driver for more details.
Example: `mongodb://ferretdb:27018/?tls=true&tlsCAFile=companyRootCA.pem`.

## Setting up TLS connections

In the following examples, FerretDB uses TLS certificates to secure the connection.
The `ferretdb` server uses TLS server certificate file, TLS private key file and root CA certificate file.

```text
server-certs/
├── rootCA-cert.pem
├── server-cert.pem
└── server-key.pem
```

The client uses TLS client certificate file and root CA certificate file.

```text
client-certs/
├── client.pem
└── rootCA-cert.pem
```

### Setting up TLS via Docker

When using Docker to run `ferretdb` server, the `docker-compose.yml` below shows how to set up TLS connections.
The Docker host requires certificates `server-certs` directory,
and volume is mounted from `./server-certs` of Docker host to `/etc/certs` of Docker container.

<!-- TODO https://github.com/FerretDB/FerretDB/issues/4726 -->

```yaml
services:
  postgres:
    image: ghcr.io/ferretdb/postgres-documentdb:16
    platform: linux/amd64
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=postgres
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2
    restart: on-failure
    ports:
      - 27018:27018
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://username:password@localhost:5432/postgres
      - FERRETDB_LISTEN_TLS=:27018
      - FERRETDB_LISTEN_TLS_CERT_FILE=/etc/certs/server-cert.pem
      - FERRETDB_LISTEN_TLS_KEY_FILE=/etc/certs/server-key.pem
      - FERRETDB_LISTEN_TLS_CA_FILE=/etc/certs/rootCA-cert.pem
    volumes:
      - ./server-certs:/etc/certs

networks:
  default:
    name: ferretdb
```

To start `ferretdb`, run the following command:

```sh
docker compose up
```

In the following example, a client connects to MongoDB URI using TLS certificates as `username`.
It uses Docker volume to mount `./clients-certs` of Docker host to `/clients` Docker container.

Connect to `ferretdb` server using `mongosh` client:

```sh
docker run --rm -it \
  --network=ferretdb \
  --volume ./client-certs:/clients \
  --entrypoint=mongosh \
  mongo 'mongodb://username:password@host.docker.internal:27018?tls=true&tlsCertificateKeyFile=/clients/client.pem&tlsCaFile=/clients/rootCA-cert.pem'
```
