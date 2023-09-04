---
sidebar_position: 9
---

# Migrating from MongoDB to FerretDB

To ensure a smooth and successful migration from MongoDB, we offer several methods through which you can test your application with FerretDB.

## Operation modes

We offer multiple operation modes which help facilitate the testing of your application by enabling FerretDB to act as a proxy.
For more details, refer to the [official documentation](./configuration/operation-modes.md).

### Manual and automated testing with `diff-normal` mode

For details on how to install FerretDB, refer to the [quickstart guide](./quickstart-guide/).

You can manually test your application or use integration tests, among other methods.
Afterward, you can inspect the differential output for errors or inconsistencies between responses that require your attention.

As an example, let us say that your application performs some complex query and you'd like to test it in `diff-normal` mode.
You would do the following:

1. Start FerretDB in `diff-normal` mode.
   This can be achieved by using the `--mode` [flag](./configuration/flags.md) or by setting the `FERRETDB_MODE` environment variable.
2. Run `mongosh` and insert some documents.
3. Run a query to fetch the first post from each author sorted by date and author.

   Please note that due to running in `diff-normal` mode, any error returned from FerretDB will be transmitted to the client, allowing us to promptly identify the issue.
   In the majority of cases, this does not necessitate additional scrutiny of the diff output.
   Nevertheless, if FerretDB does not handle the error, additional inspection becomes necessary.

   ```sh
   # run mongosh
   $ mongosh
   ```

   ```js
   // insert some documents
   db.posts.insertMany([
     {
       title: 'title A',
       body: 'some content',
       author: 'Bob',
       date: ISODate('2023-08-29T10:33:23.134Z')
     },
     {
       title: 'another title',
       body: 'some content',
       author: 'Bob',
       date: ISODate('2023-08-28T10:33:23.134Z')
     },
     {
       title: 'title B',
       body: 'some content',
       author: 'Alice',
       date: ISODate('2023-08-20T10:33:23.134Z')
     },
     {
       title: 'some other title',
       body: 'some content',
       author: 'Alice',
       date: ISODate('2023-08-21T10:33:23.134Z')
     }
   ])

   // run a query
   db.posts.aggregate([
     { $sort: { date: 1, author: 1 } },
     {
       $group: {
         _id: '$author',
         firstPost: { $first: '$date' }
       }
     }
   ])
   // the below error is returned to the client:
   // MongoServerError: $group accumulator "$first" is not implemented yet
   ```

### Manual and automated testing with `diff-proxy` mode

Continuing with the same example above, we can further examine the diff output while in `diff-proxy` mode.

1. Run FerretDB in `diff-proxy` mode.
   This can again be achieved by using the `--mode` [flag](./configuration/flags.md) or by setting the `FERRETDB_MODE` environment variable as already shown.
2. Re-run the query.

   ```js
   db.posts.aggregate([
     { $sort: { date: 1, author: 1 } },
     {
       $group: {
         _id: '$author',
         firstPost: { $first: '$date' }
       }
     }
   ])
   // the query was handled by MongoDB, so the following documents are returned:
   // { _id: 'Alice', firstPost: ISODate("2023-08-20T10:33:23.134Z") }
   // { _id: 'Bob', firstPost: ISODate("2023-08-28T10:33:23.134Z") }
   ```

In the diff output below, however, we have discovered that the query cannot be serviced by our application because the `$first` accumulator operator is not implemented in FerretDB.

```diff
2023-08-29T13:25:09.048+0200  WARN  // 127.0.0.1:33522 -> 127.0.0.1:27017  clientconn/conn.go:360 Header diff:
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

```text
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
# clone the repository and run the compat-tool
$ git clone https://github.com/FerretDB/amazon-documentdb-tools.git && cd amazon-documentdb-tools/compat-tool
$ python3 compat.py --directory=/path/to/myapp --version=FerretDB
```
