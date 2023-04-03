---
sidebar_position: 9
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
but clients that use `mongodb://user2:pass2@ferretdb:27018/?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of TLS is highly recommended.
