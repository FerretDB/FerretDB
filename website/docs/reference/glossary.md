---
sidebar_position: 2
---

# Glossary

## List of FerretDB terminologies

*This section contains a list of common terminologies related to FerretDB*.

### A

#### aggregation

A way of processsing documents in a collection and passing them through various operations or stages.
See [list of supported aggregation operations and commands here](supported-commands.md#aggregation-pipelines).

#### aggregation pipeline

A set of operators that lets you perform complex operations that aggregate and summarize values.
See [list of supported aggregation pipeline operators](supported-commands.md#aggregation-pipeline-operators) here.

---

### B

#### Beacon

The telemetry service of FerretDB.
See [telemetry](../telemetry.md) for more details.

#### BSON

BSON is a serialized binary file format for storing JSON-like documents.

#### BSON types

The list of types that the BSON format supports.
BSON offers support for additional data types compared to JSON, such as `timestamp`, `date`, `ObjectId`, and `binary`.

---

### C

#### collection

A group of documents in a non-relational database.
It is comparable to a table in a relational database.

#### CRUD

The four basic operations of a database: Create, Read, Update, and Delete.
See [Basic FerretDB CRUD operations here](../basic-operations/index.md).

---

### D

#### database

An organized repository for collections containing its own sets of documents, and data.

#### database command

The set of commands in FerretDB.
For more information, see [supported commands](supported-commands.md) for more details.

#### document

A record in a collection that comprises key-value pairs.
See [Documents](../understanding-ferretdb.md#documents) for more.

#### dot notation

Dot notation is used to reference or access the elements in an array or in an embedded document.
See [dot notation](../understanding-ferretdb.md#dot-notation) for more details.

---

### F

#### field

Similar to columns in a relational database.
They are represented as field name-value pairs and describe the kind of data in a document.

---

### I

#### index

A data structure used for identifying and querying records in a collection.
It helps to limit the number of documents to search through or inspect in a collection.
Examples include `_id` index, user-defined index, hashed index, and partial index.

---

### J

#### JSON

An acronym for JavaScript Object Notation.
It is a structured data format with human-readable text to store data objects composed of attribute-value pairs.

#### JSONB

JSONB is a data type of PostgreSQL that stores JSON data as a decomposed binary format.

#### ObjectId

A defining 12-byte type that ensures singularity and uniques within a collection and are used to represent the default values for the `_id` fields.

#### operator

A keyword that starts with a `$` character to query, update, or transform data.

---

### P

#### primary key

An immutable identifier for a record.
The primary key of a documents is stored in the `_id` field, which typically contains the `ObjectId`.

#### proxy

Proxy is any MongoDB-compatible database that is running in parallel with FerretDB.
It's used to test differences between FerretDB and other databases.
See [Operation modes](../configuration/operation-modes.md) for more details.

#### PostgreSQL

An open source relational database.
FerretDB uses PostgreSQL as a database engine.

---

### T

#### Tigris

A database platform used by FerretDB as a database engine.

---
