---
slug: building-scalable-community-forums-nodebb-and-ferretdb
title: 'Building Scalable Community Forums with NodeBB and FerretDB'
authors: [alex]
description: >
  Discover how to power your NodeBB community forum with FerretDB, leveraging the flexibility of a document database and the reliability of PostgreSQL.
image: /img/blog/ferretdb-nodebb.jpg
tags: [compatible applications, open source, community]
---

![Building Scalable Community Forums with NodeBB and FerretDB](/img/blog/ferretdb-nodebb.jpg)

Creating and nurturing online communities is vital for many organizations and projects.
NodeBB stands out as a powerful, modern forum software, built for real-time interaction and scalability that uses MongoDB as its primary database.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.
This unique combination allows you to benefit from the flexible document model while enjoying the ACID compliance and robustness of a traditional relational database.

In this guide, we're excited to explore how [NodeBB](https://nodebb.org/), the popular open-source forum platform, seamlessly integrates with FerretDB, ensuring a fully open-source stack for your online community.

## What is NodeBB?

NodeBB is a next-generation forum software built with Node.js.
Unlike traditional forum platforms, NodeBB offers real-time capabilities, a modern user interface, and a plugin architecture that allows for extensive customization.

NodeBB is known for its modern feature set, making it an excellent choice for building vibrant online communities, support forums, and discussion boards.

## Why use NodeBB with FerretDB?

NodeBB officially supports MongoDB as one of its primary database options.
Given that FerretDB is designed to be a true open source alternative to MongoDB, it can serve as a drop-in replacement for NodeBB's database backend.
This powerful combination offers several compelling advantages:

- **Open-source:** Both NodeBB and FerretDB are open-source projects, providing transparency, flexibility, and strong community backing, aligning perfectly with an open-source ethos.
- **PostgreSQL reliability for forum data:** By using FerretDB as NodeBB's backend, your crucial forum data (posts, users, topics) benefits from PostgreSQL's battle-tested reliability, and advanced data management capabilities.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating NodeBB with FerretDB can streamline your database management and reduce operational overhead.
- **Performance and scalability:** Leverage the performance characteristics of NodeBB's real-time engine combined with FerretDB's efficient handling of document data on a PostgreSQL backend.
- **No vendor lock-in:** Enjoy the freedom of truly open-source solutions without concerns about proprietary licensing or vendor lock-in.

## Connecting NodeBB to FerretDB

Connecting NodeBB to your FerretDB instance is straightforward, as NodeBB expects a MongoDB-compatible database.
Here's a step-by-step guide to get you started:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
   A typical FerretDB connection string looks like this:

   ```text
   mongodb://<username>:<password>@localhost:27017/
   ```

   Using a MongoDB client like `mongosh`, connect to your FerretDB instance to create a `nodebb` user within a `nodebb` collection:

   ```js
   db.createUser({
     user: 'nodebb',
     pwd: '<password>',
     roles: []
   })
   ```

   Replace `<password>` with a secure password of your choice.
   This user will be used by NodeBB to connect to FerretDB.

2. **Install NodeBB:** Follow the official NodeBB installation guide to set up your NodeBB instance.
   This typically involves cloning the repository and running `./nodebb setup` from the NodeBB root directory.
   See the [NodeBB Installation Documentation](https://docs.nodebb.org/installing/os/) for more.
3. **Configure and setup NodeBB's database connection:** During the setup process, you will be asked to choose a database type; ensure to select `mongo`.
   You will prompted to enter the connection string to your FerretDB instance; use the `nodebb` user and password credential you created.
   It should look like this:

   ```text
   mongodb://nodebb:<password>@localhost:27017/nodebb
   ```

   NodeBB also requires an admin user to be created during the setup process.
   Do this by providing the username, email, and password credentials when prompted.
   You will need the admin credentials to log in to the admin panel of your NodeBB forum.

4. **Launch NodeBB and test:** Once configured, launch your NodeBB application by running `./nodebb start` from the NodeBB root directory.
5. **Access your NodeBB forum:**
   After starting NodeBB, it will initialize and connect to FerretDB, setting up the necessary collections, and allow you to access your forum.
   You can then access your NodeBB forum at `http://localhost:4567` (or the port you specified during setup).
   Log in with the admin credentials you created earlier to access.

You can now proceed with setting up your forum, creating users, topics, and posts, with all data seamlessly stored in FerretDB.

## Example of NodeBB forum with FerretDB as the database backend

Here's a screenshot of the admin dashboard of the NodeBB forum running with FerretDB as the backend:

![A screenshot of the NodeBB admin dashboard with FerretDB as the backend](/img/blog/nodebb-admin-dashboard.png)

Below is an example of a topic created in the NodeBB forum, showcasing the real-time capabilities and modern UI that NodeBB provides:

![A screenshot of a NodeBB topic created with FerretDB as the backend](/img/blog/nodebb-topic.png)

This data is stored in FerretDB and can be seen via the MongoDB shell or any compatible GUI tool.

Using `mongosh`, connect to the FerretDB instance with the FerretDB connection string.
Explore the `nodebb` database to see the collections and documents created by NodeBB:

```text
nodebb> show collections
objects
sessions
```

Below we use Compass to connect to FerretDB and view the NodeBB data:

![A screenshot of NodeBB data in FerretDB using Compass](/img/blog/nodebb-compass.png)

This output demonstrates that NodeBB successfully writes and reads its document-based data in FerretDB, which in turn stores it efficiently in PostgreSQL.

## Conclusion

The integration of NodeBB and FerretDB offers a robust, scalable, and fully open-source solution for building and managing modern online communities.
By leveraging FerretDB, you can seamlessly run your stack completely in open source, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
