---
slug: seamlessly-manage-ferretdb-mongo-express
title: 'Seamlessly Manage Your FerretDB Databases with Mongo Express'
authors: [alex]
description: >
  Learn how to easily connect and manage your FerretDB databases using mongo-express, a web-based MongoDB admin interface.
image: /img/blog/ferretdb-mongoexpress.jpg
keywords: [mongodb gui, mongodb gui tools, open source mongodb gui]
tags: [mongodb gui, compatible applications, tutorial]
---

![Seamlessly Manage Your FerretDB Databases with Mongo Express](/img/blog/ferretdb-mongoexpress.jpg)

At FerretDB, we're committed to providing a truly open-source alternative to MongoDB, allowing you to leverage the power of PostgreSQL with the flexibility of a document database.

<!--truncate-->

A key part of that commitment is ensuring seamless integration with the tools you already know and love.

In this guide, we're excited to highlight how effortlessly you can manage your [FerretDB](https://www.ferretdb.com/) databases using Mongo Express, a popular web-based MongoDB admin interface.

## What is Mongo Express?

[Mongo Express](https://github.com/mongo-express/mongo-express) is a web-based administrative interface.
It offers a user-friendly way to connect and interact with multiple databases, browse collections, execute queries, and manage documents directly from your web browser.

Tools like Mongo Express that are designed for MongoDB work out-of-the-box with FerretDB.
This means you can:

- **Utilize a familiar interface**: If you're already accustomed to Mongo Express, you don't need to learn a new tool to manage your FerretDB instances.
- **Simplify database administration**: Perform common database operations visually, without needing to write complex commands.
- **Gain quick insights**: Easily view your data structure and content for development and debugging.

## Connecting Mongo Express to FerretDB

Connecting Mongo Express to your FerretDB instance is straightforward.
Here's what you need to do:

1. **Ensure FerretDB is running**: Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
2. **Install/run Mongo Express**: Follow the [official documentation for Mongo Express](https://github.com/mongo-express/mongo-express) to install and run it.
   This often involves a simple `npm install` or running a Docker container.
3. **Configure the connection**: When prompted for connection details in Mongo Express, specify the `ME_CONFIG_MONGODB_URL` environment variable or fill in the connection form with the following details:
   - Host/IP Address: This should point to your FerretDB instance.
     If you're running FerretDB locally, you can use `localhost` or `127.0.0.1`.
   - Port: Typically, FerretDB runs on port `27017`, but you can adjust this if you've configured it differently.
   - Database Name: (Optional) You can specify a database here, or browse all databases after connecting.
   - Username/Password: (Optional) If you have [authentication enabled on your FerretDB instance](https://docs.ferretdb.io/security/authentication/), provide the credentials.

     Example for Mongo Express using environment variables set in a `docker-compose.yml` file under the Mongo Express service:

     ```yaml
     ME_CONFIG_MONGODB_URL: mongodb://<username>:<password>@<host>:<port>/
     ME_CONFIG_BASICAUTH_USERNAME: <admin-username>
     ME_CONFIG_BASICAUTH_PASSWORD: <admin-password>
     ME_CONFIG_MONGODB_ENABLE_ADMIN: <true|false>
     ```

     Ensure to adjust `admin-username` and `admin-password` as needed, or omit if not using basic auth for Mongo Express itself.

4. **Connect and explore**: Once configured, Mongo Express will connect to your FerretDB instance, and you can start managing your databases and collections!

### Example of Mongo Express with FerretDB

Here is a view of a `books` collection in FerretDB using Mongo Express:

![Mongo Express view of a collection in FerretDB](/img/blog/mongoexpress-ferretdb-collection.png)

Next, you can perform operations on the data from within Mongo Express, such as querying the `books` collection:

![Mongo Express query example of a collection in FerretDB](/img/blog/mongoexpress-query.png)

## Conclusion

Integrating Mongo Express with FerretDB empowers you with a familiar and efficient tool for managing your document databases.
This is just one example of how FerretDB allows you to continue using your preferred ecosystem in open source while benefiting from the robust and reliable backend of PostgreSQL.

We're constantly working to expand our integrations and support for various tools.
Stay tuned for more updates, and feel free to reach out to us if you have any questions or suggestions!

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
