---
slug: easily-deploy-managed-ferretdb-elestio
title: 'Easily Deploy Managed FerretDB on Elestio'
authors: [alex]
description: >
  Learn how to deploy a fully managed FerretDB instance in production in any cloud environment using Elestio.
image: /img/blog/ferretdb-elestio.jpg
tags: [tutorial, open source, cloud]
---

![Easily deploy FerretDB on Elestio](/img/blog/ferretdb-elestio.jpg)

Do you want to deploy a fully managed FerretDB instance in production in any cloud environment?

<!--truncate-->

[Elestio](https://elest.io/) lets you easily run, monitor, backup, maintain, and secure your [FerretDB](https://www.ferretdb.com/) instance.
It is a DevOps platform that lets you manage and deploy open-source software in production environments.

You can deploy your service on any cloud ([AWS](https://aws.amazon.com/), [DigitalOcean](https://www.digitalocean.com/), [Hetzner](https://www.hetzner.com/), etc.) or on-premise.

In this blog post, you will learn to deploy FerretDB on Elestio in any cloud environment under 5 minutes.

## Prerequisites

- [Elestio account](https://elest.io/)
- `mongosh`

## How to deploy FerretDB on Elastio

FerretDB is an open source document database alternative to MongoDB with Postgres as a backend.
To start creating a FerretDB service on Elestio, [simply follow this link](https://elest.io/open-source/ferretdb).

### Start by creating a FerretDB service

Select "FerretDB" service.

![Select FerretDB service](/img/blog/ferretdb-elestio/select-service.png)

### Select cloud provider

Next, select a service cloud provider to use for your project.
There are different options – DigitalOcean, Hetzner, Amazon, Amazon Linode, Vultr, Scaleway, and BYOS if you prefer.

For this example, let's set up FerretDB on DigitalOcean.

![Set up service cloud provider](/img/blog/ferretdb-elestio/cloud-provider.png)

You can also select the "Service Cloud Region" and "Service Plan" for the instance.

### Select support and advanced configuration

On the next page, select the kind of technical support you want.
For example, length of remote backup retention, service snapshots, response time, SLA, priority queuing, etc.

![Set up support & advanced configuration](/img/blog/ferretdb-elestio/service-dashboard.png)

Once you're done, create the service.

It may take a few minutes to provision the instance and resources.

That's all you need to set up FerretDB using Elastio!

### Connect to FerretDB using `mongosh`

To connect to the database, you need the FerretDB connection string for your instance.
Select "Display DB Credentials" to get the connection string.

![FerretDB service dashboard](/img/blog/ferretdb-elestio/service-dashboard.png)

Connect to your FerretDB instance via mongosh in the following format:

```sh
mongosh 'mongodb://username:password@host-address/ferretdb?authMechanism=PLAIN'
```

And that connects you to the FerretDB instance!

## Run basic CRUD operations on FerretDB instance

You can now populate the FerretDB instance with data.

Start by inserting the following document into a `record` collection.

```json5
db.record.insertOne({
    username: "JD",
    content: "Enjoying the beautiful weather today! 🌞 #sunnyday",
    likes: 120,
    timestamp: new Date()
});
```

Once it's inserted, view the documents by running `db.record.find()`:

The output:

```json
[
  {
    _id: ObjectId('66d6a9346e70f5ffc91022c0'),
    username: 'JD',
    content: 'Enjoying the beautiful weather today! 🌞 #sunnyday',
    likes: 120,
    timestamp: ISODate('2024-09-03T06:14:12.634Z')
  }
]
```

Next, update the `likes` of JD's post to 150.

```json5
db.record.updateOne(
    { username: "JD" },
    { $set: { likes: 150 } }
);
```

The output:

```json
{
  "acknowledged": true,
  "insertedId": null,
  "matchedCount": 1,
  "modifiedCount": 1,
  "upsertedCount": 0
}
```

You can run `db.record.find()` again just to be sure it's updated.

Finally, delete the singular document from the collection.

```json5
db.record.deleteOne({ username: "JD"});
```

## Conclusion

Like that, you can have a managed FerretDB database production-ready on Elestio.
No need to worry about DevOps or infrastructure concerns!
Plus, it's open source, with no vendor lock-in, so you can easily migrate your data to any cloud anytime.

If you want to know more about FerretDB, do check out:

- [FerretDB documentation](https://docs.ferretdb.io/)
- [Community channels](https://docs.ferretdb.io/#community)

[Get started with managed FerretDB on Elestio](https://elest.io/open-source/ferretdb)
