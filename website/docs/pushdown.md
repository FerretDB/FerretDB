---
sidebar_position: 1
---

# Query pushdown

Pushdown is the method of optimizing a query by reducing the amount of data read and processed. 
It saves memory space, and network bandwidth, and reduces the query execution time by not prefetching 
unnecessary data to the database management system.

As FerretDB mission is to become true open-source MongoDB-compatible database, it's important
to handle all operations, comparisons, data types, and commands in the same fashion as MongoDB.

Doing so can be really challenging on different database backends. For example, it's difficult to
make PostgreSQL abide BSON types comparison order. There are other things like the fact that PostgreSQL
uses numeric type to floating-point numbers, which is not IEEE 754 compilant (like MongoDB), and creates
some comparison issues on larger values.

Even, if it's possible, the SQL query that checks for all of the logic might be huge and really hard to maintain.

Because of that, FerretDB fetch all necessary data from collection and applies filters on it's own. 
That solution allows to implement all complicated logic safely, easily and quickly. Maintainers are not overwhelmed
by database backend specifics, which makes FerretDB a great, universal solution to implement new backends.

The downside of this is obvious - fetching much data from collection on every single query can be inefficient and
time-consuming (especially for larger collections).

Query pushdown is a great compromise between complexity and performance.
It allows to handle all MongoDB logic without many issues related to underyling database architecture,
while keeping good query performance.


The following table shows all operators and types that FerretDB pushdowns on PostgreSQL backend.
If filter use specified type and operator, FerretDB will prefetch less data and the query will be more performent.

As pushdown approach allows developers to implement query optimizations later, the table will 
be updated with every change.

If your application requires better performance for specific operation, feel free to join slack channel.

|            | Object   | Array   | Double                 | String   | Binary   | ObjectID   | Boolean   | Date   | Null   | Regex   | Integer   | Timestamp   | Long                   |
| ---------- | -------- | ------- | --------               | -------- | -------- | ---------- | --------- | ------ | ------ | ------- | --------- | ----------- | ------                 |
| `=`        | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub> | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub> |
| `$eq`      | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub> | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub>       |
| `$gt`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$gte`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$lt`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$lte`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$in`      | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |
| `$ne`      | ✖️        | ✖️       | ⚠️ <sub>[[1]](#1)</sub>       | ✅       | ✖️        | ✅         | ✅        | ✅     | ✖️      | ✖️       | ✅        | ✖️           | ⚠️ <sub>[[1]](#1)</sub>       |
| `$nin`     | ✖️        | ✖️       | ✖️                      | ✖️        | ✖️        | ✖️          | ✖️         | ✖️      | ✖️      | ✖️       | ✖️         | ✖️           | ✖️                      |


###### [1] {#1}
Numbers outside the range of the safe IEEE 754 precision (`< -9007199254740991.0, 9007199254740991.0 >`), 
will prefetch all numbers larger/smaller than max/min value of the range.
