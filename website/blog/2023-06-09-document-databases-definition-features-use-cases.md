---
slug: document-databases-definition-features-use-cases
title: 'Document Database: Definition, Features, Use Cases'
authors: [alex]
image: /img/blog/document-databases.jpg
description: >
  Learn about NoSQL document databases, their features, benefits, use cases, and popular examples such as FerretDB, MongoDB, and Couchbase.
keywords: [
    document databases
    NoSQL databases
    schema design
  ]
tags: [document databases]
---

![Document databases](/img/blog/document-databases.jpg)

In this blog post, we will discuss document databases, diving into their history, features, benefits, use cases, and some examples like FerretDB, MongoDB, and Couchbase.

<!--truncate-->

Since the inception of databases in the 1960s, the world of data has undergone significant changes, going from relational databases, which were the norm for decades, to the emergence of NoSQL databases in the 2000s.
A constant during this period has been the need for databases to handle large volumes of data and offer fast data retrieval in a secure and consistent manner.

While relational databases like Oracle, MySQL, and PostgreSQL have long been the stalwarts of the database world, there was rising concern on their capabilities towards handling unstructured data without a fixed schema and the need to pigeonhole data into tables and rows.

In the last two decades, there's been a growing number of applications with huge volumes of unstructured data that would normally be challenging to store and process using traditional relational databases.
This concern, among others, gave rise to the NoSQL and document database movement.

In this article, you will learn about document databases, their unique benefits, use cases, and examples that have made it quite popular among developers.

## What is a Document Database?

Document databases – or document-oriented databases – are a type of NoSQL database that stores data as JSON-like documents instead of rows, columns, and tables commonly associated with traditional SQL databases.

In relational databases, every data record goes into a table-column-row format that ideally requires a fixed schema beforehand; this was not the case with document databases.

So imagine you are to collate data on a number of published books; one book contains information on the author, title, and number of page, while another adds more information with the publisher, genre, and ISBN.

When modeling these records a relational database, you would have to create a table for these books, and the table would have to contain the same columns, even if some of the columns are empty.

```js
| id | title        | author      | number_of_pages | publisher  | genre | isbn   |
|----|--------------|-------------|-----------------|------------|-------|--------|
| 1  | Book Title 1 | Author 1    | 200             | NULL       | NULL  | NULL   |
| 2  | Book Title 2 | Author 2    | 300             | Publisher 2| Genre 2| ISBN-2 |
```

Situations like this are where document databases truly thrive.
Instead of having empty columns, these two books can be stored as separate documents with each containing all the necessary information for that particular book - no fixed schema or structure.

First document:

```js
{
    "_id": "uniqueId1",
    "title": "Book Title 1",
    "author": "Author 1",
    "number_of_pages": 200
}
```

Second document:

```js
{
    "_id": "uniqueId2",
    "title": "Book Title 2",
    "author": "Author 2",
    "number_of_pages": 300,
    "publisher": "Publisher 2",
    "genre": "Genre 2",
    "isbn": "ISBN-2"
}
```

### Understanding Documents Data Model

Documents are at the heart of everything in a document database.
Documents are akin to real-life blank "documents" where you can enter as much information as possible for that particular record.
And just like you have related real-life documents in a drawer, related documents in a document database are stored in collections.

A document in a collection represents a single record of information about an object and any associated metadata, all stored as key-value pairs of data, containing various data types such as numbers, strings, objects, arrays, etc.
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

Data types are another interesting thing to note here; there's no fixed schema so you can model the fields in any data type necessary for that data record.
In the example above, the `genre` field currently stored as an array `["High Fantasy", "Adventure"]` can also be updated or available in another document record as a single string object `"High Fantasy"`, or even an object..
For example, the `genre` currently stored as a can be stored `["High Fantasy", "Adventure"]` can also be updated or available in another document record as a single string `"High Fantasy"`, or even an object.

This flexibility - often unavailable in relational databases - makes document databases suitable for semi-structured and flexible data that allows them to adapt to a company's evolving needs.

This structure also makes data retrieval easier and faster; instead of sifting through multiple records in different tables using complex joins, which are quite resource and time-intensive, you can query a single document and get all the information you need in a single query.

