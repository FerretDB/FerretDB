Select queries of the form `{_id: <value>}`

## The problem

[The pushdown term](https://www.quora.com/What-do-we-mean-when-we-say-SQL-pushdown)

### Strings

string comparison difference between MongoDB and PostgreSQL (i.e. case sensetive etc)

### Numeric

While looking for a numerica values in Postgres it might be different types as
* `int32`
* `int64`
* `double`
* etc

And search should look on all.

### `NaN`, `+-Inf`

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

NB: There is no new functionality from the user perspective – we already support _id values that are documents, for example, and that should not change.

## Solution

### Build tag

Add a build tag:
If the build tag is enabled:
* then for queries `{_id: <value>}`, use a simple pushdown
* for all other queries, process as usual.

## SQL samples

```sql
CREATE TABLE test (
	_jsonb jsonb NOT NULL DEFAULT '{}'
);

insert into test values('{"_id": 1.23}');
insert into test values('{"_id": 1}');
insert into test values(`{"_id": "s"}`);
insert into test values('{"_id": {"foo": "bar"} }');
insert into test values('{"_id": {"$f":"NaN"} }');
insert into test values('{"_id": {"$f":"-Infinity"} }');
insert into test values('{"_id": null }');
insert into test values('{"_id": [1] }');
insert into test values('{"_id": [null] }');
insert into test values('{"_id": { "_id": [null] } }');
insert into test values('{"_id": { "_id": { "$f":"NaN"} } }');

-- output all
select * from test;
             _jsonb
---------------------------------
 {"_id": 1.23}
 {"_id": "s"}
 {"_id": [1]}
 {"_id": {"foo": "bar"}}
 {"_id": {"$f": "NaN"}}
 {"_id": {"$f": "-Infinity"}}
 {"_id": null}
 {"_id": [null]}
 {"_id": {"_id": [null]}}
 {"_id": {"_id": {"$f": "NaN"}}}
 {"_id": 1}
(11 rows)

-- see types
select _jsonb->'_id' v, jsonb_typeof(_jsonb->'_id') from test;
           v            | jsonb_typeof
------------------------+--------------
 1.23                   | number
 "s"                    | string
 [1]                    | array
 {"foo": "bar"}         | object
 {"$f": "NaN"}          | object
 {"$f": "-Infinity"}    | object
 null                   | null
 [null]                 | array
 {"_id": [null]}        | object
 {"_id": {"$f": "NaN"}} | object
 1                      | number
(11 rows)

-- example
select jsonb_typeof(_jsonb->'_id') from test where jsonb_typeof(_jsonb->'_id') = 'number';

-- number
select * from test where jsonb_typeof(_jsonb->'_id') = 'number' and (_jsonb->'_id')::numeric = 1.23;
    _jsonb
---------------
 {"_id": 1.23}
(1 row)

-- string
select * from test where jsonb_typeof(_jsonb->'_id') = 'string' and (_jsonb->'_id')::text = '"s"';
    _jsonb
--------------
 {"_id": "s"}

-- document
select * from test where jsonb_typeof(_jsonb->'_id') = 'object' and _jsonb->'_id' = '{"foo": "bar"}'::jsonb;
         _jsonb
-------------------------
 {"_id": {"foo": "bar"}}
(1 row)

-- NaN
select * from test where jsonb_typeof(_jsonb->'_id') = 'object' and _jsonb->'_id' = '{"$f":"NaN"}'::jsonb;
         _jsonb
------------------------
 {"_id": {"$f": "NaN"}}
(1 row)

-- Inf
select * from test where jsonb_typeof(_jsonb->'_id') = 'object' and _jsonb->'_id' = '{"$f":"-Infinity"}'::jsonb;
            _jsonb
------------------------------
 {"_id": {"$f": "-Infinity"}}
(1 row)

-- null
select * from test where jsonb_typeof(_jsonb->'_id') = 'null';
    _jsonb
---------------
 {"_id": null}
(1 row)

-- however
select * from test where jsonb_typeof(_jsonb->'_id') IS NULL;
 _jsonb
--------
(0 rows)


-- [null]
select * from test where jsonb_typeof(_jsonb->'_id') = 'array' and _jsonb->'_id' = '[null]'::jsonb;
     _jsonb
-----------------
 {"_id": [null]}
(1 row)


-- [1]
select * from test where jsonb_typeof(_jsonb->'_id') = 'array' and _jsonb->'_id' = '[1]'::jsonb;
    _jsonb
--------------
 {"_id": [1]}
(1 row)

```

### Tigris

Tigris API provides querying by `_id`. Let's use it.

## Testing

* Unit-tests for fetch function, that it returns not all the collection documents but a single document when `{_id: <value>}` is queried.
* Integrational tests for insert function after inserting all expected records in the database.

## Cases

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

TODO: add queries


## Documentation

Document build tag in `README.md`

