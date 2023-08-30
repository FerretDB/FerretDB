---
sidebar_position: 3
---

# Migrating from MongoDB to FerretDB

To ensure a smooth and successful migration from MongoDB, we offer several methods through which you can test your application with FerretDB.

## Operation modes

We support different operation modes, by default FerretDB will always run on `normal` mode.
In this mode all client requests are processed _only_ by FerretDB and returned to the client.

You can set modes using either the `--mode` flag or the `FERRETDB_MODE` variable, accepting values such as `normal`, `proxy`, `diff-normal`, and `diff-proxy`.

### Manual and automated testing with `diff-normal` mode

`diff-normal` mode acts a proxy and will forward client requests to another MongoDB-compatible database, and will log the difference between them.

You can manually test your application or use integration tests, among other methods.
Afterward, you can inspect the differential output for errors or inconsistencies between responses that require your attention.

As an example, let us say that your application performs some complex query and you'd like to test it in `diff-normal` mode.
You would do the following:

1. Start the environment and run FerretDB.

```sh
# in a terminal run env-up to start the environment
me@foobar:~/FerretDB$ bin/task env-up
# in another terminal run the debug build which runs in diff-normal mode by default
me@foobar:~/FerretDB$ bin/task run
```

Please note that due to running in `diff-normal` mode, any error returned from FerretDB will be transmitted to the client.
In the majority of cases, this does not necessitate additional scrutiny of the diff output.
Nevertheless, if FerretDB does not handle the error, additional inspection becomes necessary.

```sh
# run mongosh
me@foobar:~/FerretDB$ mongosh --quiet
```
```js
// insert some documents
test> db.posts.insertMany([
...   {
...     title: 'title A',
...     body: 'some content',
...     author: 'Bob',
...     date: ISODate("2023-08-29T10:33:23.134Z"),
...   },
...   {
...     title: 'another title',
...     body: 'some content',
...     author: 'Bob',
...     date: ISODate("2023-08-28T10:33:23.134Z"),
...   },
...   {
...     title: 'title B',
...     body: 'some content',
...     author: 'Alice',
...     date: ISODate("2023-08-20T10:33:23.134Z"),
...   },
...   {
...     title: 'some other title',
...     body: 'some content',
...     author: 'Alice',
...     date: ISODate("2023-08-21T10:33:23.134Z"),
...   },
... ]);
{
  acknowledged: true,
  insertedIds: {
    '0': ObjectId("64edcfc975dd10e7bfb36add"),
    '1': ObjectId("64edcfc975dd10e7bfb36ade"),
    '2': ObjectId("64edcfc975dd10e7bfb36adf"),
    '3': ObjectId("64edcfc975dd10e7bfb36ae0")
  }
}
test> // run a query to fetch the first post from each author sorted by date and author
test> db.posts.aggregate(
...   [
...     { $sort: { date: 1, author: 1 } },
...     {
...       $group:
...         {
...           _id: "$author",
...           firstPost: { $first: "$date" }
...         }
...     }
...   ]
... );
MongoServerError: $group accumulator "$first" is not implemented yet
test>
```

### Manual and automated testing with `diff-proxy` mode

`diff-proxy` mode will forward client requests to a MongoDB-compatible database, but return the proxy responses to the client, and will log the difference between them.

Continuing with the same example above, we can further examine the diff output.

1. Run FerretDB in `diff-proxy` mode: `bin/task run-proxy`.
2. Re-run the query.

```js
test> db.posts.aggregate(
...   [
...     { $sort: { date: 1, author: 1 } },
...     {
...       $group:
...         {
...           _id: "$author",
...           firstPost: { $first: "$date" }
...         }
...     }
...   ]
... );
[
  { _id: 'Alice', firstPost: ISODate("2023-08-20T10:33:23.134Z") },
  { _id: 'Bob', firstPost: ISODate("2023-08-28T10:33:23.134Z") }
]
test>
```

In the diff output below we have discovered that the query cannot be serviced by our application because the `$first` accumulator operator is not implemented in FerretDB.

