---
sidebar_position: 2
---

# Understanding FerretDB

FerretDB is an open-source proxy that translates MongoDB wire protocol queries to SQL,
with PostgreSQL or SQLite as the database engine.
It uses the same commands, drivers, and tools as MongoDB.

```mermaid
flowchart LR
  A["Any application\nAny MongoDB driver"]
  F{{FerretDB}}
  P[(PostgreSQL)]
  S[("SQLite")]

  A -- "MongoDB protocol\nBSON" --> F
  F -- "PostgreSQL protocol\nSQL" --> P
  F -. "SQLite library\nSQL" .-> S
```

:::tip
New to FerretDB?

Check out our:

- [Installation guide](./quickstart-guide/docker.md)
- [Key differences](./diff.md)
- [Basic CRUD operations](./basic-operations/index.md)

:::

## Supported backends

:::caution
FerretDB is under constant development.
As with any database, before moving to production, please verify if it is suitable for your application.
:::

### PostgreSQL

PostgreSQL backend is our main backend and is fully supported.

PostgreSQL should be configured with `UTF8` encoding and one of the following locales:
`POSIX`, `C`, `C.UTF8`, `en_US.UTF8`.

MongoDB databases are mapped to PostgreSQL schemas in a single PostgreSQL database that should be created in advance.
MongoDB collections are mapped to PostgreSQL tables.
MongoDB documents are mapped to rows with a single [JSONB](https://www.postgresql.org/docs/current/datatype-json.html) column.
Those mappings will change as we work on improving compatibility and performance,
but no breaking changes will be introduced without a major version bump.

### SQLite

We also support the [SQLite](https://www.sqlite.org/) backend.

MongoDB databases are mapped to SQLite database files.
MongoDB collections are mapped to SQLite tables.
MongoDB documents are mapped to rows with a single [JSON1](https://www.sqlite.org/json1.html) column.
Those mappings will change as we work on improving compatibility and performance,
but no breaking changes will be introduced without a major version bump.

### SAP HANA (alpha)

Currently, [we are also working](https://blogs.sap.com/2022/12/13/introduction-to-sap-hana-compatibility-layer-for-mongodb-wire-protocol/)
with SAP on HANA compatibility.
It is not officially supported yet.

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
'array_name.index'
```

:::note
When using dot notations, the field name of the array and the specified value must be enclosed in quotation marks.
:::

For example, let's take the following array field in a document:

```js
animals: ['dog', 'cat', 'fish', 'fox']
```

To reference the fourth field in the array, use the dot notation `"animals.3"`.

Here are more examples of dot notations on arrays:

- [Query an array](basic-operations/read.md#retrieve-documents-containing-a-specific-value-in-an-array)
- [Update an array](basic-operations/update.md#update-an-array-element)

### Embedded documents

To reference or query a field in an embedded document, concatenate the name of the embedded document and the field name using the dot notation.

```js
'embedded_document_name.field'
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

- [Query an embedded document](basic-operations/read.md#query-on-an-embedded-or-nested-document)
- [Update an embedded document](basic-operations/update.md#update-an-embedded-document)

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
      first: 'Alan',
      last: 'Turing',
      born: 1912
    },
    {
      first: 'Thomas',
      last: 'Edison',
      birth: 1847
    },
    {
      first: 'Nikola',
      last: 'Tesla',
      birth: 1856
    }
  ]
}
```
