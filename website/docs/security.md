---
sidebar_position: 11
slug: /security/ # referenced in README.md
description: TLS and authentication
---

# Security

## Securing connections with TLS

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

## Authentication

FerretDB does not store authentication information (usernames and passwords) itself but uses the backend's authentication mechanisms.
The default username and password can be specified in FerretDB's connection string,
but the client could use a different user by providing a username and password in MongoDB URI.
For example, if the server was started with `postgres://user1:pass1@postgres:5432/ferretdb`,
anonymous clients will be authenticated as user1,
but clients that use `mongodb://user2:pass2@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of TLS is highly recommended.

### PostgreSQL backend with default username and password

PostgreSQL server may start with default username and password.

In the example, default username and password are specified in FerretDB's connection string `user1:pass1`.
See more about [creating PostgreSQL user](https://www.postgresql.org/docs/current/sql-createuser.html)
and [PostgreSQL authentication methods](https://www.postgresql.org/docs/current/auth-methods.html).

Using `ferretdb` package, specify `--postgresql-url` with default username and password.

```sh
ferretdb --postgresql-url=postgres://user1:pass1@postgres:5432/ferretdb
```

For using docker, specify `FERRETDB_POSTGRESQL_URL` with default username and password.

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
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://user1:pass1@postgres:5432/ferretdb

networks:
  default:
    name: ferretdb
```

An anonymous user is authenticated with `user1` from `FERRETDB_POSTGRESQL_URL`.
An example of anonymous user is below using `mongosh`.

```sh
mongosh `mongodb://127.0.0.1/ferretdb`
```

If you don't have mongosh, run the following command to run it inside the temporary MongoDB container,
attaching to the same Docker network. To use this, start `ferretdb` with docker compose.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo `mongodb://ferretdb/ferretdb`
```

Clients that specify username and password in MongoDB URI such as below is authenticated as `user2` using `mongosh`.

```sh
mongosh `mongodb://user2:pass2@127.0.0.1/ferretdb?authMechanism=PLAIN`
```

For running it inside the temporary MongoDB container, use below.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
`mongodb://user2:pass2@ferretdb/ferretdb?authMechanism=PLAIN`
```

### PostgreSQL backend with TLS

Using TLS is recommended if username and password are transferred in plain text.

In this example, FerretDB uses TLS certificates to secure the connection.
Example certificates are found in [build/certs](https://github.com/FerretDB/FerretDB/tree/main/build/certs).

The server uses TLS server certificate file, TLS private key file and root CA certificate file.

```console
server-certs/
├── rootCA-cert.pem
├── server-cert.pem
└── server-key.pem
```

This connects to localhost PostgreSQL instance using TLS with certificates in `server-certs` directory.
Be sure to check that `server-certs` directory is present.

```sh
ferretdb \
--postgresql-url=postgres://user1:pass1@localhost:5432/ferretdb \
--listen-tls=:27018 \
--listen-tls-cert-file=./server-certs/server-cert.pem \
--listen-tls-key-file=./server-certs/server-key.pem \
--listen-tls-ca-file=./server-certs/rootCA-cert.pem
```

For using docker to run FerretDB and those who do not have PostgreSQL server setup,
`docker-compose.yml` example for TLS is provided in below.

```console
.
├── docker-compose.yml
└── server-certs
    ├── rootCA-cert.pem
    ├── server-cert.pem
    └── server-key.pem
```

In this example, it uses docker volume to mount `./server-certs` of docker host to `/etc/certs` docker container.

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
      - FERRETDB_POSTGRESQL_URL=postgres://user1:pass1@postgres:5432/ferretdb
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

The client uses TLS client certificate file and root CA certificate file.
In this example, it uses `client-certs` for client authentication.

```console
client-certs/
├── client.pem
└── rootCA-cert.pem
```

If you have mongosh installed, following connects to mongosh to connect as `user2` using TLS
with certificates in `client-certs` directory.
Be sure to check that `client-certs` directory is present.

```sh
mongosh 'mongodb://user2:pass2@127.0.0.1:27018/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=./client-certs/client.pem&tlsCaFile=./client-certs/rootCA-cert.pem'
```

Otherwise, you can use docker to connect as `user2`.
In this example, it uses docker volume to mount `./server-certs` of docker host to `/clients` docker container.

```sh
docker run --rm -it --network=ferretdb -v ./client-certs:/clients --entrypoint=mongosh mongo \
'mongodb://user2:pass2@host.docker.internal:27018/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=/clients/client.pem&tlsCaFile=/clients/rootCA-cert.pem'
```

Note that it uses `host.docker.internal` host in the above, because it
needs to match certificate's [altnames](https://github.com/FerretDB/FerretDB/blob/main/build/certs/Makefile).
