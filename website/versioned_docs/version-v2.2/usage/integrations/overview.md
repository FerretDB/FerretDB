---
sidebar_position: 1
---

# Overview

This section focuses on how to integrate FerretDB with other third-party tools and services.

## GUI applications

Most GUI applications follow the same process.
You can connect and explore your data using the connection string of your FerretDB instance.
Your connection string should look like this:

```text
mongodb://<username>:<password>@<host>:<port>/<database>
```

Once a connection is established, you can explore your data using the GUI application.

For example, below we showcase a connection to FerretDB using Compass.

![GUI connection to Compass showing serverStatus](/img/docs/gui-connection.jpg)

The image shows the `serverStatus` command being run in Compass on a FerretDB instance.
