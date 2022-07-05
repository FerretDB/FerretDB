* `{_id: <value>}`

## The problem
Our fetch methods read all documents before returning them to the user.
That requires too much memory.

## Solution

Let's go step by step and first, implement a simple query pushdown for queries containing `{_id: <value>}`,
i.e. add SQL condition `_id = $<placeholder>` in WHERE clause passed to the PostgreSQL backend.

* Add a function to `pgPool` that queries `_id`  and returns a single record, nothing or error.
* In `pg` handler, modify `(h *Handler) fetch`, add the usage of that function.

Changes affect:
- count
- delete
- find
- find and modify
- update

### Delete

Now, the `delete` function loads records that we are about to delete, before the deletion.
To not change the behavior, let's add a build tag:
If the build tag is enabled:
* then for queries `{_id: <value>}`, use a simple pushdown,
*  for all other queries process as usual.

As it is necessary to make the code consistent, let's work through the build tag with all other commands.

## Not in scope

Not in scope:
- Tigris handler .
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


## Testing

Unit-tests for fetch function, that it returns not all the collection documents but a single document when `{_id: <value>}` is queried.

## Border cases

Checks for `_id`:
* too long `_id` variable
* empty `_id`
* `_id` containing binary data
* attempt for SQL injection

## Documentation

Document build tag in `README.md`
