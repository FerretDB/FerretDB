---
slug: ferretdb-meteor-mongodb-alternative
title: 'How to Spin up FerretDB database in Minutes'
authors: [alex]
description: >
Want to find our how FerretDB? Take a look at our demo to learn more.
image: /img/blog/ferretdb-meteor.jpg
keywords: [FerretDB, meteor, mongodb alternative]
tags: [installation guide]
---


# FerretDB Demo: Launch a FerretDB Database on Civo Cloud in Minutes

Want to know what a [truly open-source MongoDB alternative](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/) looks like? No worries! We’ve got you covered.

We’ve set up the [FerretDB](https://www.ferretdb.io/) demo page so you can experiment and see what FerretDB looks like in action: creating your own database, performing CRUD operations, running indexes, or aggregation pipeline commands, basically all the commands you’re already used to.

If you’re new to FerretDB and wondering what it’s all about, FerretDB is an open-source database committed to restoring MongoDB workloads to their open-source origins while enabling developers to use the same syntax and commands. Basically, we convert MongoDB wire protocols to SQL with PostgreSQL as the backend engine.

Now that we know what FerretDB is all about, let’s jump into the demo and try it out.

Let’s get started.


## How to Get Started with Running a FerretDB Demo

In this blog post, we’ll be showcasing the FerretDB demo page and what you can achieve with it.

The FerretDB demo is quick to set up, and you can get started through the demo site.



* Demo site

Once you are on the site, you can click the “launch a database for free” button to create your own FerretDB database at Civo.com - a cloud marketplace and service platform.

Note that all FerretDB databases created in this demo are deployed on a managed Kubernetes cluster provided by Civo, and are only available for 2 hours after creation.

[Add Image on HomeScreen of Demo Page]()

This process involves getting the necessary credentials for the setup from Civo and then installing the latest FerretDB version on the Kubernetes cluster for you. This whole setup takes about 2-4 minutes to complete; once the database is ready, you’ll get a connection URI with your authentication credentials.

[Add Image on Connection Process]

Remember that the FerretDB instance is only available for 2 hours, so if you would like to go all the way and test out FerretDB for a longer period of time, you can install FerretDB through the Civo Marketplace or through any of our installation guides.


## Prerequisites

Before proceeding, please make sure you have:



* `mongosh`. Please [install mongosh](https://www.mongodb.com/docs/mongodb-shell/install/) if you don’t currently have it installed.


## How to Connect to FerretDB Using the Provided URI

Now that you have your FerretDB connection URI, you can connect to your FerretDB instance.

Copy the FerretDB connection URI; it typically follows this format:

```sh

mongodb://username:password@host:port/database?authMechanism=PLAIN

```

Connect to FerretDB through `mongosh` with the connection URI:

```sh

mongosh 'mongodb://username:password@host:port/database?authMechanism=PLAIN'

```

And that’s about it! You have your FerretDB database running.


## Test out FerretDB with Basic Examples

Now that you're connected to FerretDB, let's explore some basic database operations:

**Inserting data:** Let’s try inserting 4 documents into a database collection `demo`:

```js

db.demo.insertMany([

  {

    name: 'Chinedu Eze',

    age: 20,

    email: 'chinedu.eze@example.com',

    major: 'Computer Science',

    sports: ['Basketball', 'Running']

  },

  {

    name: 'Maria Rodriguez',

    age: 21,

    email: 'maria.rodriguez@example.com',

    major: 'Business Administration',

    sports: ['Yoga', 'Swimming']

  },

  {

    name: 'Kelly Li',

    age: 19,

    email: 'kelly.li@example.com',

    major: 'Engineering',

    sports: ['Soccer', 'Cycling']

  },

  {

    name: 'Sara Nguyen',

    age: 22,

    email: 'sara.nguyen@example.com',

    major: 'Biology',

    sports: ['Tennis', 'Dancing']

  }

]);

```

**Indexing:** In FerretDB, we can set an index on the `email` field the same way you would on MongoDB, like this:

```js

>db.demo.createIndex({ email: 1 });

email_1

```

**Delete data:** Let’s remove a document from the database using a `name` field as the specified query.

```js

> db.demo.deleteOne({ name: 'Sara Nguyen' });

{ acknowledged: true, deletedCount: 1 }

```

**Update data:** Let’s update the documents in the database with `age` less than or equal to `21`.

```js

db.demo.updateMany(

  { age: { $lt: 20 } },

  { $addToSet: { sports: "aerobics" } }

)

```

A `read` command should give you a good view of all the changes made to the collection: \
 \
```js

db.demo.find()

```

Output:

```json5

[

  {

    _id: ObjectId("648bc0619df6027209a65b40"),

    name: 'Chinedu Eze',

    age: 20,

    email: 'chinedu.eze@example.com',

    major: 'Computer Science',

    sports: [ 'Basketball', 'Running' ]

  },

  {

    _id: ObjectId("648bc0619df6027209a65b41"),

    name: 'Maria Rodriguez',

    age: 21,

    email: 'maria.rodriguez@example.com',

    major: 'Business Administration',

    sports: [ 'Yoga', 'Swimming' ]

  },

  {

    _id: ObjectId("648bc0619df6027209a65b42"),

    name: 'Kelly Li',

    age: 19,

    email: 'kelly.li@example.com',

    major: 'Engineering',

    sports: [ 'Soccer', 'Cycling' ]

  }

]

```

You can try several combinations of commands; here is a list of the [supported commands on FerretDB](https://docs.ferretdb.io/reference/supported-commands/).

Feel free to experiment and explore all the capabilities of FerretDB.


## Conclusion

With Civo's seamless setup process, you can have a FerretDB instance up and running in minutes, but note that the database cluster is only up for two hours before it’s removed, so you may need to create another one, or install [FerretDB through the Civo Marketplace](https://www.civo.com/marketplace/FerretDB) or through the [FerretDB installation guide](https://docs.ferretdb.io/quickstart-guide/).

Here’s a video demonstration showcasing the FerretDB demo.

[Add video here]

Remember, FerretDB is available on Civo Cloud, so you can take advantage of all the managed cloud operations available to you. You can also check out the FerretDB documentation for more details and insight on how to run FerretDB.