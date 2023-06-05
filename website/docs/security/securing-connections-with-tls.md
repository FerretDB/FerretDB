---
sidebar_position: 2
description: Learn to secure connections using TLS
---

# Securing connections with TLS

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

## PostgreSQL backend with TLS

Using TLS is recommended if username and password are transferred in plain text.

In following examples, FerretDB uses TLS certificates to secure the connection.
Example certificates are found in [build/certs](https://github.com/FerretDB/FerretDB/tree/main/build/certs).
The `ferretdb` server uses TLS server certificate file, TLS private key file and root CA certificate file.

```console
server-certs/
├── rootCA-cert.pem
├── server-cert.pem
└── server-key.pem
```

The client uses TLS client certificate file and root CA certificate file.

```console
client-certs/
├── client.pem
└── rootCA-cert.pem
```

### Using TLS with `ferretdb` package

The example below connects to localhost PostgreSQL instance using TLS with certificates in `server-certs` directory.
Be sure to check that `server-certs` directory and files are present.

```sh
ferretdb \
--postgresql-url=postgres://localhost:5432/ferretdb \
--listen-tls=:27018 \
--listen-tls-cert-file=./server-certs/server-cert.pem \
--listen-tls-key-file=./server-certs/server-key.pem \
--listen-tls-ca-file=./server-certs/rootCA-cert.pem
```

Using `mongosh`, a client connects to ferretdb as `user2` using TLS certificates in `client-certs` directory.
Be sure to check that `client-certs` directory and files are present.

```sh
mongosh 'mongodb://user2:pass2@127.0.0.1:27018/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=./client-certs/client.pem&tlsCaFile=./client-certs/rootCA-cert.pem'
```

### Using TLS with docker

For using docker to run `ferretdb` server, `docker-compose.yml` example for TLS is provided in below.
The docker host requires certificates `server-certs` directory,
and volume is mounted from `./server-certs` of docker host to `/etc/certs` of docker container.

```yml
services:
  postgres:
    image: postgres
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=ferretdb
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    ports:
      - 27018:27018
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://postgres:5432/ferretdb
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

To start `ferretdb`, use docker compose.

```sh
docker compose up
```

In the following example, a client connects to MongoDB URI using TLS certificates as `user2`.
It uses docker volume to mount `./clients-certs` of docker host to `/clients` docker container.

```sh
docker run --rm -it \
--network=ferretdb \
--volume ./client-certs:/clients \
--entrypoint=mongosh mongo \
'mongodb://user2:pass2@host.docker.internal:27018/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=/clients/client.pem&tlsCaFile=/clients/rootCA-cert.pem'
```

Note that MongoDB URI uses `host.docker.internal` host in above, because it
needs to match certificate's [altnames](https://github.com/FerretDB/FerretDB/blob/main/build/certs/Makefile).
