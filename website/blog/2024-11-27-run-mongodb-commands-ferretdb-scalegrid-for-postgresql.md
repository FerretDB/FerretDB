---
slug: run-mongodb-commands-ferretdb-scalegrid-for-postgresql
title: 'Run MongoDB Commands on FerretDB with ScaleGrid for PostgreSQL'
authors: [alex]
description: >
  Learn how to run MongoDB workloads in FerretDB with a fully managed PostgreSQL service like ScaleGrid for PostgreSQL.
image: /img/blog/ferretdb-scalegrid.jpg
tags: [compatible applications, tutorial, cloud, postgresql tools, open source]
---

![Run MongoDB commands on FerretDB with ScaleGrid for PostgreSQL](/img/blog/ferretdb-scalegrid.jpg)

Many MongoDB users often express concerns about licensing restrictions, its proprietary nature, and the risk of vendor lock-in.
FerretDB, a truly open source MongoDB alternative, lets you run your workloads in PostgreSQL, using familiar syntaxes and commands.

<!--truncate-->

By translating MongoDB queries into SQL, FerretDB allows developers to leverage MongoDB-like functionality while using a relational database backend, such as [PostgreSQL](https://www.postgresql.org/).
When paired with [ScaleGrid for PostgreSQL](https://scalegrid.io/postgresql/), a fully managed database hosting service, you get a reliable and scalable solution for running MongoDB workloads without compromising performance or flexibility.
With ScaleGrid, developers can focus on building their applications while enjoying features like automated backups, monitoring, and high availability.
In this blog post, we'll explore how to use FerretDB and ScaleGrid for PostgreSQL to run MongoDB workloads.

## Prerequisites

Ensure to have the following installed before you start:

- [ScaleGrid account](https://scalegrid.io/)
- [`mongosh`](https://www.mongodb.com/docs/mongodb-shell/)
- [Docker](https://www.docker.com/)
- [psql](https://www.postgresql.org/docs/current/app-psql.html)

## Create a PostgreSQL deployment in ScaleGrid

FerretDB requires a PostgreSQL instance configured as the storage engine.
You can then connect to the database by passing the connection string to the `FERRETDB_POSTGRESQL_URL` environment variable or `--postgresql-url` flag.

Start by creating a fully managed Postgres deployment in ScaleGrid on any cloud platform of your choice.
[Follow this documentation to create a PostgreSQL deployment on ScaleGrid](https://help.scalegrid.io/docs/postgresql-new-cluster).

You'll need the connection string for the PostgreSQL instance once it's ready – it may take a few minutes to be provisioned.

## Run and connect FerretDB to ScaleGrid for PostgreSQL via Docker

From your local terminal, start by running the FerretDB container via Docker.
You'll need the connection string for your PostgreSQL instance for this.

```sh
docker run -e FERRETDB_POSTGRESQL_URL='postgresql://<username>:<password>@<host>/<database>' -p 27017:27017 ghcr.io/ferretdb/ferretdb
```

Ensure to replace `username`, `password`, `host`, and `port` with your ScaleGrid for PostgreSQL connection details.

With the FerretDB instance now running, connect to it via `mongosh` using the following connection string (replace the `username` and `password` with the user credentials for your PostgreSQL instance on ScaleGrid):

```sh
mongosh "mongodb://<username>:<password>@localhost/ferretdb?authMechanism=PLAIN"
```

## Run CRUD commands on FerretDB

Let's start by inserting some documents into the database.
The following command will insert two documents into a "cities" collection:

```js
db.cities.insertMany([
  {
    name: 'Kyoto',
    country: 'Japan',
    population: 1475000,
    landmarks: ['Fushimi Inari-taisha', 'Kinkaku-ji'],
    average_temperature: {
      winter: 5,
      summer: 28
    }
  },
  {
    name: 'Barcelona',
    country: 'Spain',
    population: 5500000,
    landmarks: ['Sagrada Familia', 'Park Güell'],
    average_temperature: {
      winter: 10,
      summer: 30
    }
  }
])
```

With the documents inserted successfully, let's find the city with a population greater than 2 million.

```js
db.cities.find({ population: { $gt: 2000000 } })
```

That should retrieve only "Barcelona" as a city with more than the specified population.
The output will look like this:

```json5
[
  {
    _id: ObjectId('67460732d1f590718c455c6f'),
    name: 'Barcelona',
    country: 'Spain',
    population: 5500000,
    landmarks: [ 'Sagrada Familia', 'Park Güell' ],
    average_temperature: { winter: 10, summer: 30 }
  }
]
```

Next, you can update the collection to include an additional landmark for "Kyoto".
Here, you'll search for cities named "Kyoto" and push an element into the landmark array.

```js
db.cities.updateOne(
  { name: 'Kyoto' },
  { $push: { landmarks: 'Arashiyama Bamboo Grove' } }
)
```

Now say there is a population increase of 200,000 in Barcelona, you want to update that as well.

````js
db.cities.updateOne(
    { "name": "Barcelona" },
    { $inc: { "population": 200000 } }
);
```

Run `db.cities.find()` to see the newly updated collection  –  the population of "Barcelona" should have increased to "5700000" and "Kyoto" should now have three elements in its "landmarks" array.

```json5
[
  {
    _id: ObjectId('67460732d1f590718c455c6e'),
    name: 'Kyoto',
    country: 'Japan',
    population: 1475000,
    landmarks: [ 'Fushimi Inari-taisha', 'Kinkaku-ji', 'Arashiyama Bamboo Grove' ],
    average_temperature: { winter: 5, summer: 28 }
  },
  {
    _id: ObjectId('67460732d1f590718c455c6f'),
    name: 'Barcelona',
    country: 'Spain',
    population: 5700000,
    landmarks: [ 'Sagrada Familia', 'Park Güell' ],
    average_temperature: { winter: 10, summer: 30 }
  }
]
```

Finally, let's delete a city with an average winter temperature less than or equal to 5 ℃.

```js
db.cities.deleteMany({ "average_temperature.winter": { $lte: 5 } });
```

When you run `db.cities.find()`, it should leave you with a single document  –  "Barcelona".

```json5
[
  {
    _id: ObjectId('67460732d1f590718c455c6f'),
    name: 'Barcelona',
    country: 'Spain',
    population: 5700000,
    landmarks: [ 'Sagrada Familia', 'Park Güell' ],
    average_temperature: { winter: 10, summer: 30 }
  }
]
```

If you're interested in seeing how the database looks like in Postgres, ScaleGrid for PostgreSQL provides a `psql` command with the connection string. Or you can use the `FERRETDB_POSTGRESQL_URL` from earlier.

```sh
psql 'postgresql://<username>:<password>@<host>/<database>'
```

Once you're in, set the search path to `postgres` and display the data:


```text
postgres=# set search_path to postgres;
SET
postgres=# \dt
                      List of relations
  Schema  |            Name             | Type  |   Owner
----------+-----------------------------+-------+------------
 postgres | _ferretdb_database_metadata | table | sgpostgres
 postgres | cities_fb4544d2             | table | sgpostgres
(2 rows)
postgres=# SELECT * from cities_fb4544d2;
                                                                                                                                                                                                                                                                                                             _jsonb
---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "country": {"t": "string"}, "landmarks": {"i": [{"t": "string"}, {"t": "string"}], "t": "array"}, "population": {"t": "int"}, "average_temperature": {"t": "object", "$s": {"p": {"summer": {"t": "int"}, "winter": {"t": "int"}}, "$k": ["winter", "summer"]}}}, "$k": ["_id", "name", "country", "population", "landmarks", "average_temperature"]}, "_id": "67460732d1f590718c455c6f", "name": "Barcelona", "country": "Spain", "landmarks": ["Sagrada Familia", "Park Güell"], "population": 5700000, "average_temperature": {"summer": 30, "winter": 10}}
(1 row)
```

## Conclusion

Your entire document and collection data, all stored in Postgres  –  that's what FerretDB lets you do. And with that, you can take advantage of the simplified and scalable setup of ScaleGrid for PostgreSQL to manage your data.

To get started with your migration from MongoDB to FerretDB, here's some materials to make the process easier.

[Migrate your MongoDB workloads to FerretDB](https://docs.ferretdb.io/migration/migrating-from-mongodb/)
[Quickstart guide for FerretDB](https://docs.ferretdb.io/quickstart-guide/docker/)
````
