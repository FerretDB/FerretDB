---
sidebar_position: 7
hide_table_of_contents: true
---

# Query pushdown

**Query pushdown** is the method of optimizing a query by reducing the amount of data read and processed.
It saves memory space, network bandwidth, and reduces the query execution time by moving some parts
of the query execution closer to the data source.

Initially FerretDB retrieved all data related to queried collection, and applies filters on its own, making
it possible to implement complex logic safely and quickly.
To make this process more efficient, we minimize the amount of incoming data, by applying WHERE clause on SQL queries.

:::info
You can learn more about query pushdown in our [blog post](https://blog.ferretdb.io/ferretdb-fetches-data-query-pushdown/).
:::

## Supported types and operators

The following table shows all operators and types that FerretDB supports for query pushdown on PostgreSQL backend.
If filter uses type and operator, that's marked as pushdown-supported on this list,
FerretDB will prefetch fewer results, resulting in more performant query.

If your application requires better performance for specific operation,
feel free to share this with us in our [community](/#community)!

:::tip
As query pushdown allows developers to implement query optimizations separately from the features,
the table will be updated frequently.
:::

<!-- markdownlint-capture -->
<!-- markdownlint-disable MD001 MD033 MD051 -->

|        | Object | Array | Double                  | String | Binary | ObjectID | Boolean | Date | Null | Regex | Integer | Timestamp | Long                    |
| ------ | ------ | ----- | ----------------------- | ------ | ------ | -------- | ------- | ---- | ---- | ----- | ------- | --------- | ----------------------- |
| `=`    | ✖️     | ✖️    | ⚠️ <sub>[[1]](#1)</sub> | ✅     | ✖️     | ✅       | ✅      | ✅   | ✖️   | ✖️    | ✅      | ✖️        | ⚠️ <sub>[[1]](#1)</sub> |
| `$eq`  | ✖️     | ✖️    | ⚠️ <sub>[[1]](#1)</sub> | ✅     | ✖️     | ✅       | ✅      | ✅   | ✖️   | ✖️    | ✅      | ✖️        | ⚠️ <sub>[[1]](#1)</sub> |
| `$gt`  | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |
| `$gte` | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |
| `$lt`  | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |
| `$lte` | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |
| `$in`  | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |
| `$ne`  | ✖️     | ✖️    | ⚠️ <sub>[[1]](#1)</sub> | ✅     | ✖️     | ✅       | ✅      | ✅   | ✖️   | ✖️    | ✅      | ✖️        | ⚠️ <sub>[[1]](#1)</sub> |
| `$nin` | ✖️     | ✖️    | ✖️                      | ✖️     | ✖️     | ✖️       | ✖️      | ✖️   | ✖️   | ✖️    | ✖️      | ✖️        | ✖️                      |

###### [1] {#1}

Numbers outside the range of the safe IEEE 754 precision (`< -9007199254740991.0, 9007199254740991.0 >`),
will prefetch all numbers larger/smaller than max/min value of the range.

<!-- markdownlint-restore -->

## Supported pushdown on `find` command arguments

The following table shows supported pushdown and combinations of pushdown on `find` command arguments for the PostgreSQL backend.
It applies `WHERE` clause for `filter` argument, `LIMIT` clause for `limit` argument and `ORDER BY` clause for `sort` argument on SQL queries.

<!-- markdownlint-capture -->
<!-- markdownlint-disable MD001 MD033 MD051 -->

| `find` command arguments | Supported               |
| ------------------------ | ----------------------- |
| `filter`                 | ⚠️ <sub>[[2]](#2)</sub> |
| `filter`,`limit`         | ✖️                      |
| `limit`                  | ✅                      |
| `limit`, `skip`          | ✖️                      |
| `limit`, `sort`          | ⚠️ <sub>[[3]](#3)</sub> |
| `skip`                   | ✖️                      |
| `sort`                   | ⚠️ <sub>[[4]](#4)</sub> |

###### [2] {#2}

See [supported types and operators](#supported-types-and-operators).

###### [3] {#3}

When a command contains `limit` and `sort` arguments, limit pushdown is applied
only if experimental [sort pushdown configuration](configuration/flags.md#query-pushdown) is enabled by `--test-enable-sort-pushdown` flag.

###### [4] {#4}

Sort pushdown is an experimental [configuration](configuration/flags.md#query-pushdown) enabled by `--test-enable-sort-pushdown` flag.