```sh
2023-08-29T13:25:09.048+0200	WARN	// 127.0.0.1:33522 -> 127.0.0.1:27017 	clientconn/conn.go:360	Header diff:
--- res header
+++ proxy header
@@ -1 +1 @@
-length:   140, id:    2, response_to:  156, opcode: OP_MSG
+length:   181, id:  360, response_to:  156, opcode: OP_MSG

Body diff:
--- res body
+++ proxy body
@@ -7,13 +7,41 @@
         "$k": [
-          "ok",
-          "errmsg",
-          "code",
-          "codeName"
+          "cursor",
+          "ok"
         ],
+        "cursor": {
+          "$k": [
+            "firstBatch",
+            "id",
+            "ns"
+          ],
+          "firstBatch": [
+            {
+              "$k": [
+                "_id",
+                "firstPost"
+              ],
+              "_id": "Alice",
+              "firstPost": {
+                "$d": 1692527603134
+              }
+            },
+            {
+              "$k": [
+                "_id",
+                "firstPost"
+              ],
+              "_id": "Bob",
+              "firstPost": {
+                "$d": 1693218803134
+              }
+            }
+          ],
+          "id": {
+            "$l": "0"
+          },
+          "ns": "test.posts"
+        },
         "ok": {
-          "$f": 0
-        },
-        "errmsg": "$group accumulator \"$first\" is not implemented yet",
-        "code": 238,
-        "codeName": "NotImplemented"
+          "$f": 1
+        }
       },
```

### Response metrics

Metrics are captured and written to standard output (`stdout`) upon exiting in [Debug builds](https://pkg.go.dev/github.com/FerretDB/FerretDB@v1.8.0/build/version#hdr-Debug_builds).
This is a useful way to quickly determine what commands are not implemented for the client requests sent by your application.

```sh
# we ran task run-proxy and then sent an interrupt ctrl+c after some time
^Ctask: Signal received: "interrupt"
# HELP ferretdb_client_requests_total Total number of requests.
# TYPE ferretdb_client_requests_total counter
ferretdb_client_requests_total{command="aggregate",opcode="OP_MSG"} 105
ferretdb_client_requests_total{command="find",opcode="OP_MSG"} 398
ferretdb_client_requests_total{command="findAndModify",opcode="OP_MSG"} 6
ferretdb_client_requests_total{command="hello",opcode="OP_MSG"} 4
ferretdb_client_requests_total{command="insert",opcode="OP_MSG"} 10
ferretdb_client_requests_total{command="ismaster",opcode="OP_MSG"} 17
ferretdb_client_requests_total{command="unknown",opcode="OP_QUERY"} 28
ferretdb_client_requests_total{command="update",opcode="OP_MSG"} 59
# HELP ferretdb_client_responses_total Total number of responses.
# TYPE ferretdb_client_responses_total counter
ferretdb_client_responses_total{argument="$first (accumulator)",command="aggregate",opcode="OP_MSG",result="NotImplemented"} 18
ferretdb_client_responses_total{argument="fields",command="findAndModify",opcode="OP_MSG",result="NotImplemented"} 6
ferretdb_client_responses_total{argument="unknown",command="aggregate",opcode="OP_MSG",result="ok"} 87
ferretdb_client_responses_total{argument="unknown",command="find",opcode="OP_MSG",result="ok"} 398
ferretdb_client_responses_total{argument="unknown",command="hello",opcode="OP_MSG",result="ok"} 4
ferretdb_client_responses_total{argument="unknown",command="insert",opcode="OP_MSG",result="ok"} 10
ferretdb_client_responses_total{argument="unknown",command="ismaster",opcode="OP_MSG",result="ok"} 17
ferretdb_client_responses_total{argument="unknown",command="unknown",opcode="OP_REPLY",result="ok"} 28
ferretdb_client_responses_total{argument="unknown",command="update",opcode="OP_MSG",result="ok"} 59
```

### Other tools

We also have a fork of the Amazon DocumentDB Compatibility Tool [here](https://github.com/FerretDB/amazon-documentdb-tools/tree/master/compat-tool).
The tool examines files to identify queries that use unsupported operators in FerretDB.
Please note that this tool is not highly accurate and may generate inaccurate reports, as it does not parse query syntax with contextual information about the originating command.
For example, an unsupported operator might appear within a `find` or `aggregate` command, which the tool does not differentiate.
Note that we also mark operators as unsupported if they are not supported in _all_ commands, which could result in false negatives.

Running the tool to check FerretDB compatibility:

```sh
me@foobar:~$ git clone https://github.com/FerretDB/amazon-documentdb-tools.git && cd amazon-documentdb-tools/compat-tool
me@foobar:~amazon-documentdb-tools/compat-tool$ python3 compat.py --directory=/path/to/myapp --version=FerretDB
```
