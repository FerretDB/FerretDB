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

[PostgreSQL functions](https://www.postgresql.org/docs/14/functions-json.html)


## Test

Integration tests: Provide a new test flag that enables a query pushdown.
Run with and without flag: results must be the same.
When explain feature will be ready, compare the query plan.

## Tigris

Support tables where the primary key is only one field.

The tjson package ensures that primary key is always `["_id"]`.
So no need to check:
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