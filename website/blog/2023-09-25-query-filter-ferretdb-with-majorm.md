---
slug: query-filter-ferretdb-with-majorm
title: 'MongoDB GUI: Query and Filter FerretDB Instance with MajorM'
authors: [alex]
description: >
  Explore and use FerretDB with MajorM, and explore ways to leverage FerretDB for MongoDB GUI applications.
image: /img/blog/majorm.jpg
tags:
  [
    tutorial,
    mongodb compatible,
    mongodb gui,
    compatible applications,
    document databases
  ]
---

![Query FerretDB instance with MajorM](/img/blog/majorm.jpg)

With FerretDB, you're never short of MongoDB GUI options to choose from, and in this blog post, we'll be diving into one of them: MajorM.

<!--truncate-->

Often, with proprietary software solutions like MongoDB, many find themselves constrained in terms of flexibility and costs and are always at [risk of vendor lock-in](https://blog.ferretdb.io/5-ways-to-avoid-database-vendor-lock-in/).
That's why _truly_ open-source software like FerretDB is so important and attractive!

FerretDB won't constrict you or lock you in.
What's even better, you get to use many of your favorite MongoDB tools with FerretDB.
Because of that, you can easily leverage a MongoDB GUI like MajorM for FerretDB, just as you would for MongoDB.

MajorM is a simple MongoDB GUI desktop app that provides fast and complex query searches, visual query builder, and drag-and-drop functionality.
It also provides a `mongoshell` UI for advanced operations on your FerretDB instance.

This guide is just one of a series of blog posts focusing on using MongoDB GUI tools to query and filter FerretDB instances.
See other blog posts below:

- [Using FerretDB 1.0 with Studio 3T](https://blog.ferretdb.io/using-ferretdb-with-studio-3t/)
- [Using Mingo to Analyze and Visualize FerretDB Data](https://blog.ferretdb.io/using-mingo-analyze-visualize-ferretdb-data/)
- [MongoDB GUI: Using FerretDB with NoSQLBooster](https://blog.ferretdb.io/mongodb-gui-using-ferretdb-nosqlbooster/)

## Prerequisites

Before we get started, ensure to have the following set up before proceeding to other parts of the blog post.

- [Download the MajorM desktop application from the website](https://www.majorm.app/)
- [Install Docker](https://docs.docker.com/get-docker/) to set up the FerretDB instance
- mongosh

## Set up FerretDB

Follow this Docker installation guide to launch an instance of FerretDB and have your connection string ready.

For this tutorial, we'll be setting up our docker-compose file with the following configurations.

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

Here, we've set up the `ferretdb` service to connect to the `postgres` service, as it translates MongoDB commands into SQL for the underlying PostgreSQL backend.
Then we'll launch the services with `run docker compose up -d`.

### Set up the MajorM connection to FerretDB

Launch your installed MajorM application.

Click "New Connection" to connect to FerretDB.

![New Connection section](/img/blog/ferretdb-majorm/1-connections-tab.png)

This takes you to the connection window, where you'll need to specify the connection string to the FerretDB instance.

![New Connection window](/img/blog/ferretdb-majorm/2-connections-window.png)

In the connection window, specify the name and connection URI for the FerretDB instance: `mongodb://username:password@127.0.0.1:27017/?authMechanism=PLAIN`.
Then, test the connection to be sure it's valid.
If successful, save the connection.

![Connection window with URI](/img/blog/ferretdb-majorm/3-connection-uri.png)

:::note
For the MongoShell in MajorM to work, you'll need to specify your `mongosh` binary path.
If you have `mongosh` installed.
You can get this by running `which mongosh` on macOS and other Linux systems or `where mongosh` for Windows systems.
:::

Once the connection is successful, MajorM will display all the databases and collections in that instance.

## Running FerretDB operations in MajorM

Great!
Let's run some basic queries on the FerretDB instance.
We'll be querying the data in the `supply` collection of the `ferretdb` database.

### Test1: Get a single data record using Mongo Shell

MajorM provides the option to use Mongo shell to execute queries and scripts.
From the Mongo shell, let's run `db.supply.findOne()` to retrieve a single record from the collection.

Note that you will need to set the database to `ferretdb`.

![Set database to `ferretdb`](/img/blog/ferretdb-majorm/4-set-database-ferretdb.png)

The MajorM UI will display the data record once you click execute.

![findOne command](/img/blog/ferretdb-majorm/5-findone-command.png)

### Test2: Display all data records

You can also get all the records in the collection by navigating to `supply` collection and double-clicking.
This action displays all the data in the `supply` collection as you can see in the image below:

![List of all data records](/img/blog/ferretdb-majorm/6-data-records.png)

### Test3: Use visual query builder to query collection

Here, we'll use the MajorM's Query Builder to filter the FerretDB collection.

![Query Builder](/img/blog/ferretdb-majorm/7-query-builder.png)

Using the visual query builder, drag and drop fields from the `supply` collection into the builder.

![Drag and drop fields to Query Builder](/img/blog/ferretdb-majorm/8-drag-and-drop.gif)

### Test4: Use sort and limit to filter

You can filter the response from the last example further by applying `sort` and `limit` on the returned documents.

![Sort and limit](/img/blog/ferretdb-majorm/9-sort-limit.png)

### Test5: Run aggregation operations using Mongo shell

You can also take advantage of the Mongo shell in MajorM to run all the supported aggregation operations in FerretDB.
For instance, let's select the top 5 customers by the total quantity of books purchased.

![Drag and drop fields to Query Builder](/img/blog/ferretdb-majorm/10-aggregation-operation.png)

## Query FerretDB instance with MajorM

Sure, it's never easy switching to a new technology; then again, ensuring that it works with all your other applications is another matter entirely.
When moving away from MongoDB to adopt FerretDB, knowing that you can use the same MongoDB commands and leverage just about any tool as you would in MongoDB is a huge perk.

MongoDB GUI applications like MajorM are very useful when managing large amounts of data.
They offer users simple, and engaging UI, making it easy to run queries in FerretDB, and we can't wait for you to try them out.

If you're interested in learning more about the FerretDB project, check [our GitHub repo here](https://github.com/FerretDB/FerretDB).

Feel free to message us on any of [our community channels](https://docs.ferretdb.io/#community) if you have any questions, suggestions, or interest in FerretDB – we look forward to hearing from you.