## Benefits of Document Databases

While most of these are already covered, let's look at some of the advantages of using a document database:

### Flexible Database Schema

Many applications today do not have a defined data format or structure since new types of information are constantly being added in different data types; emails, social media posts, customer reviews are all examples that show the necessity of a flexible of flexible schema.
In these cases, each data record may have different elements such as text, images, hashtags, location data, emojis, etc.

Document databases are incredibly flexible and can are able to accommodate these kind of data, and their unusual nature.
Each document is a separate entity with its own structure; there's no need to predefine a schema or structure.
All the documents in a collection don't have to have the same fields, or even data types.
Even at that, the documents can be updated to accommodate new fields and data types without affecting other documents in the collection.

While in relational databases, you often end up with many null values for optional columns, fields that don't have a value simply do not need to be included in the document.
You can even have complex data structures with nested objects and arrays.

Such flexibility is truly unheard of in relational databases.

### High Scalability

As your applications grow with higher traffic load, scalability becomes an important factor to consider since your original set up and resources – CPU, RAM, hard disk etc. – may not be able to handle the increased load.

One significant advantage of document databases over traditional relational databases is their ability to scale horizontally (also know as "sharding"), which is the ability to add more servers (nodes) to your database cluster to handle increased traffic and storage needs.
This option, in contrast to vertical scaling, is more cost-effective and offers better performance.

Both relational and non-relational databases have the option to scale vertically where you increase the computational resources based on your needs.
However, often times, the performance and costs of vertical scaling do not scale linearly - you might reach a point of diminishing returns where more resources do not necessary lead to an equal increase in performance.
In such cases, you might need to scale horizontally by adding more servers to your database cluster.
Moreover, even though it's possible, it's quite challenging and complex to scale horizontally in relational databases due to the presence of multiple related data across nodes.

Horizontal scaling in document databases makes them more fault-tolerant and highly available; even when some nodes fail, the system can remain operational with no single-point of failure.
Plus they enable low latency for applications that are globally distributed.

### Performance

The two previous benefits mentioned above (flexible schema and high scalability) culminates in document databases being highly-performant, particularly when working with nested objects and documents; you can easily query and update nested objects in a single atomic operation.
Applications where this can be a huge advantage include content management systems, social media apps, real-time analytics, IoT applications, and any use case where you need to handle numerous data types and structures.

With the possibility of horizontal scaling, document databases can handle large amounts of data and high traffic loads by just spreading them across multiple distributed nodes.
And since related object data are stored in a single document and no need for complex JOIN operations, along with the chance to create indexes for any field - even in a nested object - data retrieval is so much faster.

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
| Development | Developer-friendly, with a more natural data model for object-oriented programming languages                           | Less user-friendly data model, more complex SQL queries required                        |
| Consistency | Lower consistency guarantees, with more focus on performance and availability                                          | Higher consistency guarantees, with more focus on data integrity and accuracy           |
| Use Cases   | Best for handling large volumes of semi-structured or unstructured data, suited for modern web and mobile applications | Best for handling structured data, suited for traditional business applications         |

