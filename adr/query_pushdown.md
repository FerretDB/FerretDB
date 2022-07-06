* `{_id: <value>}`

## The problem

[The pushdown term](https://www.quora.com/What-do-we-mean-when-we-say-SQL-pushdown)

Before we filtered tuples in Go code due to task complexity with some filtering operations.
For now we are ready to start supporting query pushdown.
Let's go step by step, and first, implement a simple query pushdown for queries containing `{_id: <value>}`,
i.e. add SQL condition `_jsonb->`+ `p.Next()` + `=` in WHERE clause passed to the PostgreSQL backend.

NB: There is no new functionality from user perspective – we already support _id values that are documents, for example, and that should not change.

## Solution

* Add a function to `pgPool` that queries `_id`  and returns a single record, nothing or error.
* In `pg` handler, modify `(h *Handler) fetch`, add the usage of that function.

### Build tag
To not change the behavior, let's add a build tag:
If the build tag is enabled:
* then for queries `{_id: <value>}`, use a simple pushdown
* for all other queries process as usual.

### Examples on insert

```sh
test_id> db.test.insertOne({"_id": 1.23})
{ acknowledged: true, insertedId: 1.23 }
test_id> db.test.insertOne({"_id": "123"})
{ acknowledged: true, insertedId: '123' }
test_id> db.test.insertOne({"_id": [1]})
{ acknowledged: true, insertedId: [ 1 ] }
test_id> db.test.insertOne({"_id": {"foo": "bar"}})
{ acknowledged: true, insertedId: { foo: 'bar' } }
test_id> db.test.find()
[
  { _id: 1.23 },
  { _id: '123' },
  { _id: [ 1 ] },
  { _id: { foo: 'bar' } }
]
```
### Tigris

Tigris API provides querying by `_id`. Let's use it.

## Testing

* Unit-tests for fetch function, that it returns not all the collection documents but a single document when `{_id: <value>}` is queried.
* Integrational tests for insert function that after the insert all expected records in database.

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

## Documentation

Document build tag in `README.md`


## Not in scope

Not in scope:
- Cases for `{_id: expression}` where expression are not scope:
  - $nor
  - $or
  - $and
  - $eqa
  - $ne
  - $gt
  - $gte
  - $lt
  - $lte
  - $in
  - $nin
  - $not
  - $regex
  - $options
  - $elemMatch
  - $size
  - $bitsAllClear
  - $bitsAllSet
  - $bitsAnyClear
  - $bitsAnySet
  - $mod
  - $exists
  - $type

