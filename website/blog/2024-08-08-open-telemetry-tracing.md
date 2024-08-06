---
slug: <tbd>
title: <tbd>
authors: [elena]
description: >
  <tbd>
tags: [<tbd>]
---

Nowadays, distributed systems have become a norm in software development. When we talk about making such systems 
reliable we often mention observability as a tool to get feedback  and troubleshoot such systems. 
This is where distributed tracing comes in, and where OpenTelemetry offers a standard to use it.

While it is relatively easy to start using OpenTelemtry when writing applications in various programming languages, 
passing the context to databases is more complicated. Most databases don’t support native approach to work with tracing-related context. 
There are some approaches such as SQLCommenter (https://google.github.io/sqlcommenter/) to make it possible to connect 
the data of the current trace and the queries executed in the DB. In this case, the information about the trace is injected 
on the ORM level and passed through SQL comments to the database. The operator of the database can enable query logs to link 
the exact queries with the exact requests of the caller application.

But what about document databases? Would it be possible to do something similar with MongoDB query language? At FerretDB 
we believe that giving an ability to link the exact query lifecycle with the caller’s response can help developers make 
their applications more predictable and reliable.

So, let’s say we want to try an approach similar to SQLCommenter. We need to leverage the existing capabilities that MongoDB 
offers to attach metadata to queries. One such feature is the comment field, which can be used to include additional information 
in MongoDB queries. This field is not executed as part of the query logic but can be used to provide some additional context.

At FerretDB, we want to give users an ability to pass such context. So, we decided to use the `comment` field to do it. 
We parse the comment field, and if it’s a json document with the `ferretDB` key being set, we check whether it contains the tracing data. 
If so, FerretDB sets parent’s context when a span for the current operation is created.

An example of a comment with tracing data could look like this:


```json
{
  "ferretDB": {
    "traceID": "1234567890abcdef1234567890abcdef",
    "spanID": "fedcba9876543210"
  }
}
```

The strings `traceID` and `spanID` are hex-encoded strings that represent the corresponding trace and span identifiers. 
For example, the JSON above represents the following data:

```go
traceID := [16]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef}
spanID := [8]byte{0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}
``` 

With such an approach, it’s relatively easy to visualize client’s requests to FerretDB showing them as part of the initial client’s trace.

However, not all the operations support `comment` field. For example, `insert` or `listCollections` operations don’t have `comment` field.

Another problem with `comment` field is that it’s typed as string, and, in principle, any string can be passed there. 
In our example where we let FerretDB receive and parse tracing data through the comment, we need to decode the json passed 
in the comment (and the driver needs to encode it).

A much better approach would be to have a special field to pass some additional request-related context. 