**Read more:** [PostgreSQL vs MongoDB - Understanding a Relational Database vs Document Database](https://blog.ferretdb.io/mongodb-vs-postgresql-database-comparison/)

## Examples of NoSQL Document Databases

### FerretDB

[FerretDB](https://www.ferretdb.io/) is an open source document database alternative to MongoDB with PostgreSQL as the backend.
This database was borne out of a need to offer a truly open-source alternative to MongoDB after it's switch to SSPL back in 2018.

While it's relatively new to the scene with its first GA release in 2023, FerretDB is already gaining traction and being leveraged by users seeking freedom away from the vendor lock-in associated with MongoDB.

With MongoDB compatibility built-in, FerretDB converts MongoDB wire protocols to SQL in PostgreSQL, allowing you to run MongoDB workloads on PostgreSQL.
It translates documents using MongoDB's BSON format to JSONB in PostgreSQL (preserving the order and data types of the document field) through its own mapping system called PJSON (Learn more about this [in this blog post](https://blog.ferretdb.io/pjson-how-to-store-bson-in-jsonb/)).

Users are able to leverage similar syntax and query language as MongoDB, so a insert statement and query in FerretDB looks like this:

<!-- write an insert statemnt for that document below -->

```js
db.users.insert({
  name: 'John Doe',
  age: 25,
  address: {
    street: '123 Main Street',
    city: 'New York',
    state: 'NY',
    zip: '10001'
  }
})
```

```js
db.users.find({ 'address.state': 'NY' })
```

OUTPUT:

```js
{
  "_id": ObjectId("5f5b9e2e8b0c0f0001a2b2c3"),
  "name": "John Doe",
  "age": 25,
  "address": {
    "street": "123 Main Street",
    "city": "New York",
    "state": "NY",
    "zip": "10001"
  }
}
```

It looks like MongoDB, doesn't it?
But it's actually FerretDB using PostgreSQL under the hood.

Besides, experienced PostgreSQL users can manage FerretDB using all the extensions and administrative features already available in PostgreSQL, such as replication, backup, and monitoring, while still enjoying the flexibility and ease-of-use associated with MongoDB.
You can also [use FerretDB with familiar MongoDB GUI applications](https://blog.ferretdb.io/using-ferretdb-with-studio-3t/) like [Studio3T](https://studio3t.com/), [Mingo](https://mingo.io/), NoSQLBooster, and more.

In terms of performance, while FerretDB's primary focus is to enable more compatibility with MongoDB, it's also working on improving performance by pushing more queries to the backend.
[Read more about this here](https://blog.ferretdb.io/ferretdb-fetches-data-query-pushdown/).

In addition to PostgreSQL, FerretDB is also building support for other database backends, like Tigris ([beta level](https://www.tigrisdata.com/docs/concepts/mongodb-compatibility/)), [basic experimental support for SQLite (not available yet)](https://blog.ferretdb.io/ferretdb-v-1-2-0-minor-release/), and more on the way.

### MongoDB

Among all document databases, [MongoDB](https://www.mongodb.com/) is indisputably the most popular - and by a wide margin.
In particular, it's rich query language, complex query support, aggregation framework, secondary indexes, multi-language support, and all-round ease-of-use play a favorable role in its popularity.
Another contributing factor is its inclusion in popular JavaScript stacks like MEAN and MERN, which is widely used for web application development.

Besides its popularity, MongoDB has been highly influential in the increased use of document databases today.
With drivers available for a large number of programming languages, MongoDB makes it possible for developers to work in their preferred programming language.
You also have access to a strong ecosystem of tools and services, including MongoDB Atlas (fully managed cloud database service), MongoDB Compass (database GUI), and many more.

MongoDB started as an open-source project, a factor that helped it gain initial adoption.
However, it has since made away with its open-source license for a more proprietary and controversial SSPL license; this move was a [motivation behind the creation of FerretDB](https://blog.ferretdb.io/mangodb-overwhelming-enthusiasm-for-truly-open-source-mongodb-replacement/), an open-source alternative to MongoDB.

Designed to be horizontally scalable, MongoDB is built to handle large volumes of data and traffic, sharding data across multiple nodes.
It also enables high availability through replica sets that provide redundancy and failover; if a primary node fails, a new primary is elected from the remaining secondary nodes.

In the early days, one of the biggest concerns with MongoDB was its lack of support for ACID (Atomicity, Consistency, Isolation, and Durability) transactions.

Ensuring consistency across multiple nodes is a challenge for distributed databases, and MongoDB was no exception.
There's an interesting theorem on this called the CAP theorem, which states that it's impossible for a distributed database to provide more than two of these three guarantees: Consistency, Availability, and Partition tolerance.

Many document database settle for an option between being choosing consistency (CP) - all nodes have the same data and remain consistent - or availability (AP), where all nodes can answer queries, even with stale data.
In MongoDB's case, they chose to sacrifice consistency for availability and partition tolerance, which was why it didn't support ACID transactions in the early days.

Since MongoDB version 4.0, MongoDB has added support for multi-document ACID transactions; support for distributed multi-document ACID transactions was added in version 4.2.

### Couchbase

[Couchbase](https://www.couchbase.com/) is an open-source NoSQL distributed multi-model database renowned built and optimized for interactive applications.
Similar to MongoDB, Couchbase uses a flexible JSON model which doesn't require a fixed data model and can be modified on the fly.
That's where the similarity ends however.
Couchbase offers its own unique query language called N1QL (pronounced "nickel"), a sort of SQL-like query language for JSON.

A typical query in Couchbase looks like this:

```sql
SELECT title, author
FROM `bucket`
WHERE genre = "Novel" AND publication_date BETWEEN "1850-01-01" AND "1860-12-31"
```

As you can see, it's more akin to SQL than MongoDB's query language, and allows a lot of the same operations you would do in SQL, including joining, filtering, aggregating, ordering, and more.
In Couchbase, `bucket` represents the name of your Couchbase bucket, an analogous term for "database".

Using a distributed architecture with sharding and load balancing, Couchbase is built for easy scalability, replication, and failover.
It also provides distributed ACID transactions across multiple documents, buckets, or nodes.

### RavenDB

[RavenDB](https://ravendb.net/) is a NOSQL document database that is fully transactional (ACID) across the database and across clusters, perfectly suited for complex, semi-structured, and hierarchical data in ML/AI models.
Asides being one of the first NoSQL databases to support ACID transactions, RavenDB is a multi-model database that supports document, relational, graph, and key-value data models.

Like other document databases (and unlike relational databases), RavenDB uses a jSON-like flexible data model that doesn't require a fixed data schema.

Using its own SQL-like query language called RavenDB Query Language (RQL), RavenDB supports a wide range of queries, including full-text search, spatial queries, and more.
RavenDB also supports LINQ queries, which is a popular query language for .NET developers.

An example query might look like this:

```sql
from Employees
where hiredAt > '2000-01-01T00:00:00.0000000'
```

Interestingly, queries in RavenDB always uses indexes with no support for full scans.
This is because RavenDB is designed to be fast, and full scans are not suitable for large datasets.

When you run a query in RavenDB, the query optimizer searches for an existing index to satisfy the query and if it doesn't find one, a new index is created.
While this approach may result in many indexes which could potentially have adverse effects, RavenDB's query optimizer tries to mitigate this by modifying an existing index for the new query, when possible.

In addition to indexing, RavenDB uses caching and batching to optimize server and network resources, along with provisions for multi-master replication, sharding, and replication to enhance availability and scalability.
RavenDB also uses a built-in Lucene-based full text search – a top-tier highly customizable, fully featured, and near real-time search engine.

As an all-in-one database, RavenDB also includes a cloud service, a Time Series model, ML processing, and an online analytical processing (OLAP) plugin for business analysis.

### Firebase

[Firebase](https://firebase.google.com/) is a fully managed cloud-based NoSQL document database provided by Google as part of its Google Cloud Platform (GCP) services.
It's a serverless database that is fully managed by Google, so you don't have to worry about provisioning, scaling, or managing your database.

Interacting with Firebase is done through the Firebase SDK, which is available for a wide range of platforms, including web, Android, iOS, and Unity.
Firebase also provides an API for different programming languages, including JavaScript, Node.js, Java, Python, and Go.
However, while they support the same set of features, the syntax and usage vary slightly depending on the language.

For instance, a typical query in JavaScript looks like this:

```js
const db = firebase.firestore()

db.collection('users')
  .where('address.state', '==', 'NY')
  .get()
  .then((querySnapshot) => {
    querySnapshot.forEach((doc) => {
      console.log(doc.id, ' => ', doc.data())
    })
  })
  .catch((error) => {
    console.log('Error getting documents: ', error)
  })
```

## Getting Started with Document Databases

Document databases represent a significant departure from traditional relational databases in storing and accessing data.
This evolution is interesting for developers looking to build their applications with a flexible data model that offers high scalability and agility.

If you want to be part of a growing community of NoSQL document database enthusiasts, the [Document Database Community](https://documentdatabase.org/) is a global network of developers where you can learn more about recent trends, technologies, and news in the document database space.

If you're looking for a document database to practice and build with, Ferret is a good option for you.
It's open source nature and compatibility with MongoDB wire protocols and queries makes it an attractive option.

To get started, checkout the [installation guide](https://docs.ferretdb.io/quickstart-guide/) for FerretDB.
