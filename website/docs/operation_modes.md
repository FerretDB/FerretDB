---
---

# Operation modes

To stay compatible with MongoDB drivers, FerretDB introduces a couple of operation modes.
Operation modes specify how FerretDB handles incoming requests.
They might be mostly useful for testing, debugging, or bug reporting.

You can specify modes by using the `--mode` flag or `FERRETDB_MODE` variable,
which accept following types of values: `normal`, `proxy`, `diff-normal`, `diff-proxy`.

By default FerretDB always run on `normal` mode, which means that all client requests
are processed only by FerretDB and returned to the client.

## Proxy

Proxy is another MongoDB-compatible database, accessible from the machine.
You can specify its connection URL with `--proxy-addr` flag or with the `FERRETDB_PROXY_ADDR` variable.

To forward all requests to proxy and return them to the client, use `proxy` operation mode.

## Diff modes

Diff modes (`diff-normal`, `diff-proxy`) forward requests to both databases, and log the difference between them.

The `diff-normal` afterwards returns the response from FerretDB and `diff-proxy` - from the specified proxy handler.

Example diff output:

```sh
--- res header
+++ proxy header
@@ -1 +1 @@
-length:   100, id:    4, response_to:   14, opcode: OP_MSG
+length:   100, id:   16, response_to:   14, opcode: OP_MSG

Body diff:
--- res body
+++ proxy body
@@ -12,7 +12,6 @@
           "$k": [
-            "firstBatch",
             "id",
-            "ns"
+            "ns",
+            "firstBatch"
           ],
-          "firstBatch": [],
           "id": {
@@ -20,3 +19,4 @@
           },
-          "ns": "test.values"
+          "ns": "test.values",
+          "firstBatch": []
         },
```
