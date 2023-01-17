---
slug: pjson-how-to-store-bson-in-jsonb
title: "How FerretDB stores BSON in JSONB"
author: Chi Fujii
image: ../static/img/blog/six_ferrets-1024x917.jpg
date: 2022-11-08
---

![How FerretDB stores BSON in JSONB](../static/img/blog/christian-wiediger-WkfDrhxDMC8-unsplash-1024x683.jpg)

<!--truncate-->

At FerretDB, we are converting MongoDB wire protocol queries into SQL, to store information from BSON to PostgreSQL.
To achieve this, we created our own mapping called PJSON which translates MongoDB’s BSON format into PostgreSQL’s JSONB.

Using BSON with JSONB offers us significantly faster operation, greater range of data types compared to regular JSON, and more flexibility in modeling data of any structure.

In this article, we’ll show you how FerretDB converts and stores MongoDB data in PostgreSQL.

## What is BSON?

[BSON](https://bsonspec.org/) holds a collection of field name/value pairs which is called a document.
It contains length information allowing serializer/deserializer to utilize it for the performance benefit.
Additionally, it preserves the order of the fields.
BSON also supports additional data types such as DateTime and binary.

Let’s look at an example of how a JSON is encoded in BSON.
Here the hexadecimal `\x00` notation represents `0000 0000` in bits and similarly `\x12` represents `0001 0010`.
In BSON, `\x00` is a byte used as a terminator to indicate the end of a document.
In BSON, field names use `cstring` which are UTF-8 characters followed by `\x00`.

{“foo”: “bar”}

|                         |                                                                                                                       |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------- |
| \x12\x00\x00\x00        | Document length in little-endian int32 (18 bytes document)                                                            |
| \x02                    | string field type `\x02`, see the [BSON spec](https://bsonspec.org/spec.html "")                                      |
| foo\x00                 | Cstring field name                                                                                                    |
| \x04\x00\x00\x00bar\x00 | String length in little-endian int32 `\x04\x00\x00\x00` (4 bytes string) followed by string value and trailing `\x00` |
| \x00                    | Document terminator                                                                                                   |

You notice that BSON is binary serialized so it is not human readable.
It contains information about the length, and an explicitly defined field type.

At FerretDB, we use PostgreSQL as a database engine and we store JSON data in [JSONB](https://www.postgresql.org/docs/15/datatype-json.html) data type.
In light of this, we need to store BSON equivalent information in JSON format without losing type information.

## Introducing PJSON

To embed necessary information from BSON, we introduced special keys prefixed with `$`.
PJSON contains `$` followed by a character to embed information about types and the order of the fields.
PJSON is designed to be serialized to JSONB.
Let’s look at a simple BSON with an `ObjectId` field and see how it is represented in PJSON.

```js
{"_id”: ObjectId("635202c8f75e487c16adc141")}
```

```js
\x16\x00\x00\x00\x07_id\x00\x63\x52\x02\xc8\xf7\x5e\x48\x7c\x16\xad\xc1\x41\x00
```

In PJSON, we store the order of fields in the `$k` field, and `ObjectId` in the `$o` field.
The duplicate fields are not allowed in PJSON.

```js
{
  "$k": [
    "_id"
  ],
  "_id": {
    "$o": "635202c8f75e487c16adc141"
  }
}
```

The `$` prefixes apply to types that are not native to JSON.
The current mappings to PJSON are defined below.

|                              |                                                                                                                  |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Document                     | `{“$k”: [“<key 1>”, “<key 2>”, …], “<key 1>”: <value 1>, “<key 2>”: <value 2>, …}`                               |
| Array                        | JSON array                                                                                                       |
| 64-bit binary floating point | `{“$f”: JSON number}`                                                                                            |
| UTF-8 string                 | JSON string                                                                                                      |
| Binary data                  | `{“$b”: “<base 64 string>”, “s”: <subtype number>}` // s is binary subtype, the detail is found in the BSON spec |
| ObjectId                     | `{“$o”: “<ObjectID as 24 character hex string”}`                                                                 |
| Boolean                      | JSON true / false values                                                                                         |
| UTC datetime                 | `{“$d”: milliseconds since epoch as JSON number}`                                                                |
| Null                         | JSON null                                                                                                        |
| Regular expression           | `{“$r”: “<string without terminating 0x0>”, “o”: “<string without terminating 0x0>”}`                            |
| Timestamp                    | `{“$t”: “<number as string>”}`                                                                                   |
| 64-bit integer               | `{“$l”: “<number as string>”}`                                                                                   |

In addition to `$` prefixes, binary data have an additional field `s`  to indicate the type of binary.
Also, regular expressions have an additional field `o` to specify options such as case sensitivity.

## PJSON Example: Translating and storing BSON in JSONB

Let’s look at an example of inserting BSON and storing it as PJSON using the following document.

```js
db.groceries.insert({
  _id: ObjectId("635202c8f75e487c16adc141"),
  name: “milk”,
  quantity: 3
})
```

### 1. Deserialize from BSON

In our BSON deserializer implementation, we use the byte tag according to [BSON spec](https://bsonspec.org/spec.html)to extract field name value pairs.

```js
\x33\x00\x00\x00 Document length
\x07 ObjectId field type
_id\x00
\x63\x52\x02\xc8\xf7\x5e\x48\x7c\x16\xad\xc1\x41
\x02 String field type
name\x00
\x05\x00\x00\x00milk\x00
\x10 int32 field type
quantity\x00
\x03\x00\x00\x00
\x00
quantity\x00
\x03\x00\x00\x00
\x00

```

A BSON document contains the document length in `int32` at the very first entry of the document.
BSON uses a little-endian format for `int32` type, the little-endian orders bytes from the least significant byte to the most significant byte.
The document length `\x33\x00\x00\x00` is in bytes so this document is 51 bytes in length in decimal representation.

The first field type `\x07` is `ObjectId`.
The field name `_id\x00` is obtained by reading until the first `\x00`.
The BSON field name `_id\x00` is `cstring` type which is a UTF-8 `string` followed by `\x00`.
We convert `cstring` to `string` by dropping the `\x00`, and we have the field name `_id`.
The `ObjectId` field value is specified in BSON spec as 12 bytes, we read subsequent 12 bytes as `ObjectId` which are `\x63\x52\x02\xc8\xf7\x5e\x48\x7c\x16\xad\xc1\x41`.

The second field type `\x02` is `string`, the field name `name\x00`  is obtained by reading the `cstring` field name until the first `\x00`.
The string field value `\x05\x00\x00\x00milk\x00` contains the string length followed by the value.
The first part `\x05\x00\x00\x00` is string length in little-endian int32 format which says the string is 5 bytes length.
We read the subsequent 5 bytes `milk\x00` and we drop the `\x00` to get the field value of `milk`.

The last field `\x10` is `int32` type, encoded in little-endian with 8 bytes length.
The field name `quantity\x00` is obtained by reading until the first `\x00`, and the field value of `int32` is obtained by reading the subsequent 8 bytes `\x03\x00\x00\x00`.
The field value is little-endian format `int32` which is 3.

The terminator `\x00` indicates it’s the end of the BSON document.

### 2. Serialize to PJSON

We use the array `$k` to preserve the order of the fields.
In PJSON, the format and types, such as `string` and `int32`, are the same as JSON types.
On the other hand, `ObjectId` is represented by the `$o` key which is specific to BSON.

```js
{
  "$k": [
    "_id",
    "name",
    "quantity"
  ],
  "_id": {
    "$o": "635202c8f75e487c16adc141"
  },
  "name": "milk",
  "quantity": 3
}
```

### 3. Store PJSON to JSONB

We use a table `_ferretdb_settings` to keep track of existing collections.
For each collection, we create a table with a JSONB column and it stores the documents of the collection.

When we insert a document to the `groceries` collection, an entry related to a new collection `groceries` was added to `_ferretdb_settings`.
This happened because it was the first document inserted into the `groceries` collection.
And a new document was inserted to a new generated table `groceries_6a5f9564`.

```js
ferretdb=# \d test._ferretdb_settings
          Table "test._ferretdb_settings"
  Column  | Type  | Collation | Nullable | Default
----------+-------+-----------+----------+---------
 settings | jsonb |           |          |
ferretdb=# SELECT settings FROM test._ferretdb_settings;
                                             settings
--------------------------------------------------------------------------------------------------
 {"$k": ["collections"], "collections": {"$k": ["groceries"], "groceries": "groceries_6a5f9564"}}
(1 row)
```

```js
ferretdb=# \d test.groceries_6a5f9564
         Table "test.groceries_6a5f9564"
 Column | Type  | Collation | Nullable | Default
--------+-------+-----------+----------+---------
 _jsonb | jsonb |           |          |
ferretdb=# SELECT _jsonb FROM test.groceries_6a5f9564;
                                                    _jsonb
---------------------------------------------------------------------------------------------------------------
 {"$k": ["_id", "name", "quantity"], "_id": {"$o": "635202c8f75e487c16adc141"}, "name": "milk", "quantity": 3}
(1 row)
```

This is how we store BSON information at FerretDB in PostgreSQL JSONB.

## Roundup

So far, we’ve shown how we use PJSON mappings to store MongoDB BSON data in JSONB while preserving type information and field orders.
This process includes deserializing BSON and mapping it to PJSON, and finally storing them in JSONB.

For our community contributors and users, understanding how we convert BSON data in MongoDB to JSONB of PostgreSQL helps them gain more insight into how FerretDB works.

To start contributing to FerretDB, read our [contribution guidelines](https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md).
