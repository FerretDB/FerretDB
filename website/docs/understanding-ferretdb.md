---
sidebar_position: 2
---

# Understanding FerretDB

FerretDB is a document database that uses similar commands, drivers, and tools to those of MongoDB for storing data.
Unlike relational databases, which use tables, rows, and columns to define the schema of the database, document databases use key-value pairs to build the structure of a document.

Document databases can collect, store, and retrieve data in a variety of data types.
In FerretDB, data is stored as [BSON](https://bsonspec.org/spec.html) - a binary representation of JSON - so you can store more types of data than in regular JSON.
However, we do not currently support 128-bit decimal floating point values.

Before inserting data into a document, you do not need to declare a schema.
That makes it ideal for applications and workloads requiring flexible schemas, such as blogs, chat apps, and video games.

:::note
For Tigris, FerretDB requires you to declare a JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).
Get more information on the key differences [here](diff.md).
:::

## Documents

Documents are self-describing records containing both data types and a description of the data being stored.
They are similar to rows in relational databases.
Here is an example of a single document:

```js
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
However, there are a few restrictions, which you can find [here](diff.md).
:::

For complex documents, you can nest objects (subdocuments) inside a document.

```js
{
  name: {
    first: "Thomas",
    last: "Edison"
  },
  invention: "Lightbulb",
  birth: 1847
}
```

In the example above, the `name` field is a subdocument embedded into a document.

## Dot notation

Dot notations `(.)` are used to reference a field in an embedded document or its index position in an array.

### Arrays

Dot notations can be used to specify or query an array by concatenating a dot `(.)` with the index position of the field.

```js
"array_name.index"
```

:::note
When using dot notations, the field name of the array and the specified value must be enclosed in quotation marks.
:::

For example, let's take the following array field in a document:

```js
animals: ["dog", "cat", "fish", "fox"]
```

To reference the fourth field in the array, use the dot notation `"animals.3"`.

Here are more examples of dot notations on arrays:

* [Query an array](basic-operations/read.md#retrieve-documents-containing-a-specific-value-in-an-array)
* [Update an array](basic-operations/update.md#update-an-array-element)

### Embedded documents

To reference or query a field in an embedded document, concatenate the name of the embedded document and the field name using the dot notation.

```js
"embedded_document_name.field"
```

Take the following document, for example:

```js
{
   name:{
      first: "Tom",
      last: "Barry"
   },
   contact:{
      address:{
         city: "Kent",
         state: "Ohio"
      },
      phone: "432-124-1234"
   }
}
```

To reference the `city` field in the embedded document, use the dot notation `"contact.address.city"`.

For dot notation examples on embedded documents, see here:

* [Query an embedded document](basic-operations/read.md#query-on-an-embedded-or-nested-document)
* [Update an embedded document](basic-operations/update.md#update-an-embedded-document)

## Collections

Collections are a repository for documents.
To some extent, they are similar to tables in a relational database.
If a collection does not exist, FerretDB creates a new one when you insert documents for the first time.
A collection may contain one or more documents.
For example, the following collection contains three documents.

```js
{
  Scientists: [
    {
      first: "Alan",
      last: "Turing",
      born: 1912
    },
    {
      first: "Thomas",
      last: "Edison",
      birth: 1847
    },
    {
      first: "Nikola",
      last: "Tesla",
      birth: 1856
    }
  ]
}
```

## Supported backends

:::caution
FerretDB is under constant development.
As with any database, before moving to production, please verify if it is suitable for your application.
:::

### PostgreSQL

PostgreSQL backend is our main backend and is fully supported.
MongoDB collections are mapped to PostgreSQL tables,
and MongoDB documents are mapped to rows with a single [JSONB](https://www.postgresql.org/docs/current/datatype-json.html) column.
The physical layout might change as we work on improving compatibility and performance.

### Tigris (beta)

We also support the [Tigris](https://www.tigrisdata.com) backend on a beta level.
Read more [here](https://www.tigrisdata.com/docs/concepts/mongodb-compatibility/).

### SAP HANA (alpha)

Currently, [we are also working](https://blogs.sap.com/2022/12/13/introduction-to-sap-hana-compatibility-layer-for-mongodb-wire-protocol/)
with SAP on HANA compatibility.
It is not officially supported yet.
