---
slug: install-ferretdb-mongodb-alternative-ubuntu
title: 'How to Install FerretDB on Ubuntu'
author: Klinsmann Öteyo
description: >
  This is a guide on how to install FerretDB - the MongoDB alternative - on Ubuntu.
image: /img/blog/install-ferretdb-ubuntu.jpg
keywords: [FerretDB, mongodb alternative]
---

<head>
  <link rel="canonical" href="https://tutornix.com/how-to-install-ferretdb-mongodb-alternative-on-ubuntu-22-0420-04/" />
</head>

![How to install FerretDB on Ubuntu](/img/blog/install-ferretdb-ubuntu.jpg)

*This is a duplicate post from [Tutornix](https://tutornix.com/how-to-install-ferretdb-mongodb-alternative-on-ubuntu-22-0420-04/) about installing FerretDB on Ubuntu; we are grateful to the author for the permission to repost this on our blog.*

<!--truncate-->

Welcome to this guide on how to install [FerretDB](https://www.ferretdb.io/) - MongoDB Alternative - on Ubuntu.
But before we dive into the nub of this matter, we need to know what FerretDB is!

Databases are majorly categorised as Relational and Non-relational.
Today, our key focus is on non-relational databases(also known as NoSQL).
These databases do not observe the traditional tabular relational database model of rows and columns with pre-defined schemas, but instead use a variety of data models, such as key-value, document, column-family, and graph models, to store and manage data.

They are often used to store and manage unstructured or semi-structured data, such as text, images, videos, social media posts, and sensor data, that may not fit well into traditional relational databases.
They are known for better scalability, performance, and availability than traditional relational databases, particularly for web-scale applications with high volumes of data and traffic.
The most popular NoSQL databases are [MongoDB](https://www.mongodb.com/), [Redis](https://redis.io/), [Couchbase](https://www.couchbase.com/), [Cassandra](https://cassandra.apache.org/_/index.html), [Neo4j](https://neo4j.com/) etc.

FerretDB is a free and open-source alternative to MongoDB, built on [Postgres](https://www.postgresql.org/).
It translates the MongoDB wire protocol queries to SQL with the help of PostgreSQL as the database engine.
FerretDB was introduced after the popular MongoDB deviated from its open-source roots and switched to the SSPL license.
It acts as a drop-in replacement for MongoDB users looking for an open-source alternative to MongoDB.

One of the amazing features is that FerretDB is compatible with MongoDB drivers and can serve as a direct replacement for MongoDB 6.0+.
It is cross-platform as it can be installed on Linux, MacOS and Windows systems.
There is also a docker image for those who want to spin it in containers.

To run FerretDB in Docker containers, follow the guide below:

- [Run FerretDB MongoDB Alternative in Docker Containers](https://tutornix.com/run-ferretdb-mongodb-alternative-in-docker-containers/)

Now let's plunge in!

## Step 1 – Install and Configure PostgreSQL on Ubuntu

FerretDB uses PostgreSQL as the database engine.
For that reason, we need to install it on our system.

Before we begin, you need to ensure that your system and all the installed packages are updated to their latest stable versions:

```sh
sudo apt update
sudo apt upgrade
```

Once the system has been updated, install PostgreSQL

```sh
sudo apt install postgresql -y
```

After installing, ensure that the service is runnning:

```sh
$ systemctl status postgresql
● postgresql.service - PostgreSQL RDBMS
     Loaded: loaded (/lib/systemd/system/postgresql.service; enabled; vendor preset: enabled)
     Active: active (exited) since Fri 2023-04-14 18:23:20 EAT; 1min 7s ago
    Process: 4519 ExecStart=/bin/true (code=exited, status=0/SUCCESS)
   Main PID: 4519 (code=exited, status=0/SUCCESS)
        CPU: 1ms
....
```

Connect to the instance:

```sh
sudo -u postgres psql
```

Create the FerretDB user:

```js
CREATE USER ferretuser WITH PASSWORD 'Passw0rd!';
```

Create a database for the user:

```sql
CREATE DATABASE ferretuser OWNER ferretuser;
```

Once created, exit the shell:

```sh
\q
```

Create the user on your system:

```sh
sudo adduser ferretuser
```

Verify if you can connect to the database using the user:

```sh
$ sudo -u ferretuser psql
could not change directory to "/home/ubuntu22": Permission denied
psql (14.7 (Ubuntu 14.7-0ubuntu0.22.04.1))
Type "help" for help.

ferretuser=> \q
```

## Step 2 – Download and Install FerretDB on Ubuntu

To download the latest FerretDB release, visit the official [GitHub Release page](https://github.com/FerretDB/FerretDB/releases).
Alternatively, you can pull the latest release using the command:

```sh
VER=v1.2.0
wget https://github.com/FerretDB/FerretDB/releases/download/$VER/ferretdb.deb
```

Once downloaded, install FerretDB using the command:

```sh
sudo apt install ./ferretdb.deb
```

Sample Output:

```sh
Reading package lists... Done
Building dependency tree
Reading state information... Done
Note, selecting 'ferretdb' instead of './ferretdb.deb'
The following package was automatically installed and is no longer required:
  gir1.2-goa-1.0
Use 'sudo apt autoremove' to remove it.
The following NEW packages will be installed:
  ferretdb
0 upgraded, 1 newly installed, 0 to remove and 95 not upgraded.
Need to get 0 B/15.1 MB of archives.
After this operation, 30.1 MB of additional disk space will be used.
Get:1 /home/ubuntu/ferretdb.deb ferretdb amd64 0.0.0~rc0 [15.1 MB]
Selecting previously unselected package ferretdb.
(Reading database ... 234140 files and directories currently installed.)
Preparing to unpack /home/ubuntu/ferretdb.deb ...
Unpacking ferretdb (0.0.0~rc0) ...
Setting up ferretdb (0.0.0~rc0) ...
```

Check the installed version:

```sh
$ ferretdb --version
version: v1.2.0
commit: 3153c8fbf185126af1fe8fb364ac166d2287d093
branch: unknown
dirty: true
package: deb
debugBuild: false
```

## Step 3 – Start and Enable FerretDB on Ubuntu

FerretDB takes several arguments that include:

- **-h, –help**: Show context-sensitive help.
- _–version_: Print version to stdout and exit.
- **–handler="pg"** : Backend handler: 'dummy', 'pg', 'tigris' \
  ($FERRETDB_HANDLER).
- –**-mode="normal"**: Operation mode: 'normal', 'proxy', 'diff-normal', \
  'diff-proxy' ($FERRETDB_MODE).
- **–state-dir="."**: Process state directory ($FERRETDB_STATE_DIR).
- **–listen-addr="127.0.0.1:27017″**: Listen TCP address ($FERRETDB_LISTEN_ADDR).
- **–listen-unix=""**: Listen Unix domain socket path ($FERRETDB_LISTEN_UNIX).
- **–listen-tls=""**: Listen TLS address ($FERRETDB_LISTEN_TLS).
- **–listen-tls-cert-file=""** : TLS cert file path ($FERRETDB_LISTEN_TLS_CERT_FILE).
- **–listen-tls-key-file=""** : TLS key file path ($FERRETDB_LISTEN_TLS_KEY_FILE).
- **–listen-tls-ca-file=""** : TLS CA file path ($FERRETDB_LISTEN_TLS_CA_FILE).
- **–postgresql-url="postgres://127.0.0.1:5432/ferretdb"** : PostgreSQL URL for 'pg' handler

Since FerretDB has not been started, we will create a systemd service file with all the desired variables:

```sh
sudo vim /etc/systemd/system/ferretdb.service
```

In the file, add the below lines

```js
[Unit]
Description=FerretDB
After=network-online.target

[Service]
Type=simple
ExecStart=/bin/bash -c "ferretdb --mode="normal" --listen-addr="127.0.0.1:27017" --postgresql-url='postgres://ferretuser:Passw0rd!@127.0.0.1:5432/ferretuser'"
Restart=always
RestartSec=2
TimeoutStopSec=5
SyslogIdentifier=ferretdb

[Install]
WantedBy=multi-user.target
```

In the above command, remember to replace the credentials for **PostgreSQL** before you proceed with the below command to reload the system daemon:

```sh
sudo systemctl daemon-reload
```

Start and enable the service:

```sh
sudo systemctl enable ferretdb
sudo systemctl start ferretdb
```

Check the status of the service:

```sh
$ systemctl status ferretdb
●  ferretdb.service - FerretDB
     Loaded: loaded (/etc/systemd/system/ferretdb.service; enabled; vendor preset: enabled)
     Active: active (running) since Fri 2023-04-14 18:35:07 EAT; 5s ago
   Main PID: 9424 (ferretdb)
      Tasks: 7 (limit: 4629)
     Memory: 3.9M
        CPU: 10ms
     CGroup: /system.slice/ferretdb.service
             └─9424 ferretdb --mode=normal --listen-addr=127.0.0.1:27017 "--postgresql-url=postgres://ferretuser:Passw0rd!@127.0.0.1:5432/ferretuser"

Apr 14 18:35:07 tutornix.lab systemd[1]: Started FerretDB.
Apr 14 18:35:07 tutornix.lab ferretdb[9424]: 2023-04-14T18:35:07.544+0300        INFO        ferretdb/main.go:231        Starting FerretDB v1.0.0...        {"version": "v1.0.0", "commit": ">
Apr 14 18:35:07 tutornix.lab ferretdb[9424]: 2023-04-14T18:35:07.545+0300        INFO        listener        clientconn/listener.go:95        Listening on TCP 127.0.0.1:27017 ...
Apr 14 18:35:07 tutornix.lab ferretdb[9424]: 2023-04-14T18:35:07.545+0300        INFO        listener        clientconn/listener.go:183        Waiting for all connections to stop...
Apr 14 18:35:07 tutornix.lab ferretdb[9424]: 2023-04-14T18:35:07.546+0300        INFO        debug        debug/debug.go:95        Starting debug server on http://127.0.0.1:8088/
```

## Step 4 – Use FerretDB on Ubuntu

Once installed and started, FerretDB can be used just like MongoDB.
You can connect to it using the Mongo shell.

You need to have the Mongo Shell installed on your system before you proceed.
First, add the MongoDB repo to your system.

```sh
sudo apt install wget curl gnupg2 software-properties-common apt-transport-https ca-certificates lsb-release -y
curl -fsSL https://www.mongodb.org/static/pgp/server-6.0.asc|sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/mongodb-6.gpg
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
```

Install the Mongo Shell on Ubuntu:

```sh
sudo apt update && sudo apt install mongodb-mongosh
```

Now connect the FerretDB:

```sh
mongosh
```

Sample Output:

![Run FerretDB on Ubuntu](/img/blog/ferretdb-ubuntu/run-ferretdb-ubuntu.webp)

Once connected, switch to the database created on PostgreSQL

```sh
test> use ferretuser
switched to db ferretuser
ferretuser> show dbs
public  0 B
ferretuser>
```

Add tables to the database:

```js
db.userdetails.insertOne({
  F_Name: 'Tutornix',
  L_NAME: 'Home',
  ID_NO: '124345',
  AGE: '70',
  TEL: '25465445642'
})
```

View the added collections:

```js
> show collections
userdetails
```

This added table can also be viewed on the PostgreSQL database:

```sh
sudo -u ferretuser psql
```

Now view if the tables exist here too:

![View FerretDB tables](/img/blog/ferretdb-ubuntu/view-ferretdb-postgres.webp)

## Verdict

That marks the end of this amazing guide on how to install FerretDB on Ubuntu.
You can now enjoy the awesomeness of FerretDB, a free and open-source alternative to MongoDB.
I hope this was helpful.
