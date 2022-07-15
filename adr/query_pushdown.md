The goal of this task is to enable simple query pushdown for querier that look like `{_id: <ObjectID>}`.
Only that field name (`_id`) is supported, and only that value type (`ObjectID`).

In the future, we will add support for other fields (starting with simple scalar fields),
other values (starting with strings, numbers, and other simple scalar values),
other conditions like `{_id: <ObjectID>, field: value}` (so only the first part is pushed down).

## Postgres

If field name is different, value type is not `ObjectID`, or some other condition is present,
we should use the previous version of the code that SELECTs the whole table without any WHERE condition.

If value type is `ObjectID` but the data in the value is somehow corrupted, raise error (`fjson` unmarshal).

If those conditions are met, we send a SELECT query with a WHERE condition.

Proof of concept for a `{_id: <ObjectID>}` pushdown query, PostgreSQL:

GIN index is not used by PostgreSQL in any queries below.
So let's use the first one just because it is first.

```sql
-- this will not use these indexes:
-- CREATE INDEX values_id_idx ON public."values" USING gin ((((_jsonb -> '_id'::text) -> '$o'::text)))
-- CREATE INDEX values_id_idx ON values (((_jsonb ->> '_id')::text));
select * from values where (_jsonb->'_id')::jsonb->>'$o' = '507f1f77bcf86cd799439011'; -- no, seq scan will it use index? not that indexes
```

```sql
-- this will not use these indexes:
-- CREATE INDEX values_id_idx ON public."values" USING gin ((((_jsonb -> '_id'::text) -> '$o'::text)))
-- CREATE INDEX values_id_idx ON values (((_jsonb ->> '_id')::text));
select * from values where _jsonb->'_id' = '{"$o":"507f1f77bcf86cd799439011"}'::jsonb; --  will that one? // no, seq scan: conversion from jsonb to text
```


```sql
-- this will not use these index, but the results are faster
--- after planned has done it's job https://github.com/FerretDB/FerretDB/pull/847#issuecomment-1182871445:
-- CREATE INDEX values_id_idx ON public."values" USING gin ((((_jsonb -> '_id'::text) -> '$o'::text)))
select * from values where ((_jsonb->'_id'::text)->'$o')::text = '507f1f77bcf86cd799439011';
```

## Test

Integration tests: Provide a new test flag that enables a query pushdown.
Run with and without flag: results must be the same.
When explain feature will be ready, compare the query plan.

## Tigris

Support tables where the primary key is only one field.

The tjson package ensures that primary key is always `["_id"]` and always 1 length.
So no need to check:
* it's not necessary to get the schema description to check Primary Key fields and it's count.
* if `len(schema.PrimaryKey) > 1` raise error.
  * because user expects single one value in `_id`.
  * to comply and be compatible with MongoDB.
* if `len(schema.PrimaryKey) == 0` raise error.

But need to check if valaue type if not `ObjectID` (raise error in that case).

Proof of concept for a `{_id: <ObjectID>}` pushdown query, Tigris:


```go
var f filter.Expr
id := types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x01}
f = filter.Eq("_id", tjson.ObjectID(id))
id.Build()
it, err := db.Read(ctx, "coll1", f, nil)
```

# Other types

Here I described the observed complexity for future pushdown.

### Strings

String comparison difference between MongoDB and PostgreSQL i.e.
* case sensetive
* zero values
* etc

### Numeric

While looking for a numerical values in Postgres it might be different types as
* `int32`
* `int64`
* `double`
* etc

And search should look on all.

### `NaN`, `+-Inf`

Probably, re remove the support of `NaN`s.

### Arrays

```sh
test> db.values.find({value: 42})
[
  { _id: 'double-whole', value: 42 },
  { _id: 'int32', value: 42 },
  { _id: 'int64', value: Long("42") },
  { _id: 'array', value: [ 42 ] },
  { _id: 'array-three', value: [ 42, 'foo', null ] },
  { _id: 'array-three-reverse', value: [ null, 'foo', 42 ] }
]
```
## Query examples

```sql
CREATE TABLE values (
	_jsonb jsonb NOT NULL DEFAULT '{}'
);

CREATE INDEX values_id_idx ON values (((_jsonb ->> '_id')::text));

insert into values values('{"_id": 1.23}');
insert into values values('{"_id": 1}');
insert into values values(`{"_id": "s"}`);
insert into values values('{"_id": {"foo": "bar"} }');
insert into values values('{"_id": {"$f":"NaN"} }');
insert into values values('{"_id": {"$o": "507f1f77bcf86cd799439011"} }');
insert into values values('{"_id": {"$f":"-Infinity"} }');
insert into values values('{"_id": null }');
insert into values values('{"_id": [1] }');
insert into values values('{"_id": [null] }');
insert into values values('{"_id": { "_id": [null] } }');
insert into values values('{"_id": { "_id": { "$f":"NaN"} } }');
```

## Testing (overall)

Checks for `_id`:
* too long `_id` variable
* empty `_id`
* `_id` containing binary data
* attempt for SQL injection
* NaN
* nil
* arrays
* embedded document
* array of empty arrays
* array with nil value as an element
* arrays of arrays
* +/-Inf

## Documentation
Document behaviour in `README.md`

## Useful links

[PostgreSQL functions](https://www.postgresql.org/docs/14/functions-json.html)
[MongoDB restrictions on `_id`](https://www.mongodb.com/docs/manual/reference/limits/).
[MongoDB Comparison/Sort Order](https://www.mongodb.com/docs/manual/reference/bson-type-comparison-order/).

## Questions good to ask

How will we ensure that Tigris/Postgres compares values the same way as MongoDB?
