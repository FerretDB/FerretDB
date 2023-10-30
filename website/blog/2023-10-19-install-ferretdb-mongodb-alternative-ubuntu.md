---
slug: install-ferretdb-mongodb-alternative-ubuntu
title: How to Install FerretDB on Ubuntu
authors: [alex]
description: >
  In this tutorial, you will learn to install FerretDB on Ubuntu, using either PostgreSQL or SQLite backend.
image: /img/blog/install-ferretdb-ubuntu.jpg
tags: [tutorial, product, open source]
---

![How to install FerretDB on Ubuntu](/img/blog/install-ferretdb-ubuntu.jpg)

[FerretDB](https://www.ferretdb.io/) is an open-source document database that allows you to run MongoDB commands and queries with options to use either [PostgreSQL](https://www.postgresql.org/) or [SQLite](https://www.sqlite.org/) as the backend.
While you can run FerretDB on many operating systems including [Ubuntu](https://ubuntu.com/), there are separate requirements for each OS.

<!--truncate-->

Installing FerretDB on Ubuntu can be a tad tricky, especially when setting up your backend choice, or `systemd` service.

In this tutorial, you will learn to install FerretDB on Ubuntu, and how to set it using a `systemd` file.

## Prerequisites

- [Ubuntu OS](https://ubuntu.com/desktop): This tutorial is based on Ubuntu 22.1.0 arm64.
  Before you start, it's always a good idea to have a non-root user account with `sudo` privileges configured on your Ubuntu system.
- `mongosh`

## A guide on installing FerretDB on Ubuntu

In this guide, we will focus specifically on running the PostgreSQL backend for FerretDB.

Before we start, please ensure that your system and packages are up-to-date.
Do that by running:

```sh
sudo apt update
sudo apt upgrade
```

### Step 1: Install PostgreSQL

Since we are setting up FerretDB to use the PostgreSQL backend, we'll need to [install PostgreSQL](https://www.postgresql.org/download/), if it's not already installed.

```sh
sudo apt install postgresql
```

Run `psql --version` to confirm that PostgreSQL is installed correctly.

Before connecting to a PostgreSQL instance, do note that FerretDB requires that you have a `ferretdb` database on `PostgreSQL` associated with an authentication credential.

To do this, we need to connect to the default PostgreSQL instance:

```sh
sudo -u postgres psql
```

This will open up an instance where you can create a new user and assign it to the FerretDB instance.

```text
~$ sudo -u postgres psql
psql (14.9 (Ubuntu 14.9-0ubuntu0.22.04.1))
Type "help" for help.

postgres=#
```

Create a new user and password (make sure to update the credentials with the correct details before running the command):

```sql
CREATE ROLE <username> WITH PASSWORD '<password>';
```

Create a new database named `ferretdb` and grant all privileges to the created user.

```sql
CREATE DATABASE ferretdb;
```

Grant all privileges to the user:

```sql
GRANT ALL PRIVILEGES ON DATABASE ferretdb TO <username>;
```

Our PostgreSQL backend is ready!

Exit the PostgreSQL prompt with `\q`, and then connect back to the `ferretdb` database using the username and the password you created.

```sh
psql -h localhost -U <username> -d ferretdb
```

```text
psql (14.9 (Ubuntu 14.9-0ubuntu0.22.04.1))
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

ferretdb=>
```

### Step 2: Download and install FerretDB

To download FerretDB on Ubuntu, go to the [official releases page of FerretDB](https://github.com/FerretDB/FerretDB/releases).
FerretDB offers both `amd64` and `arm64` binaries.

Choose the package suitable for your Ubuntu operating system.
For this tutorial, we are using the `arm64` deb package for FerretDB v1.12.1.

```sh
wget https://github.com/FerretDB/FerretDB/releases/download/v1.12.1/ferretdb-arm64.deb
```

Or you can use `curl` to download the package:

```sh
curl -LJO https://github.com/FerretDB/FerretDB/releases/download/v1.12.1/ferretdb-arm64.deb
```

From the directory where `ferretdb-arm64.deb` is located, install FerretDB:

```sh
sudo apt install ./ferretdb-arm64.deb
```

Check that FerretDB has installed successfully:

```text
ferretdb --version
version: v1.12.1
commit: d1486f2b5d86eadfa6d148752b14fdde49cb5db9
branch: unknown
dirty: true
package: deb
debugBuild: false
```

FerretDB provides numerous [configuration flags](https://docs.ferretdb.io/configuration/flags/) that you can tailor to your needs.
Get a full list of the flags by running `ferretdb --help`.

```text
~$ ferretdb --help
Usage: ferretdb

Flags:
  -h, --help                             Show context-sensitive help.
      --version                          Print version to stdout and exit.
      --handler="pg"                     Backend handler: 'pg', 'sqlite' ($FERRETDB_HANDLER).
      --mode="normal"                    Operation mode: 'normal', 'proxy', 'diff-normal',
                                         'diff-proxy' ($FERRETDB_MODE).
      --state-dir="."                    Process state directory ($FERRETDB_STATE_DIR).
      --listen-addr="127.0.0.1:27017"    Listen TCP address ($FERRETDB_LISTEN_ADDR).
      --listen-unix=""                   Listen Unix domain socket path ($FERRETDB_LISTEN_UNIX).
      --listen-tls=""                    Listen TLS address ($FERRETDB_LISTEN_TLS).
      --listen-tls-cert-file=""          TLS cert file path ($FERRETDB_LISTEN_TLS_CERT_FILE).
      --listen-tls-key-file=""           TLS key file path ($FERRETDB_LISTEN_TLS_KEY_FILE).
      --listen-tls-ca-file=""            TLS CA file path ($FERRETDB_LISTEN_TLS_CA_FILE).
      --proxy-addr=""                    Proxy address ($FERRETDB_PROXY_ADDR).
      --debug-addr="127.0.0.1:8088"      Listen address for HTTP handlers for metrics, pprof,
                                         etc ($FERRETDB_DEBUG_ADDR).
      --postgresql-url="postgres://127.0.0.1:5432/ferretdb"
                                         PostgreSQL URL for 'pg' handler ($FERRETDB_POSTGRESQL_URL).
      --postgresql-new                   Use new PostgreSQL backend ($FERRETDB_POSTGRESQL_NEW).
      --sqlite-url="file:data/"          SQLite URI (directory) for 'sqlite' handler
                                         ($FERRETDB_SQLITE_URL).
      --log-level="info"                 Log level: 'debug', 'info', 'warn', 'error'
                                         ($FERRETDB_LOG_LEVEL).
      --[no-]log-uuid                    Add instance UUID to all log messages ($FERRETDB_LOG_UUID).
      --[no-]metrics-uuid                Add instance UUID to all metrics ($FERRETDB_METRICS_UUID).
      --telemetry=undecided              Enable or disable basic telemetry. See
                                         https://beacon.ferretdb.io ($FERRETDB_TELEMETRY).
      --test-records-dir=""              Testing: directory for record files
                                         ($FERRETDB_TEST_RECORDS_DIR).
      --test-disable-filter-pushdown     Experimental: disable filter pushdown
                                         ($FERRETDB_TEST_DISABLE_FILTER_PUSHDOWN).
      --test-unsafe-sort-pushdown        Experimental: unsafe sort pushdown
                                         ($FERRETDB_TEST_UNSAFE_SORT_PUSHDOWN).
      --test-telemetry-url="https://beacon.ferretdb.io/"
                                         Telemetry: reporting URL ($FERRETDB_TEST_TELEMETRY_URL).
      --test-telemetry-undecided-delay=1h
                                         Telemetry: delay for undecided state
                                         ($FERRETDB_TEST_TELEMETRY_UNDECIDED_DELAY).
      --test-telemetry-report-interval=24h
                                         Telemetry: report interval
                                         ($FERRETDB_TEST_TELEMETRY_REPORT_INTERVAL).
      --test-telemetry-report-timeout=5s
                                         Telemetry: report timeout
                                         ($FERRETDB_TEST_TELEMETRY_REPORT_TIMEOUT).
      --test-telemetry-package=""        Telemetry: custom package type
                                         ($FERRETDB_TEST_TELEMETRY_PACKAGE).
```

### Step 3: Start FerretDB

We will explore two ways to start FerretDB: via terminal and using a `systemd` file.

#### Start FerretDB via terminal

Before we create a `systemd` for FerretDB, let's try running it via terminal by providing the appropriate flags, including `--postgresql-url`:

```sh
ferretdb --postgresql-url="postgres://username:password@localhost/ferretdb"
```

Update the PostgreSQL credentials to match the one you created before.
Other necessary flags are set to their default values: `--mode="normal"`, `--listen-addr="127.0.0.1:27017"`.

```text
~$ ferretdb --postgresql-url="postgres://username:password@localhost/ferretdb"
2023-10-19T12:48:23.202+0100    INFO    ferretdb/main.go:253    Starting FerretDB v1.12.1...    {"version": "v1.12.1", "commit": "d1486f2b5d86eadfa6d148752b14fdde49cb5db9", "branch": "unknown", "dirty": true, "package": "deb", "debugBuild": false, "buildEnvironment": {"-buildmode":"exe","-compiler":"gc","CGO_ENABLED":"0","GOARCH":"arm64","GOOS":"linux","go.version":"go1.21.2","vcs":"git","vcs.time":"2023-10-10T12:15:18Z"}, "uuid": "d66f7807-aec3-406a-9779-2b5ec190e65a"}
2023-10-19T12:48:23.214+0100    INFO    telemetry   telemetry/reporter.go:148   The telemetry state is undecided; the first report will be sent in 1h0m0s. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
2023-10-19T12:48:23.217+0100    INFO    debug   debug/debug.go:86   Starting debug server on http://127.0.0.1:8088/
2023-10-19T12:48:23.217+0100    INFO    listener    clientconn/listener.go:97   Listening on TCP 127.0.0.1:27017 ...
```

#### Start FerretDB using a `systemd` file

Creating a `systemd` service file for FerretDB will allow the database be managed by the `systemd` system and service manager.

To create a `systemd` service file, open a new file in the `/etc/systemd/system` directory.

```sh
sudo nano /etc/systemd/system/ferretdb.service
```

Add the following content to the file:

```text
[Unit]
Description=FerretDB service
After=network-online.target
Wants=network-online.target

[Service]
User=ferret
ExecStart=/usr/bin/ferretdb --postgresql-url="postgres://username:password@127.0.0.1:5432/ferretdb"
Restart=always

[Install]
WantedBy=multi-user.target
```

Of course, this is just a basic setup, you might want to include additional details or security measures, including setting up TLS connections, or other config settings.
After creating the service file, reload the `systemd` configurations:

```sh
sudo systemctl daemon-reload
```

Then, enable the service so that it starts on boot:

```bash
sudo systemctl enable ferretdb
```

Start the service:

```sh
sudo systemctl start ferretdb
```

Check the status of your service to be sure it's running:

```sh
sudo systemctl status ferretdb
```

To check the logs for the FerretDB service, run this command:

```sh
sudo journalctl -u ferretdb -f
```

Now that FerretDB is running, let's connect to it via `mongosh`.

### Step 4: Connect via `mongosh`

With FerretDB running, open another terminal and connect to FerretDB via `mongosh` using your connection URI.
Use the same username and password credentials for the connection.

```sh
mongosh "mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN"
```

This will connect to your FerretDB instance.

```text
~$ mongosh "mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN"
Current Mongosh Log ID: 65304b6bb7bd5de804c5e34e
Connecting to:      mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.0.2
Using MongoDB:      6.0.42
Using Mongosh:      2.0.2

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2023-10-18T21:17:31.774Z: Powered by FerretDB v1.12.1 and PostgreSQL 14.9.
   2023-10-18T21:17:31.774Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2023-10-18T21:17:31.774Z: The telemetry state is undecided.
   2023-10-18T21:17:31.774Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb>
```

### Step 5: Insert documents into FerretDB

FerretDB allows you to run MongoDB commands and queries, so you can easily go ahead and try out your favorite MongoDB operations.

Let's insert a document into FerretDB via mongosh:

```text
ferretdb> db.playerstats.insertOne({"futbin_id" : 7,"player_name" : "Vieri", "player_extended_name" : "Christian Vieri", "quality" : "Gold - Rare", "revision" : "Icon", "overall" : 88 })
{
  acknowledged: true,
  insertedId: ObjectId("65305ac1b116d06ab74d6a33")
}
ferretdb> db.playerstats.find()
[
  {
    _id: ObjectId("65305ac1b116d06ab74d6a33"),
    futbin_id: 7,
    player_name: 'Vieri',
    player_extended_name: 'Christian Vieri',
    quality: 'Gold - Rare',
    revision: 'Icon',
    overall: 88
  }
]
ferretdb>
```

We can also inspect this data on the PostgreSQL backend to see how it's depicted.

```text
ferretdb=> set search_path to ferretdb;
SET
ferretdb=> \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | username
 ferretdb | playerstats_0f3be573        | table | username
 ferretdb | test_afd071e5               | table | username
(3 rows)

ferretdb=> SELECT * from playerstats_0f3be573;
```

We can see the result of the document we just created and how it appears in PostgreSQL.

```text
                                                                                                                                                                                                                                                         _jsonb
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "overall": {"t": "int"}, "quality": {"t": "string"}, "revision": {"t": "string"}, "futbin_id": {"t": "int"}, "player_name": {"t": "string"}, "player_extended_name": {"t": "string"}}, "$k": ["_id", "futbin_id", "player_name", "player_extended_name", "quality", "revision", "overall"]}, "_id": "65305ac1b116d06ab74d6a33", "overall": 88, "quality": "Gold - Rare", "revision": "Icon", "futbin_id": 7, "player_name": "Vieri", "player_extended_name": "Christian Vieri"}
(1 row)
```

## Bonus: Run FerretDB with the SQLite backend on Ubuntu

Apart from the PostgreSQL backend, FerretDB also gives you the option to use the SQLite backend.
But unlike PostgreSQL, SQLite is serverless – it operates without the need for a separate server process or system.

You can run FerretDB with the SQLite backend by providing the `--handler="sqlite"` flag when running FerretDB.

With FerretDB installed, run this command in your terminal to start FerretDB.
You may need to create a folder named "data" before running the command if it doesn't exist already.

```sh
ferretdb --handler="sqlite"
```

Then you can connect to the instance via the connection URI.

```text
mongodb://localhost:27017/ferretdb
```

At present, FerretDB does not support authentication for SQLite, but you can [track its implementation here](https://github.com/FerretDB/FerretDB/issues/3008).

```text
~$ mongosh "mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN"
Current Mongosh Log ID: 65304b6bb7bd5de804c5e34e
Connecting to:      mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.0.2
Using MongoDB:      6.0.42
Using Mongosh:      2.0.2

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2023-10-18T21:17:31.774Z: Powered by FerretDB v1.12.1 and PostgreSQL 14.9.
   2023-10-18T21:17:31.774Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2023-10-18T21:17:31.774Z: The telemetry state is undecided.
   2023-10-18T21:17:31.774Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb>
```

## Conclusion

In this detailed tutorial, we've guided you through the process of installing FerretDB on Ubuntu, setting up and using the different configuration flags.
Besides that, it also includes how to set up the PostgreSQL backend and connect to your FerretDB instance via mongosh.

As a bonus for those interested in experimenting with the SQLite backend, the tutorial includes a basic section on how to set it up.
You can also check out [how to start FerretDB locally on Docker](https://blog.ferretdb.io/how-to-start-ferretdb-locally-with-docker/).

As an open source project, FerretDB welcomes all [contributions](https://docs.ferretdb.io/contributing/).
You can contribute to the development of FerretDB by contributing to code and documentation, submitting bug reports and feature requests, and even a writing a blog post, so if you would like to publish an article on FerretDB, please contact us on [any of our community channels](https://docs.ferretdb.io/#community).
