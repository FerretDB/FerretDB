---
sidebar_position: 1
---

# Query pushdown

**Query pushdown** is the method of optimizing a query by reducing the amount of data read and processed.
It saves memory space, network bandwidth, and reduces the query execution time by not prefetching
unnecessary data to the database management system.

:::info
You can read more about query pushdown in our [blog post](https://blog.ferretdb.io/ferretdb-fetches-data-query-pushdown/)!
:::

## Why do we need query pushdown

As FerretDB mission is to become true open-source MongoDB-compatible database, it's important
to handle all operations, comparisons, data types, and commands in the same fashion as MongoDB.

Doing so can be really challenging on different database backends.
For example, it's difficult to
make PostgreSQL abide BSON types comparison order, or compare large floating-point values in the IEEE 754 fashion.
And even, if some things are possible, the SQL query that implements the logic might be huge and really hard to maintain.

Because of that, FerretDB fetch all necessary data from collection and applies filters on it's own.
That solution allows to implement all complicated logic safely, easily and quickly.
Maintainers are not overwhelmed
by database backend specifics, which also makes FerretDB a great, universal solution to implement new backends.

The downside of this is obvious - fetching much data from collection on every single query can be inefficient and
time-consuming (especially for larger collections).

Query pushdown is a great compromise between complexity and performance.
It allows to handle all MongoDB logic without many issues related to underyling database architecture,
while still keeping a good query performance.

## Query pushdown supported filters

The following table shows all operators and types that FerretDB pushdowns on PostgreSQL backend.
If filter uses type and operator, that's marked as pushdown supported on this list,
FerretDB will prefetch less data, resulting with more performent query.

If your application requires better performance for specific operation,
feel free to share this with us on [FerretDB Slack](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A)!

:::tip
As query pushdown allows developers to implement query optimizations separately from the features,
the table will be updated frequently.
:::

|            | Object   | Array   | Double                 | String   | Binary   | ObjectID   | Boolean   | Date   | Null   | Regex   | Integer   | Timestamp   | Long                   |
| ---------- | -------- | ------- | --------               | -------- | -------- | ---------- | --------- | ------ | ------ | ------- | --------- | ----------- | ------                 |
| `=`        | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub> | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub> |
| `$eq`      | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub> | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub> |
| `$gt`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$gte`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$lt`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$lte`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$in`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$ne`      | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub> | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub> |
| `$nin`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |

###### [1] {#1}
Numbers outside the range of the safe IEEE 754 precision (`< -9007199254740991.0, 9007199254740991.0 >`),
will prefetch all numbers larger/smaller than max/min value of the range.
