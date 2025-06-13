---
slug: visually-explore-ferretdb-data-mongo-express
title: 'Visually Explore Your FerretDB Data with MongoDB Compass'
authors: [alex]
description: >
  Connect MongoDB Compass to your FerretDB instances to visually explore and manage your PostgreSQL-backed document data. This guide shows you how to get started easily.
image: /img/blog/ferretdb-compass.jpg
tags: [mongodb gui, compatible applications, tutorial]
---

![Visually Explore Your FerretDB Data with MongoDB Compass](/img/blog/ferretdb-compass.jpg)

Managing and exploring your database can be significantly streamlined with the right graphical user interface (GUI) tool.

<!--truncate-->

For MongoDB users,[MongoDB Compass](https://www.mongodb.com/products/compass) is a go-to choice, offering a rich environment for data exploration, query building, and performance monitoring.

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this guide, we're excited to show you how effortlessly you can connect and manage your FerretDB databases using MongoDB Compass.

## Connecting MongoDB Compass to FerretDB

Connecting MongoDB Compass to your FerretDB instance is very straightforward, following the standard MongoDB connection process.
Here's what you need to do:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
2. **Launch MongoDB Compass:** Open the MongoDB Compass application on your system.
   If you don't have it installed, you can download it from the [MongoDB Compass download page](https://www.mongodb.com/try/download/compass).
3. Configure the Connection:
   In the Compass connection window, you'll typically input a connection string.
   Your FerretDB connection string should look like this:

   ```sh
   mongodb://<username>:<password>@<host>:<port>/<database>
   ```

   Replace `<username>`, `<password>`, `<host>`, `<port>`, and `<database>` with your FerretDB instance details.

   Paste this into the URI field in Compass and click Connect.

   ![Example of FerretDB connection string in MongoDB Compass](/img/blog/compass-connection-string.png)

4. **Connect and Explore:** Once connected, Compass will display your FerretDB databases and collections.
   You can now visually explore your data, run queries, and perform various administrative tasks.

## Example: Running `serverStatus` Command in Compass

As an example, here we showcase the `serverStatus` command being run in Compass on a FerretDB instance.
This command provides an overview of the server's status and statistics.

The image clearly demonstrates that Compass is successfully communicating with FerretDB and retrieving server-side information, just as it would with a native MongoDB instance.
You can navigate to the "Performance" or "Shell" tab within Compass to execute such commands.

![GUI connection to Compass showing serverStatus](/img/blog/compass-serverstatus.png)

## Conclusion

Integrating MongoDB Compass with FerretDB provides a familiar, powerful, and intuitive graphical interface for managing your FerretDB, a PostgreSQL-backed document databases.
You can also leverage other existing tools in your ecosystem while benefiting from the reliability and features of PostgreSQL.

We're continuously working to expand our integrations and support for various tools.
Stay tuned for more updates, and feel free to reach out to us if you have any questions or suggestions!

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
