---
title: "Document Database: Definition, Features, Use Cases"
authors: [alex]
image: /img/blog/mongodb-alternatives.png
description: >
    Learn about NoSQL document databases, their features, benefits, use cases, and popular examples such as FerretDB, MongoDB, and Couchbase.
unlisted: true
---

![To be replaced](/img/blog/ferretdb-v0.9.1.jpg)

Since the inception of databases in the 1960s, the world of data has undergone significant changes.
While relational databases like Oracle, MySQL, and PostgreSQL have long been the stalwarts of the database world, document databases have been gaining momentum in recent years.

<!--truncate-->

The use of a Structured Query Language (SQL) in relational databases to store and retrieve table-based data records are not always sufficient for handling records with unstructured data and heavy workloads, particularly for organizations with scalability concerns.

To address these needs, NoSQL databases emerged as a solution for companies looking to manage unstructured data and enable faster processing.
One of these NoSQL databases is the document database.
There are other types like the key-value store, graph database, and wide-column store.

In this article, we will focus on what document databases are, their unique features, and use cases.

## What is a Document Database?

Document databases – or document-oriented databases – are a type of NoSQL database that stores data as JSON-like documents instead of rows, columns, and tables commonly associated with traditional SQL databases.

Document databases appeal to most developers due to their similarities to objects in programming techniques.
Their flexibility makes them suitable for semi-structured and flexible data that allows them to adapt to a company's evolving needs.

## Understanding Documents Data Model

Documents are at the heart of everything in a document database.
They represent a single record of information about an object and any associated metadata, all stored as key-value pairs of data, containing various data types such as numbers, strings, objects, arrays, etc.
These documents can be stored in various formats, such as JSON, BSON, YAML, or XML.

For instance, the following is a typical example of a document containing information on a book:

```js
{
 title: "The Lord of the Rings",
 author: {
   name: "J.R.R. Tolkien",
   nationality: "British",
 },
 publication_date: "July 29, 1954",
 publisher: "George Allen & Unwin",
 genre: ["High Fantasy", "Adventure"],
 isbn: "978-0618640157",
 number_of_pages: 1178,
 has_movie_adaptation: true,
 movie_adaptation: {
   director: "Peter Jackson",
   release_date: "December 19, 2001",
   awards: ["Best Picture", "Best Director", "Best Adapted Screenplay"],
   box_office: "$1.19 billion"
 },
}
```

In the example, we can see how different types of document data are used to represent various aspects of the book, such as its author, publication information, genre, and movie adaptation.
This flexibility and expressiveness make document databases a popular choice for storing and managing data in modern applications.
Moreover, NoSQL databases usually don't need a fixed schema, so they can accommodate new data types without requiring extensive modification.

## Benefits of Document Databases

The following are advantages of the document database:

### Flexible Database Schema

NoSQL databases are incredibly flexible compared to relational databases.
Each document is a separate entity with its own structure, making it easier to query data and make changes that enable the application to evolve over time.
There is no need to modify all existing records to accommodate a new structure.

These flexibility provides a faster and more flexible way to manage and create databases.
Developers can start building a database with collections of documents immediately, and they can modify it as they build their applications.
With collections of documents that can be populated right away, you don’t need to know the structure of the database.
Instead, developers can get started with their database as soon as possible.

### High Scalability

By leveraging a document database, you can enable low-cost horizontal scaling, as opposed to vertical scaling, which poses inherent limitations.
Because document databases collocate related data in a single document, sharding datasets is a common strategy for minimizing host-to-host coordination.

Document databases emphasize performance and availability, letting you balance consistency with performance, contrary to relational databases, where consistency is always the top priority.

### Performance

Document databases offer superior and faster performance for read and write operations, particularly when working with nested objects and documents.
As shown in the document data model example above, you can easily query and update nested objects in a single atomic operation.
By storing related data together, document databases reduce the need for complex joins and queries, resulting in faster access to the required information.

## Document Databases vs. Relational Databases

In a relational database, data is structured in separate tables defined by the programmer so that the same object appears in multiple tables.
To get the desired result from the database, you must use join statements.

On the other hand, you can use a document database to store data for all the information about an object as a single database instance, although each object may differ significantly from the others.

Here is a comparison table between document databases and relational databases:

| Feature     | Document Databases                                                                                                     | Relational Databases                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Structure   | Flexible schema design, documents can have different structures within the same collection                             | Predefined, rigid structure with tables, columns, and rows                              |
| Scalability | Efficient horizontal scaling, allowing for easy distribution of data across multiple hosts                             | Vertical scaling, adding resources to a single machine to handle more data              |
| Querying    | Rich querying capabilities, with support for complex nested data structures                                            | Limited support for querying nested data structures, with more focus on join operations |
| Development | Developer-friendly, with a more natural data model for object-oriented programming languages                         | Less user-friendly data model, more complex SQL queries required                            |
| Consistency | Lower consistency guarantees, with more focus on performance and availability                                          | Higher consistency guarantees, with more focus on data integrity and accuracy           |
| Use Cases   | Best for handling large volumes of semi-structured or unstructured data, suited for modern web and mobile applications | Best for handling structured data, suited for traditional business applications         |

**Read more:** [PostgreSQL vs MongoDB - Understanding a Relational Database vs Document Database](https://blog.ferretdb.io/mongodb-vs-postgresql-database-comparison/)

## Examples of NoSQL Document Databases

### FerretDB

FerretDB is an open-source document database with MongoDB-compatibility built-in, allowing you to run MongoDB workloads on other database backends, such as PostgreSQL, Tigris, SAP Hana, and many others.
With the release of the production-ready FerretDB 1.0 GA as well as a growing open source community, you can now build your applications with FerretDB while leveraging all the features of a PostgreSQL backend, or any other supported database backend.

### MongoDB

MongoDB is a proprietary documented-oriented database built to be highly scalable and performant in handling high-volume data storage.
Being one of the most popular document databases on the market, MongoDB has redefined the space with its range of drivers and APIs for various frameworks and programming languages.

### Couchbase

Couchbase is an open-source popular document database renowned for its scalability, performance, and speed.
With a flexible model based on JSON documents, Couchbase enables effortless data modeling and schema changes.
As a multi-model NoSQL database, Couchbase is tailored towards highly interactive applications requiring low latency and high throughput.

### RavenDB

RavenDB is a NoSQL document database with several features and capabilities suitable for mobile and web applications.
Built with ACID compliance, RavenDB is fully transactional within the database and across clusters.
RavenDB is an all-in-one database that includes a cloud service, a Time Series model, ML processing, and an online analytical processing (OLAP) plugin for business analysis.

## Getting Started with Document Databases

Document databases represent a significant departure from traditional relational databases in storing and accessing data.
This evolution is interesting for developers who can build their applications with a flexible data model that offers high scalability and agility.

If you want to be part of a growing community of NoSQL document database enthusiasts, the [Document Database Community](https://documentdatabase.org/) is a global network of developers where you can learn more about recent trends, technologies, and news in the document database space.

If you're looking for a document database to practice and build with, Ferret is a good option for you.
It's open source nature and compatibility with MongoDB wire protocols and queries makes it an attractive option.

To get started, checkout the [installation guide](https://docs.ferretdb.io/quickstart-guide/) for FerretDB.
