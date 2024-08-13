---
slug: otel-context-propagation-in-ferretdb
title: 'OpenTelemetry context propagation in FerretDB'
authors: [elena]
description: >
  In this blog post, we demonstrate how to pass tracing context to queries in FerretDB using OpenTelemetry.
tags: [observability, opentelemetry, tracing]
---

Nowadays, distributed systems have become a norm in software development. When we talk about making such systems 
reliable we often mention observability as a tool to get feedback and troubleshoot such systems. 
This is where distributed tracing comes in, and where OpenTelemetry offers a standard to use it.

While it is relatively easy to start using OpenTelemetry when writing applications in various programming languages, 
passing the context to databases is more complicated. Most databases don’t support native way of working with tracing-related context. 
There are some approaches such as [SQLCommenter](https://google.github.io/sqlcommenter/) to make it possible to connect 
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

However, not all the operations support `comment` field. For example, `insert` or `listCollections` operations don’t have it.

Another problem with `comment` field is that it’s typed as string, and, in principle, any string can be passed there. 
In our example where we let FerretDB receive and parse tracing data through the comment, we need to decode the json passed 
in the comment (and the driver needs to encode it).

A much better approach would be to have a special field to pass some additional request-related context, and it would
be great to agree on a standard for such fields. This way, we could have a more reliable way to pass context to the database.

However, as such standard doesn't exist yet, let's take a look at an example application that works with FerretDB and
passes the tracing context through the `comment` field.

For simplicity, most of error handling is omitted in the example below.

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		_ = tp.Shutdown(shutdownCtx)
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

If we run this application, we will see that the spans for the `InsertCustomer` and `FindCustomer` operations are created:

[![Trace without propagation](/img/blog/ferretdb-otel/without-propagation.png)](/img/blog/ferretdb-otel/without-propagation.png)

Now let's add more details. If we pass the tracing context through the `comment` field, we will see that the spans are linked.
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

This approach gives more insights into the FerretDB's behavior and helps to understand the exact query lifecycle.

In conclusion, we believe that passing context to databases is an important part of making document databases more observable.
We hope that the community will come up with a standard way to pass context to databases, and we are looking forward to
contributing to this effort.

We’d love to hear your thoughts on this approach. Have you implemented similar context propagation strategies in your projects? 
Please feel free to reach out to us to share your experiences or ask questions!
