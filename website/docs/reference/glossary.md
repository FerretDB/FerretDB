---
sidebar_position: 2
---

# Glossary

## List of FerretDB terminologies

*This section contains a list of common terminologies related to FerretDB*.

### A

#### aggregation

A group of operations used to reduce or summarize large data sets.
See [list of supported aggregation operations and commands here](./supported_commands.md#aggregation-pipelines).

#### aggregation pipeline

A set of operators that lets you perform complex operations that aggregate and summarize values.
See [list of supported aggregation pipeline operators](./supported_commands.md#aggregation-pipeline-operators) here.

---

### B

#### Beacon

The telemetry service of FerretDB.

#### BSON

Stands for “binary JSON”, which is a serialized binary file format for storing JSON-like documents.

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
See [Basic FerretDB CRUD operations here](../basic_operations/index.md).

---

### D

#### dance

The integration testing tool of FerretDB, which is named after the Ferret war dance.
See [dance repository](https://github.com/FerretDB/dance) for more details.

#### database

An organized repository for collections containing its own sets of documents, and data.

#### database command

The set of commands in FerretDB.
For more information, see [supported commands](./supported_commands.md) for more details.

#### document

A record in a collection that comprises key-value pairs.
See [Documents](../understanding_ferretdb.md#documents) for more.

#### dot notation

Dot notation is used to reference or access the elements in an array or in an embedded document.
See [dot notation](../understanding_ferretdb.md#dot-notation) for more details.

---

### F

#### field

Similar to columns in a relational database.
They are represented as field-value pairs and describe the kind of data in a document.

---

### G

#### github-actions

A repository that holds the shared GitHub Actions across all FerretDB repositories.
See [github-actions repository](https://github.com/FerretDB/github-actions) for more details.

---

### I

#### index

A data structure used for identifying and querying records in a database.

---

### J

#### JSON

An acronym for JavaScript Object Notation.
It is a structured data format with human-readable text  to store data objects composed of attribute-value pairs.

#### JSONB

JSONB is a data type that stores JSON data as a decomposed binary format.

#### ObjectId

A defining 12-byte type that ensures singularity and uniques within a collection and are used to represent the default values for the `_id` fields.
It contains a 4-byte timestamp value measured in seconds, uniquely random machine and process-generated ID, and an incremental counter.

#### operator

A keyword that starts with a `$` character to query, update, or transform data.

---

### P

#### PJSON

FerretDB’s mapping system which converts BSON into JSONB.
It uses embedded data in BSON containing special keys that starts with `$` along with information about the data types and field order to serialize data into JSONB.
See [this article for more details on PJSON](https://www.ferretdb.io/pjson-how-to-store-bson-in-jsonb/).

#### primary key

An immutable identifier for a record.
The primary key of a documents is stored in the `_id` field, which typically contains the `ObjectId`.

#### PostgreSQL

An open source relational database.
FerretDB uses PostgreSQL as a database engine.

---

### T

#### Tigris

Also known as Tigris Data.
A database platform used by FerretDB as a database engine.

---
