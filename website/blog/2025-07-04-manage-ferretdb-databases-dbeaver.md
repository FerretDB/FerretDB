---
slug: manage-ferretdb-databases-dbeaver
title: 'Manage Your FerretDB Databases with DBeaver'
authors: [alex]
description: >
  Connect DBeaver to your FerretDB instances to visually explore, query, and manage your data.
image: /img/blog/ferretdb-dbeaver.jpg
tags: [mongodb gui, compatible applications, tutorial]
---

![Manage Your FerretDB Databases with DBeaver](/img/blog/ferretdb-dbeaver.jpg)

For developers, database administrators, and data analysts, a powerful and versatile graphical user interface (GUI) is indispensable for interacting with databases.
DBeaver stands out as a universal database tool, supporting a wide array of relational and NoSQL databases, including MongoDB and FerretDB.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this guide, we're excited to show you how effortlessly you can connect and manage your FerretDB databases using DBeaver, providing a comprehensive visual management experience.

## What is DBeaver?

DBeaver is a database tool that offers a rich set of features, including:

- **Unified interface:** Connect to virtually any database (SQL and NoSQL) from a single application.
- **Visual data exploration:** Browse databases, schemas, tables/collections, and view/edit data in a user-friendly grid, JSON, or plain text format.
- **SQL/NoSQL query editor:** Execute SQL queries for relational databases and JavaScript/mongo shell queries for MongoDB-like databases.
- **Data import/export:** Easily move data between different formats and databases.
- **ER diagrams:** Generate visual representations of database schemas.
- **Administration tools:** Manage users, roles, and monitor database activity.

DBeaver's extensive support for MongoDB makes it an ideal choice for managing document databases, and by extension, FerretDB.

## Why use DBeaver with FerretDB?

DBeaver provides robust, built-in support for MongoDB.
Because FerretDB is designed to be a truly open-source alternative to MongoDB, DBeaver can connect to your FerretDB instances as well.
This powerful compatibility offers several compelling advantages:

- **Familiarity for MongoDB users:** If you're already accustomed to DBeaver for MongoDB or other databases, you can immediately apply your knowledge to manage FerretDB instances.
- **Visual data management:** Explore your PostgreSQL-backed document data through DBeaver's intuitive graphical interface, making it easier to understand, query, and debug.
- **Simplified operations:** Perform common database operations visually, from Browse collections and documents to executing queries and basic administration, without extensive command-line interactions.
- **Cross-database management:** Manage your FerretDB instances alongside any other SQL or NoSQL databases you use, all from a single DBeaver application.

## Connecting DBeaver to FerretDB

Connecting DBeaver to your FerretDB instance is a straightforward process, leveraging DBeaver's native MongoDB driver.
Here's a step-by-step guide:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
2. **Launch DBeaver:** Ensure that you have DBeaver installed on your system.
   If you don't have it yet, download it from the [DBeaver website](https://dbeaver.io/download/).
   Open the DBeaver application on your system.
3. **Create a new database connection:** Click on "New Database Connection" from the toolbar or go to Database > New Database Connection.
   In the "Connect to a database" wizard, type "MongoDB" in the search bar or navigate to the "NoSQL" section and select "MongoDB".
   Click Next.
4. **Configure connection settings:**
   In the "Main" tab, fill in the connection details by entering the host, port, and authentication information for your FerretDB instance.
   If you have [authentication enabled on your FerretDB instance](https://docs.ferretdb.io/security/authentication/), select the appropriate authentication method (e.g., "SCRAM-SHA-256") and enter your Username and Password.

   ![A screenshot of DBeaver's 'New Database Connection' wizard](/img/blog/dbeaver-connection.png)

5. **Test Connection:** Click the "Test Connection..." button.
   If the connection is successful, you'll see a confirmation message.
   Then click Finish.
   The new connection will appear in your "Database Navigator."
6. **Explore Your Data:** Expand your new FerretDB connection in the "Database Navigator".
   You can now browse databases, collections, view documents, and execute queries in the script editor or the dedicated MongoDB shell.

## Exploring FerretDB data with DBeaver

Once connected, DBeaver allows you to visually inspect your FerretDB instances.

You can also use DBeaver's editor to run queries directly against your FerretDB instance.

Here's how you might query the books collection to find all books authored by British authors:

```js
db.books.find({ 'authors.nationality': 'British' })
```

Running this in DBeaver's script editor would yield the matching documents, which you can see in the results grid.

![DBeaver's data editor showing the books collection](/img/blog/dbeaver-collection.png)

This output demonstrates DBeaver's seamless interaction with FerretDB, allowing you to manage and query your document data using familiar tools and methods.

## Conclusion

The integration of DBeaver and FerretDB provides a powerful, versatile, and fully open-source solution for visually managing your document databases.
By leveraging FerretDB, you can seamlessly integrate it into your existing database management workflows with DBeaver.
This enables developers to run their MongoDB workloads in the open-source ecosystem, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
