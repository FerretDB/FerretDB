---
title: "How FerretDB fetches data - about query pushdown"
slug: ferretdb-fetches-data-query-pushdown
author: Patryk Kwiatek
description: "Find out how FerretDB uses query pushdown to fetch data from the storage layer (also called as “backend”)"
image: /img/blog/ferretdb-query-pushdown.jpg
unlisted: true
---

![How FerretDB fetches data - about query pushdown](/img/blog/ferretdb-query-pushdown.jpg)
**Credit:** [Mohamed Hassan](https://pixabay.com/users/mohamed_hassan-5229782/)

Pushdown is the method of optimizing a query by reducing the amount of data read and processed.
It saves memory space, and network bandwidth, and reduces the query execution time by not prefetching unnecessary data to the database management system.

When you fetch less data, you spend less memory, generate less network traffic (which can be time-consuming), and overall you operate on a smaller subset of data, as you don’t need to iterate over *huge* piles of it.

This article will give you a brief technical overview of how FerretDB fetches the data from the storage layer (also called as “backend”), and how we use query pushdowns to optimize this process.

## Why does FerretDB need SQL query pushdowns?

As we aim to be as compatible with MongoDB drivers as possible, all operations, comparisons, data types, and commands need to be handled in the same fashion as MongoDB.
Because of that, we cannot rely on SQL queries and filter data just in queries.

For example how would we compare values of the different types (considering BSON types comparison order)?
The solution is to do all filtering operations on our own.
That creates a need for fetching all the data from a storage layer, which for large collections, can be really inefficient and time-consuming.

Considering this fact, query pushdowns are a really important method for decreasing the data that we must fetch for every query.
That's why FerretDB really can benefit from using them in such queries.
Fortunately, we’ve managed to introduce the query pushdown with [this PR](https://github.com/FerretDB/FerretDB/pull/1207)!

## How FerretDB fetches data using query pushdown

Let's jump right into the FerretDB internals and go step by step to see how it handles the sample query!

Let's say we have a collection with data on thousands of customers, and want to check if the one under “john.doe@example.com” email address has an active account:

```js
db.customers.find({ 'email': 'john.doe@example.com' },{ active: 1 })
```

FerretDB will extract the command (`find`), filters (`{ 'email': 'john.doe@example.com' }`, and projections `{ active: 1 }`.

To `find` the expected document, it needs to use the filter on every document inside `customers` collection, and as mentioned before, to operate on the collection’s data, FerretDB needs to fetch all of it from a storage layer (in our example - PostgreSQL).

Let’s trace the queries sent to the backend!

At the beginning, FerretDB checks if PostgreSQL's `test` schema contains the `_ferretdb_database_metadata` table:

```js
SELECT EXISTS ( SELECT 1 FROM information_schema.columns WHERE table_schema = 'test' AND table_name = '_ferretdb_database_metadata' );
```

The `_ferretdb_database_metadata` contains the mapping of all collection names to the actual PostgreSQL table names.
This method is required, as PostgreSQL has table name limitations (such as length limit of 63 characters) which could break MongoDB compatibility on some environments.

Afterwards FerretDB takes the table name that matches the `customers` collection with the following query:

```js
SELECT _jsonb FROM "test"."_ferretdb_database_metadata" WHERE ((_jsonb->'_id')::jsonb = '"customers"');

 _jsonb ----------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "string"}, "table": {"t": "string"}}, "$k": ["_id", "table"]}, "_id": "customers", "table": "customers_c09344de"}
```

As we know that `customers` collection is mapped to the `customers_c09344de` table, we are ready to fetch the documents from it:

```js
SELECT _jsonb FROM "test"."customers_c09344de";
```

```js
_jsonb -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "email": {"t": "string"}, "active": {"t": "bool"}}, "$k": ["_id", "email", "name", "active"]}, "_id": "63aa97626786637ef1c4b722", "name": "Alice", "email": "alice@example.com", "active": true}
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "email": {"t": "string"}, "active": {"t": "bool"}}, "$k": ["_id", "email", "name", "active"]}, "_id": "63aa97626786637ef1c4b723", "name": "Bob", "email": "bob@example.com", "active": true}
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "email": {"t": "string"}, "active": {"t": "bool"}, "surname": {"t": "string"}}, "$k": ["_id", "email", "name", "surname", "active"]}, "_id": "63aa97626786637ef1c4b724", "name": "Jane", "email": "jane@example.com", "active": true, "surname": "Smith"}
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "email": {"t": "string"}, "active": {"t": "bool"}, "surname": {"t": "string"}}, "$k": ["_id", "email", "name", "surname", "active"]}, "_id": "63aa97626786637ef1c4b725", "name": "John", "email": "john.doe@example.com", "active": false, "surname": "Doe"}
…
```

The previous command returned all documents from the `customers` collection.
They are stored in `jsonb` format.
You may have noticed some differences between the BSON document that's returned to the user and the json stored in the database.
In the database in addition to the document we also store the document's schema in the `$s` field.
This is needed for our internal representation of the documents called `pjson`, you can learn more about this topic in the [How FerretDB stores BSON in JSONB](https://www.ferretdb.io/pjson-how-to-store-bson-in-jsonb/) article.

* *Note*: As FerretDB is constantly evolving, please take note that the following article may contain a bit outdated information.
We changed the way of storing document types and moved them to the `$s` field that stores all field properties and keys in it.
All other information in the article should be still relevant.

With these documents, FerretDB can parse the data to internal `pjson` type, and iterate through all of them to apply the filters.
After this process, as only one document matches the filter, the projection will be applied to it, so only the `active` (and `_id`) fields will be returned:

```js
[ { _id: ObjectId("63aa97626786637ef1c4b725"), active: false } ]
```

Cool!
Now we know that the customer account is not active.
We can send him a small reminder about the account activation.

But let's go back to our query.
We only want to get information about the single person with a simple identifier.
To do that FerretDB just gets the whole collection from the PostgreSQL database.

This doesn't seem like a big problem on a small set of data, but in this example, we have thousands of customers, and let's suppose that it could grow to even *hundreds of thousands*!

Fetching all of this data creates an unnecessarily large amount of network traffic, and consumes too much memory and time.
It's unreasonable to fetch all of this data, just to apply this simple filter that only a single document in the collection will satisfy.

Let’s go back to our workflow and suppose that afterward the account was activated, so we want to ensure that.
If we know the exact `_id` of the customer’s document, we can use it to benefit from the query pushdown:

```js
db.customers.find({ '_id': ObjectId('63aa97626786637ef1c4b725') }, { active: 1 })
```

Let’s see what SQL queries FerretDB will send.
The beginning of the process is the same as previous - it checks if a collection exists and fetches the table name.
After that, to fetch the documents, it sends following query:

```js
SELECT _jsonb FROM "test"."customers_c09344de" WHERE ((_jsonb->'_id')::jsonb = '"63aa97626786637ef1c4b725"');
                                                                                                                                                        _jsonb
------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "name": {"t": "string"}, "email": {"t": "string"}, "active": {"t": "bool"}, "surname": {"t": "string"}}, "$k": ["_id", "email", "name", "surname", "active"]}, "_id": "63aa97626786637ef1c4b725", "name": "John", "email": "john.doe@example.com", "active": true, "surname": "Doe"}
```

As you can see the query that we sent now differs from the one sent on the past action.

Surely we still fetch from the customers collection’s table, but as we only want the single customer and FerretDB supports pushdowns for filters with `_id` key and the value of the `ObjectID` type - we don’t fetch all documents, but we use the WHERE clause to get only one of them.
As a result, we significantly reduced network traffic, FerretDB doesn’t iterate through thousands of records and most importantly the response time is much quicker!

## Measuring performance gain

Let’s check how pushdowns affected a simple operation in a benchmark!

For the sake of this article, we’ve recreated the environment from the [quickstart guide](https://docs.ferretdb.io/quickstart_guide/docker/).

Afterwards, we’ve restored a dump with 10000 documents from [this dataset](https://github.com/mcampo2/mongodb-sample-databases/tree/master/dump/sample_weatherdata.).

To measure differences between a pushdown query and the one without any pushdown, we’ve acknowledged that the pushdown for `_id` field is only done if the value is of the ObjectID or string type.
So, if we use any other field there, we are able to produce a non-optimized query.

Our benchmark will run 3 cases.
Two of them will cause a pushdown on `_id` field, and the third one will query the `v`  field, which at the moment of writing the article cannot be pushdowned by FerretDB.
You can find the code of the benchmark in the [FerretDB repository](https://github.com/FerretDB/FerretDB/blob/c50e8344f1ead5f25a34352eb76643c30baf4bf4/integration/benchmarks_test.go).

Now we can run the test:

```js
$ go test -bench=. -run=^#
goos: linux
goarch: amd64
pkg: benchmark
cpu: Intel(R) Core(TM) i5-8600K CPU @ 3.60GHz
BenchmarkPushdowns/ObjectID-6             10     101285307 ns/op
BenchmarkPushdowns/StringID-6             10     100922313 ns/op
BenchmarkPushdowns/NoPushdown-6            1     5216650215 ns/op
PASS
ok benchmark   7.454s
```

The results show that the tests with pushdown queries took around 101 ms, which in comparison to non-pushdown one (5217 ms), **reduced the execution time by a factor of ~52!**

## A sneak peek into the future

In the near future, [we will add a support for other fields](https://github.com/FerretDB/FerretDB/issues/4) (like simple scalar fields) and other values (starting with numbers, strings, and other simple scalar values).
As our pushdown related code is written to be easily extensible, most of the future pushdown implementations are just a matter of adding a couple of cases for each of them.

Longterm, we consider adding even more complicated pushdown cases to continue boosting the performance.
