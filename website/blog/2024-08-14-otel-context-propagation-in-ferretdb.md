---
slug: otel-context-propagation-in-ferretdb
title: 'OpenTelemetry context propagation in FerretDB'
authors: [elena]
description: >
  In this blog post, we demonstrate how to pass tracing context to queries in FerretDB using OpenTelemetry.
tags: [observability]
---

In today's world of distributed systems, achieving reliability depends on various factors, one of which is effective observability.
OpenTelemetry (OTel) has emerged as a standard for distributed tracing, but passing context to databases remains a significant challenge,
particularly with document databases like MongoDB.
At FerretDB, we're committed to addressing this challenge.

<!--truncate-->

While adopting OpenTelemetry in application development is relatively straightforward across various programming languages,
the process becomes more complex when passing context to databases.
Most databases don't natively support tracing-related context.
Approaches like [SQLCommenter](https://google.github.io/sqlcommenter/) have been developed to bridge this gap,
enabling the connection between current trace data and database queries.
In these cases, trace information is injected at the ORM level and passed to the database via SQL comments.
Database operators can then enable query logs to link specific queries with the corresponding application requests.

But is it possible to achieve something similar with MongoDB's query language?
At FerretDB, we believe that enabling the linking of a query's lifecycle with the caller's response can help developers make
their applications more predictable and reliable.

Let's consider a SQLCommenter-like approach for MongoDB.
We can leverage MongoDB's existing features to attach metadata to queries.
One such feature is the comment field, which allows additional information to be included in MongoDB queries.
This field doesn't influence query logic but can be used to provide valuable context.

At FerretDB, we want to empower users to pass such context.
We decided to utilize the `comment` field to do it.
We parse the content of the `comment` field, and if it's a JSON document with a `ferretDB` key, we check for tracing data.
If the tracing data is present, FerretDB sets the parent's context when a span for the current operation is created.

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

Using this approach, it's relatively easy to visualize client's requests to FerretDB showing them as part of the original client trace.

However, there are some limitations.
Not all operations support the comment field.
For instance, operations like `insert` or `listCollections` do not support it.

Another challenge with the `comment` field is its handling in some MongoDB drivers, where it is typed as a string.
This means that, in principle, any string can be passed.
In our scenario, where we allow FerretDB to receive and parse
tracing data through the `comment` field, the JSON must be correctly encoded by the driver and then decoded by FerretDB.

A more robust solution would involve having a dedicated field for passing request-related context,
and it would be ideal to establish a standard for such fields.
For instance, it could be a BSON document with particular tracing-related fields.
This would provide a more reliable method for passing context to the database.

Since such a standard does not yet exist, let's explore an example application that interacts
with FerretDB and passes the tracing context using the `comment` field.

For simplicity, most error handling is omitted in the example below.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    sdkresource "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
    exporter := otlptracehttp.NewUnstarted(
        otlptracehttp.WithEndpointURL("http://127.0.0.1:4318/v1/traces"),
        otlptracehttp.WithTimeout(10*time.Second),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(sdkresource.NewSchemaless(
            semconv.ServiceName("test-app"),
        )),
    )

    otel.SetTracerProvider(tp)

    defer func() {
        _ = tp.Shutdown(context.Background())
    }()

    tracer := otel.Tracer("")

    ctx, span := tracer.Start(context.Background(), "main")
    defer span.End()

    clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(ctx, clientOpts)

    if err != nil {
        log.Panicf("Failed to connect to FerretDB: %v", err)
    }

    defer func() {
        _ = client.Disconnect(ctx)
    }()

    collection := client.Database("testdb").Collection("customers")

    user := bson.D{
        {Key: "name", Value: "John Doe"},
        {Key: "email", Value: "john.doe@example.com"},
    }

    insertCtx, insertSpan := tracer.Start(ctx, "InsertCustomer")
    _, _ = collection.InsertOne(insertCtx, user)
    insertSpan.End()

    findCtx, findSpan := tracer.Start(ctx, "FindCustomer")
    filter := bson.D{{Key: "name", Value: "John Doe"}}

    var result bson.D
    _ = collection.FindOne(findCtx, filter).Decode(&result)
    findSpan.End()

    fmt.Printf("Found document: %v\n", result)
}
```

In this setup, we initialize the OpenTelemetry tracer provider and configure it to export traces to a local Jaeger endpoint.

If we run this application, we will see that the spans for the `InsertCustomer` and `FindCustomer` operations are created:

[![Trace without propagation](/img/blog/ferretdb-otel/without-propagation.png)](/img/blog/ferretdb-otel/without-propagation.png)

Let's add more details.
If we pass the tracing context through the `comment` field, we will see that the spans are linked.
Let's modify the `FindCustomer` part:

```go
    findCtx, findSpan := tracer.Start(ctx, "FindCustomer")
    filter := bson.D{{Key: "name", Value: "John Doe"}}

    traceID := findSpan.SpanContext().TraceID().String()
    spanID := findSpan.SpanContext().SpanID().String()

    traceContext := struct {
        TraceID string `json:"traceID"`
        SpanID  string `json:"spanID"`
    }{
        TraceID: traceID,
        SpanID:  spanID,
    }

    comment, _ := json.Marshal(map[string]interface{}{
        "ferretDB": traceContext,
    })

    var result bson.D
    _ = collection.FindOne(findCtx, filter, options.FindOne().SetComment(string(comment))).Decode(&result)
    findSpan.End()
```

Now, if we run the application, we will see that the spans are linked and the `FindCustomer` span is a child span created on the FerretDB side:

[![Trace with propagation](/img/blog/ferretdb-otel/with-propagation.png)](/img/blog/ferretdb-otel/with-propagation.png)

This approach gives more insights into the FerretDB's behavior and helps to understand the exact query lifecycle,
making it easier to diagnose and understand performance issues or unexpected behavior.

While this solution isn't perfect due to the limitations discussed, it is a step towards better context propagation in document databases.

In conclusion, we believe that passing context to document databases is an important part of making them more observable.
We hope that the community will come up with a standard way to pass context to databases, and we are looking forward to
contributing to this effort.

We'd love to hear your thoughts on this approach.
Have you implemented similar context propagation strategies in your projects?
Please feel free to reach out to us to share your experiences or ask questions!
