---
sidebar_position: 4
---

# Understanding FerretDB

FerretDB is a document database that uses similar commands, drivers, and tools to those of MongoDB for storing data.
Unlike relational databases, which use tables, rows, and columns to define the schema of the database, document databases use key-value pairs to build the structure of a document.

Document databases can collect, store, and retrieve any data type.
In FerretDB, data is stored as [BSON](https://bsonspec.org/spec.html) - a binary representation of JSON - so you can store more types of data than in regular JSON.
However, we do not currently support "128-bit decimal floating points".

Before inserting data into a document, you do not need to declare a schema.
That makes it ideal for applications and workloads requiring flexible schemas, such as blogs, chat apps, and video games.

## Documents

Documents are self-describing records containing both data types and a description of the data being stored.
They are similar to rows in relational databases.
Here is an example of a single document:

```json
{
    first: "Thomas",
    last: "Edison",
    invention: "Lightbulb",
    birth: 1847
}

```

The above data is stored in a single document.

:::note
FerretDB follows almost the same naming conventions as MongoDB.
However, there are a few restrictions, which you can find  [here](https://docs.ferretdb.io/diff/).
:::

For complex documents, you can nest objects (subdocuments) inside a document.

```json
{
    name: { first: "Thomas", last: "Edison" },
    invention: "Lightbulb",
    birth: 1847
}
```

In the example above, the `name` field is a subdocument embedded into a document.

## Collections

Collections are a repository for documents.
To some extent, they are similar to tables in a relational database.
If a collection does not exist, this operation creates a new one.
A collection may contain one or more documents.
For example, the following collection contains threee documents.

```json
 Scientists: {
    first : "Alan",
    last : "Turing",
    born : 1912
    },
    {
    first : "Thomas",
    last : "Edison",
    birth : 1847
    },
    {
    first : "Nikola",
    last : "Tesla",
    birth : 1856
    }

```

## Indexing

FerretDB does not currently support indexing.
Kindly check [our roadmap](https://github.com/orgs/FerretDB/projects/2) or the [open issue on indexing support](https://github.com/FerretDB/FerretDB/issues/78) to know when this will be made available available.

## Data Storage

 FerretDB converts MongoDB drivers and protocols to SQL, with the data stored in PostgreSQL or Tigris Data.
For PostgreSQL, we convert MongoDB bson to jsonb and store it as a value in a row.

In Tigris case, we convert bson to tigirs tjson.
There are a few things to take note of here.
For example, you cannot overwrite the data type of a value.
Please check here for more details on the differences.

:::caution
FerretDB is still under development and not currently suitable for production-ready environments.
:::
