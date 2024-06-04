---
slug: run-mongodb-workloads-aiven-postgresql
title: 'Run MongoDB Workloads on Aiven for PostgreSQL'
authors: [alex]
description: >
  In this blog, we'll show you how to run MongoDB workloads on Aiven for PostgreSQL using FerretDB.
image: /img/blog/ferretdb-aiven.jpg
tags: [tutorial, postgresql tools, open source, cloud]
---

![Run MongoDB Workloads on Aiven for PostgreSQL](/img/blog/ferretdb-aiven.jpg)

Adopting open-source solutions is a strategic move for modern businesses.
For those looking to migrate from MongoDB, [FerretDB](https://www.ferretdb.com/) is a truly open source alternative that lets you run MongoDB workloads on [PostgreSQL](https://www.postgresql.org/).

<!--truncate-->

With [Aiven for PostgreSQL](https://aiven.io/postgresql), you can set up a reliable backend for FerretDB to run your MongoDB workloads.
Aiven provides a unified, cloud-agnostic platform that lets you tap into several robust features, Postgres extensions, and integration into your data infrastructure.

This blog dives into how you can add MongoDB compatibility to your Postgres service on Aiven.

## Prerequisites

Before you start, ensure you have the following set up:

- [Aiven account](https://aiven.io/)
- `psql`
- [Docker](https://www.docker.com/)
- `mongosh`

## Set up Aiven for PostgreSQL

First, create an Aiven account if you don't have one.
From the Aiven dashboard, create a PostgreSQL service.
Learn more about setting up a PostgreSQL service in the [Aiven documentation](https://aiven.io/docs/products/postgresql).

![Aiven for PostgreSQL service](/img/blog/aiven-postgres.png)

Once it's set up, your Aiven for PostgreSQL connection string should look like this:

```text
postgres://<username>:<password>@<host>:port/database?sslmode=require
```

Connect to the connection string via `psql`, and create a `ferretdb` database that'll hold your FerretDB data.

```text
defaultdb=> CREATE DATABASE ferretdb owner <username>;
CREATE DATABASE
```

## Run FerretDB via Docker

Let's set up the FerretDB instance.
To do that, you need to specify the `FERRETDB_POSTGRESQL_URL` environment variable or `--postgresql-url` flag.
Along with that, specify the Postgres `username`/`password` credentials for your database.

Using Docker, run the following command to set it up.

```sh
docker run -e FERRETDB_POSTGRESQL_URL='postgres://<username>:<password>@<host>:port/ferretdb?sslmode=require'  -p 27017:27017
 ghcr.io/ferretdb/ferretdb
```

Ensure to replace `<username>`, `<password>`, `<host>`, and `<port>` with your Aiven for PostgreSQL credentials.

After that, connect to FerretDB via `mongosh` using the below MongoDB URI format:

```sh
mongosh 'mongodb://<username>:<password>@127.0.0.1:27017/ferretdb?authMechanism=PLAIN'
```

Awesome!
Now let's go ahead to run a couple of MongoDB operations using FerretDB.

### Perform MongoDB operations on FerretDB

As a fan of astronomy, let's play around with some arbitrary astronomy data containing details about different stars.

#### Insert data

Start by inserting the following data into the `astronomy` collection.

```json5
db.astronomy.insertMany([
  {
    name: "Alpha Centauri A",
    type: "Star",
    distance_from_earth: 4.37,
    mass: 2.187e30,
    diameter: 1214000,
    constellation: "Centaurus"
  },
  {
    name: "Alpha Centauri B",
    type: "Star",
    distance_from_earth: 4.37,
    mass: 1.804e30,
    diameter: 865000,
    constellation: "Centaurus"
  },
  {
    name: "Proxima Centauri",
    type: "Star",
    distance_from_earth: 4.24,
    mass: 2.446e29,
    diameter: 200000,
    constellation: "Centaurus"
  },
 {
   name: "Betelgeuse",
   type: "Star",
   distance_from_earth: 642.5,
   mass: 2.78e31,
   diameter: 1.2e9,
   constellation: "Orion"
 },
 {
   name: "Vega",
   type: "Star",
   distance_from_earth: 25.04,
   mass: 4.074e30,
   diameter: 2440000,
   constellation: "Lyra"
 }
]);
```

So we have an `astronomy` collection containing 5 documents with different star data, including their mass, diameter, and distance from Earth.
Note that the mass is in kilograms, the diameter in kilometers, and the distance from Earth in light years.

#### Query the data

Let's query the data for the stars in the constellation `Centaurus`.

```text
ferretdb> db.astronomy.find({ constellation: "Centaurus" });
[
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b6'),
    name: 'Alpha Centauri A',
    type: 'Star',
    distance_from_earth: 4.37,
    mass: 2.187e+30,
    diameter: 1214000,
    constellation: 'Centaurus'
  },
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b7'),
    name: 'Alpha Centauri B',
    type: 'Star',
    distance_from_earth: 4.37,
    mass: 1.804e+30,
    diameter: 865000,
    constellation: 'Centaurus'
  },
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b8'),
    name: 'Proxima Centauri',
    type: 'Star',
    distance_from_earth: 4.24,
    mass: 2.446e+29,
    diameter: 200000,
    constellation: 'Centaurus'
  }
]
```

We got three results back: Alpha Centauri A, Alpha Centauri B, and Proxima Centauri.

#### Query data with an operator

Next, let's query for stars with a mass of less than `1e30` kg.

```text
ferretdb> db.astronomy.find({ mass: { $lt: 1e30 } });
[
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b8'),
    name: 'Proxima Centauri',
    type: 'Star',
    distance_from_earth: 4.24,
    mass: 2.446e+29,
    diameter: 200000,
    constellation: 'Centaurus'
  }
]
```

Now we know the smallest star in the collection is Proxima Centauri.

#### Sort the data

You may also sort the documents by their distance from Earth.

```text
ferretdb> db.astronomy.find({}).sort({ distance_from_earth: 1 });
[
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b8'),
    name: 'Proxima Centauri',
    type: 'Star',
    distance_from_earth: 4.24,
    mass: 2.446e+29,
    diameter: 200000,
    constellation: 'Centaurus'
  },
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b6'),
    name: 'Alpha Centauri A',
    type: 'Star',
    distance_from_earth: 4.37,
    mass: 2.187e+30,
    diameter: 1214000,
    constellation: 'Centaurus'
  },
  {
    _id: ObjectId('665f00fb2d149942b1b2b4b7'),
    name: 'Alpha Centauri B',
    type: 'Star',
    distance_from_earth: 4.37,
    mass: 1.804e+30,
    diameter: 865000,
    constellation: 'Centaurus'
  },
  {
    _id: ObjectId('665f027e2d149942b1b2b4ba'),
    name: 'Vega',
    type: 'Star',
    distance_from_earth: 25.04,
    mass: 4.074e+30,
    diameter: 2440000,
    constellation: 'Lyra'
  },
  {
    _id: ObjectId('665f027e2d149942b1b2b4b9'),
    name: 'Betelgeuse',
    type: 'Star',
    distance_from_earth: 642.5,
    mass: 2.78e+31,
    diameter: 1200000000,
    constellation: 'Orion'
  }
]
```

### View data in `psql`

FerretDB lets you perform a wide range of MongoDB operations — from simple queries to complex aggregations — on your Postgres database.
Want to see how this all looks in Postgres?

Connect to your Postgres instance via `psql` to see the data.

```text
ferretdb=> SET SEARCH_PATH to ferretdb;
SET
ferretdb=> \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | avnadmin
 ferretdb | astronomy_5f9854f1          | table | avnadmin
(2 rows)
ferretdb=> SELECT * FROM astronomy_5f9854f1;
                                                                                                                                                                                                                                                           _jsonb
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "mass": {"t": "double"}, "name": {"t": "string"}, "type": {"t": "string"}, "diameter": {"t": "int"}, "constellation": {"t": "string"}, "distance_from_earth": {"t": "double"}}, "$k": ["_id", "name", "type", "distance_from_earth", "mass", "diameter", "constellation"]}, "_id": "665f00fb2d149942b1b2b4b6", "mass": 2187000000000000000000000000000, "name": "Alpha Centauri A", "type": "Star", "diameter": 1214000, "constellation": "Centaurus", "distance_from_earth": 4.37}
 {"$s": {"p": {"_id": {"t": "objectId"}, "mass": {"t": "double"}, "name": {"t": "string"}, "type": {"t": "string"}, "diameter": {"t": "int"}, "constellation": {"t": "string"}, "distance_from_earth": {"t": "double"}}, "$k": ["_id", "name", "type", "distance_from_earth", "mass", "diameter", "constellation"]}, "_id": "665f00fb2d149942b1b2b4b7", "mass": 1804000000000000000000000000000, "name": "Alpha Centauri B", "type": "Star", "diameter": 865000, "constellation": "Centaurus", "distance_from_earth": 4.37}
 {"$s": {"p": {"_id": {"t": "objectId"}, "mass": {"t": "double"}, "name": {"t": "string"}, "type": {"t": "string"}, "diameter": {"t": "int"}, "constellation": {"t": "string"}, "distance_from_earth": {"t": "double"}}, "$k": ["_id", "name", "type", "distance_from_earth", "mass", "diameter", "constellation"]}, "_id": "665f00fb2d149942b1b2b4b8", "mass": 244600000000000000000000000000, "name": "Proxima Centauri", "type": "Star", "diameter": 200000, "constellation": "Centaurus", "distance_from_earth": 4.24}
 {"$s": {"p": {"_id": {"t": "objectId"}, "mass": {"t": "double"}, "name": {"t": "string"}, "type": {"t": "string"}, "diameter": {"t": "int"}, "constellation": {"t": "string"}, "distance_from_earth": {"t": "double"}}, "$k": ["_id", "name", "type", "distance_from_earth", "mass", "diameter", "constellation"]}, "_id": "665f027e2d149942b1b2b4b9", "mass": 27800000000000000000000000000000, "name": "Betelgeuse", "type": "Star", "diameter": 1200000000, "constellation": "Orion", "distance_from_earth": 642.5}
 {"$s": {"p": {"_id": {"t": "objectId"}, "mass": {"t": "double"}, "name": {"t": "string"}, "type": {"t": "string"}, "diameter": {"t": "int"}, "constellation": {"t": "string"}, "distance_from_earth": {"t": "double"}}, "$k": ["_id", "name", "type", "distance_from_earth", "mass", "diameter", "constellation"]}, "_id": "665f027e2d149942b1b2b4ba", "mass": 4074000000000000000000000000000, "name": "Vega", "type": "Star", "diameter": 2440000, "constellation": "Lyra", "distance_from_earth": 25.04}
(5 rows)
(END)
```

## Conclusion

It's hardly surprising that [PostgreSQL is one of the most widely adopted open source solutions](https://www.openlogic.com/resources/state-of-open-source-report), and with Aiven for PostgreSQL as your backend, you can start running MongoDB operations on FerretDB.

To start migrating from MongoDB to the truly open-source document database alternative, [check out this FerretDB migration guide](https://docs.ferretdb.io/migration/migrating-from-mongodb/).
