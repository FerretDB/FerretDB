---
sidebar_position: 11
slug: /security/ # referenced in README.md
description: TLS and authentication
---

# Security

## Securing connections with TLS

It is possible to encrypt connections between FerretDB and clients by using TLS.
All you need to do is to start the server with the following flags or environment variables:

* `--listen-tls` / `FERRETDB_LISTEN_TLS` specifies the TCP hostname and port
  that will be used for listening for incoming TLS connections.
  If empty, TLS listener is disabled;
* `--listen-tls-cert-file` / `FERRETDB_LISTEN_TLS_CERT_FILE` specifies the PEM encoded, TLS certificate file
  that will be presented to clients;
* `--listen-tls-key-file` / `FERRETDB_LISTEN_TLS_KEY_FILE` specifies the TLS private key file
  that will be used to decrypt communications;
* `--listen-tls-ca-file` / `FERRETDB_LISTEN_TLS_CA_FILE` specifies the root CA certificate file
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
See more about [PostgreSQL Authentication Methods](https://www.postgresql.org/docs/current/auth-methods.html).

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

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo `mongodb://ferretdb/ferretdb`
```

Clients that specify username and password in MongoDB URI such as below is authenticated as `user2`.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
`mongodb://user2:pass2@ferretdb/ferretdb?authMechanism=PLAIN`
```

### PostgreSQL backend with TLS

Using TLS is recommended if username and password are transferred in plain text.

In this example, FerretDB uses TLS certificates to secure the connection.
Replace `<CERT_PATH>` with certificate path.

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
      - <CERT_PATH>/server-cert.pem:/etc/certs/server-cert.pem:ro
      - <CERT_PATH>/build/certs/server-key.pem:/etc/certs/server-key.pem:ro
      - <CERT_PATH>/build/certs/rootCA-cert.pem:/etc/certs/rootCA-cert.pem:ro

networks:
  default:
    name: ferretdb
```

Clients that specify username and password in MongoDB URI such as below uses TLS for securing connection.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
`mongodb://user2:pass2@ferretdb:27018/ferretdb?uthMechanism=PLAIN&tls=true&tlsCertificateKeyFile=./build/certs/client.pem&tlsCaFile=./build/certs/rootCA-cert.pem`
```
