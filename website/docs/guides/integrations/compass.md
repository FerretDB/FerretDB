---
sidebar_position: 2
---

# Compass

Compass is a graphical user interface (GUI) for MongoDB-compatible applications that allows you to visually explore and run queries with your data.

## Prerequisites

- Connection string to running FerretDB instance ([See installation page here](../../installation/ferretdb/docker.md))
- Compass installed on your machine

## Connect to FerretDB instance

Connect Compass to your FerretDB instance by following these steps:'

1. Open Compass and click on "Add new connection".
2. In the "New Connection" window, set the connection string for your FerretDB instance.
   The connection string should look like this:

   ```sh
   mongodb://<username>:<password>@<host>:<port>/<database>
   ```

   ![Compass](/img/docs/compass-connection.png)

3. Click on "Connect" to establish a connection to your FerretDB instance.
   You should now be able to explore your data using Compass.

## Explore and query data

Run queries, create indexes, and perform other operations on your data using Compass.

![Compass](/img/docs/explore-compass-data.png)

## Additional information

- [Authentication in FerretDB](../../security/authentication.md)
- [Troubleshooting](../../troubleshooting/overview.md)
