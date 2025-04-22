---
sidebar_position: 1
---

<!-- Category for Command "Administrative, Diagnostic, and Utility Commands" should be organized per folder as present in the compatibility document -->

# `createIndexes` Command

<!-- Short description of the command and its usage. -->

Creates one or more indexes on the specified collection.

Indexes improve query performance and enable certain query features (e.g., uniqueness, sorting, partial matches).
This command accepts a list of index specifications and applies them to a given collection.

<!-- arguments,modieifers, stage, operators, parameters, options supported by the command in a table (if necessary or available). Takes in the following parameters: -->

## Parameters

| Argument/Parameter        | Type    | Description                                                                  |
| ------------------------- | ------- | ---------------------------------------------------------------------------- |
| `createIndexes`           | string  | The name of the collection on which to create the index.                     |
| `indexes`                 | array   | An array of index specification objects. Each object defines one index.      |
| `key`                     | object  | The fields to index and their sort order or type (`1`, `-1`, `"text"`, etc). |
| `name`                    | string  | A required name for the index. Must be unique per collection.                |
| `unique`                  | bool    | Optional. If `true`, ensures all values are unique for the indexed field.    |
| `partialFilterExpression` | object  | Optional. Restricts index to documents matching the condition.               |
| `expireAfterSeconds`      | integer | Optional. For TTL indexes. Automatically deletes documents after set time.   |

## Example Usage

### 1. Create a Unique Index

```js
db.runCommand({
  createIndexes: 'users',
  indexes: [
    {
      key: { email: 1 },
      name: 'unique_email_idx',
      unique: true
    }
  ]
})
```

This ensures that the `email` field is unique across all documents in the `users` collection.

### 2. Create Compound Indexes

```js
db.runCommand({
  createIndexes: 'orders',
  indexes: [
    {
      key: { customerId: 1, orderDate: -1 },
      name: 'customer_order_idx'
    }
  ]
})
```

This compound index helps optimize queries filtering or sorting on both `customerId` and `orderDate`.

### 3. Create a Partial Index

```js
db.runCommand({
  createIndexes: 'sessions',
  indexes: [
    {
      key: { status: 1 },
      name: 'active_sessions_idx',
      partialFilterExpression: { status: { $eq: 'active' } }
    }
  ]
})
```

The index only applies to documents where `status` is `"active"`, reducing index size and write overhead.

## Additional Information

Please find advanced index creation examples here:

- [Vector Indexes](../../guides/vector-search.mdx)
- [TTL Indexes](../../guides/ttl-indexes.mdx)
- [Full-Text Indexes](../../guides/full-text-search.mdx)

## Limitations or Constraints
