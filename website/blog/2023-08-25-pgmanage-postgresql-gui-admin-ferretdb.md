---
slug: pgmanage-postgresql-gui-admin-ferretdb
title: 'Using PgManage as PostgreSQL GUI Admin for FerretDB'
authors: [alex]
description: >
  In this blog post, we’ll be exploring ways to manage your FerretDB data in Postgres using PgManage GUI admin tool.
image: /img/blog/ferretdb-pgmanage.jpg
tags: [compatible applications, tutorial, postgresql tools]
---

![Using PgManage as PostgreSQL GUI Admin for FerretDB](/img/blog/ferretdb-pgmanage.jpg)

One of the biggest advantages of using [FerretDB](https://www.ferretdb.io/) – the truly open source alternative alternative to MongoDB – is that you can manage, scale, and deploy it on almost any PostgreSQL architecture.

<!--truncate-->

And that's the case with [PgManage](https://www.commandprompt.com/products/pgmanage/).

PgManage is a product offering from [CommandPrompt, Inc.](https://www.commandprompt.com/) that provides an open source GUI management platform for Postgres.
It is able to query and manage several database management systems including PostgreSQL, MySQL, SQLite, and Oracle.
It builds on the great work previously done on [OmniDB (now inactive)](https://github.com/OmniDB/OmniDB) to provide an interface for managing queries, monitoring databases, server configurations, and more.

For FerretDB users and enthusiasts, PgManage can be a great option to consider as a Postgres GUI admin tool for managing your FerretDB data.

In this blog post, we'll be exploring ways to manage your FerretDB data in Postgres using PgManage GUI admin tool.

## Installation and setup

This tutorial was tested on a macOS system.
However, we've provided links to the installation guide for other operating systems.

### Installing PgManage

You can install PgManage using the installation guide here.
It details the different installation options for Mac, Windows, or Linux systems.

Note that in desktop-app mode, Pgmanage does not require setting up any accounts.
What needs to be done though is setting up a master password for credential encryption.

### Setting up FerretDB

For the FerretDB setup, we'll be using the Docker setup guide from the documentation – [check it out here](https://docs.ferretdb.io/quickstart-guide/docker/).

However, we'll be making a few changes to the `docker-compose` yaml file by exposing the `postgres` port using a port-forward that opens up the 5432 port and make it accessible outside the container.
This will make it easy for us to connect to the port via an application like PgManage.
Also, don't forget to update the authentication credentials for the username and password.

```yaml
services:
  postgres:
    image: postgres
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=ferretdb
    volumes:
      - ./data:/var/lib/postgresql/data
    ports:
      - 5432:5432

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://postgres:5432/ferretdb

networks:
  default:
    name: ferretdb
```

You can go ahead to start the services with `docker compose up -d`.

#### Basic operations in FerretDB

FerretDB enables you to perform MongoDB queries and commands and stores the data in Postgres.
For the purpose of this tutorial, we'll be adding sample data to the FerretDB database.

If you have `mongosh`, you can run FerretDB using this connection URI with the right authentucation credentials:

```sh
mongodb://username:password@127.0.0.1/ferretdb?authMechanism=PLAIN
```

But if you don't have `mongosh`, you can run it through a temporary MongoDB container:

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
  "mongodb://username:password@ferretdb/ferretdb?authMechanism=PLAIN"
```

We'll also go ahead and add a sample data to showcase on PgManage.
A terminal view of this would look like this:

```text
$ docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
  "mongodb://username:password@ferretdb/ferretdb?authMechanism=PLAIN"
Current Mongosh Log ID: 64e7657774614fd82dd9853c
Connecting to:          mongodb://<credentials>@ferretdb/ferretdb?authMechanism=PLAIN&directConnection=true&appName=mongosh+1.10.1
Using MongoDB:          6.0.42
Using Mongosh:          1.10.1

For mongosh info see: https://docs.mongodb.com/mongodb-shell/


To help improve our products, anonymous usage data is collected and sent to MongoDB periodically (https://www.mongodb.com/legal/privacy-policy).
You can opt-out by running the disableTelemetry() command.

------
   The server generated these startup warnings when booting
   2023-08-24T14:13:12.343Z: Powered by FerretDB v1.8.0 and PostgreSQL 15.4.
   2023-08-24T14:13:12.343Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2023-08-24T14:13:12.343Z: The telemetry state is undecided.
   2023-08-24T14:13:12.344Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb> db.testdata.insertMany([{a: "b", b: 34}, {a: "c", b: [3, 5]}])
{
  acknowledged: true,
  insertedIds: {
    '0': ObjectId("64e765e074614fd82dd9853d"),
    '1': ObjectId("64e765e074614fd82dd9853e")
  }
}
```

### Connecting PgManage with FerretDB

You can connect PgManage with the `ferretdb` database using the exposed Postgres port `5432`, along with the authentication credentials for the user.

Once you've signed in to PgManage, navigate to the "Connections" tab, and click "Manage connections"; go ahead and add a new connection.

![Connection setttings](/img/blog/ferretdb-pgmanage/connection-settings.png)

Test the connection to be sure it's successful, and then save changes.
With FerretDB connection to PgManage, you'll be able to access all the data in `ferretdb`.

As you can see in the image, PgManage displays all databases and tables in `ferretdb`.

![Import all FerretDB data](/img/blog/ferretdb-pgmanage/imported-data.png)

To view the sample data we just added, we can navigate to the `ferretdb` database (or in this case `ferretdb` schema), and then query all the data in the `testdata` table.
we can see the JSONB view of the data in Pgmanage.

![Querying data in PgManage](/img/blog/ferretdb-pgmanage/select-table-data.png)

## Performing server configuration using PgManage for FerretDB

Using PgManage, users can easily manage their database, access and modify server configuration, while leveraging its user-friendly interface and direct `ALTER SYSTEM` commands.

This straightforward approach eliminates the need to manually edit the `postgresql.conf` file.

To do this, right click the root node in the database object tree.
Then select "Server configuration" from the dropdown menu.

![Server configuration](/img/blog/ferretdb-pgmanage/server-configuration.png)

Then you can proceed to configure the server settings as you would in a regular PostgreSQL database, all within the PgManage UI.

![Server configuration UI](/img/blog/ferretdb-pgmanage/server-configuration-ui.png)

Besides, users can track and revert changes using PgManage's "Config History" dropdown.

## Conclusion

FerretDB gives users the chance to use MongoDB's user-friendly queries and syntax while having the reliability and robustness of a PostgreSQL backend - basically the best of both worlds.

And for PostgreSQL administrators, PgManage can be an incredible tool to administer a FerretDB database, providing a intuitive to manage, scale, and configure your servers – [check it out here to get started](https://pgmanage.readthedocs.io/).

The possibilities are endless, and we can't wait to find out what you do with FerretDB and PgManage.
For more information on FerretDB, [see our documentation](https://github.com/FerretDB/FerretDB) and [GitHub page](https://github.com/FerretDB/FerretDB).

If you do have any questions or feedback on using FerretDB, please reach out to us – we'd be happy to chat with you.
