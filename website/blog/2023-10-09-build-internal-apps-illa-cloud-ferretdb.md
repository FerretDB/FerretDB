---
slug: build-internal-apps-illa-cloud-ferretdb
title: How to Build Internal Apps with Illa Cloud and FerretDB
authors: [alex]
description: >
  In this blog post, we will explore ways to quickly build internal apps in minutes using Illa Cloud and FerretDB.
image: /img/blog/ferretdb-illacloud.jpg
tags: [compatible applications, tutorial, open source]
---

![How to Build Internal Apps Dashboards with Illa Cloud and FerretDB](/img/blog/ferretdb-illacloud.jpg)

In this blog post, we will explore ways to quickly build internal apps in minutes using [Illa Cloud](https://www.illacloud.com/) and [FerretDB](https://www.ferretdb.io/).

<!--truncate-->

FerretDB is an [open-source document database replacement](https://blog.ferretdb.io/open-source-is-in-danger/) for MongoDB that converts MongoDB protocols to SQL and provides [PostgreSQL](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/) as database backend options.

By connecting FerretDB with Illa Cloud, you can [use the same MongoDB commands](https://blog.ferretdb.io/mongodb-crud-operations-with-ferretdb/) to run queries and create dashboards, admin panels, or internal tools.

Let's get started.

## Prerequisites

You can get started with Illa Cloud by setting your connection URI as the data source.
[See quickstart guide](https://docs.ferretdb.io/quickstart-guide/) for more on setting up FerretDB.

Before you proceed, you will need to have the following:

- FerretDB connection URI (e.g. `mongodb://username:password@127.0.0.1:27017/?authMechanism=PLAIN`)
- [Illa Cloud account](https://cloud.illacloud.com/)

## Creating FerretDB resource

We'll use Illa Builder to create a simple dashboard UI for the FerretDB intance.

![Illa Builder](/img/blog/ferretdb-illacloud/1-illabuilder.png)

An Illa resource is a stored configuration for connecting to a data source.
This resource exposes your data source so you can run queries on it through Illa.
Illa provides several resource type connections that you can use, including PostgreSQL, [Supabase](https://supabase.com/), and MongoDB.

As a true alternative to MongoDB, FerretDB works with most of the familiar MongoDB tools and applications.
So, to create a FerretDB resource for Illa Builder, connect to your FerretDB instance using the MongoDB resource type.

![Choose Resource type for connecting Illa to FerretDB](/img/blog/ferretdb-illacloud/2-mongodbresource.png)

Once it opens the connection window for the resource, set the name of your resource and the connection string URI.

![Set the connection string](/img/blog/ferretdb-illacloud/3-connection-uri.png)

![Resource view](/img/blog/ferretdb-illacloud/4-resource-view.png)

## Creating actions in Illa

Actions in Illa represent operations to perform on the resources, such as CRUD or aggregation operations.
They also connect your data operations to components in Illa.

From the Illa Builder, we will create components to interact with our FerretDB instance.
Illa provides "Action type", "Collection", and "Transformer" methods when creating actions where the Action type represent commands, such as `bulkwrite`, `find`, `findOne`, `aggregate`, `deleteMany`, `deleteOne`, `listCollections`, and `count`.

### 1. Query all data using `find`

Let's an create action for retrieving all the data in the `supply` collection of the FerretDB instance.

![Actions to display for query all on supply collection](/img/blog/ferretdb-illacloud/5-query-all-data.png)

Then save and run it.

![Actions success for query all](/img/blog/ferretdb-illacloud/6-successful-action.png)

Let's add a table to the page and then connect it to the action.
Update the data source to read the result: `{{query_all_data.data[0].result}}`.

![Table and data source](/img/blog/ferretdb-illacloud/7-table_datasource.png)

### 2. Run aggregation operation

This is an aggregation operation that gets the total number of transactions per location.
In Illa, you can express it as:

```text
[
  {
    "$group": {
      "_id": "$customer_location",
      "totalTransactions": { "$sum": 1 }
    }
  },
   {
     "$sort": { "totalTransactions": -1 } }
]
```

![Action for aggregation operation](/img/blog/ferretdb-illacloud/8-aggregation.png)

### 3. Count all documents

Add a table component to visualize the total number of transactions in the collection.
Then use the `count` command to get the total.
Once that's we can display the result in a table with `{{count.data}}`

![Action for count operation](/img/blog/ferretdb-illacloud/9-count.png)

The combined view of the 3 actions should look like this:

![View of all actions](/img/blog/ferretdb-illacloud/10-overall-view.png)

### 4. Query with input

With Illa cloud, you can add an input text component to the page and use it's input to query the FerretDB instance.

Start by adding an input and a button component.

Then add a new action that uses the `find` command to query the collection based on the specified input.
The query for this action is: `{"product_category": "{{prod_cat.value}}"}`

![Showing the input and button components](/img/blog/ferretdb-illacloud/11-input-components.png)

Add a click event handler to the button component and assign the action to it.

![Show Gif of adding the event](/img/blog/ferretdb-illacloud/12-add-events.gif)

Finally, add a table to the screen and set the data source to read the result of the action.

![Show table with action result and all UI components](/img/blog/ferretdb-illacloud/13-add-table.png)

A preview of the UI is shown below:

![ Before before made Gif to show all UI components and search](/img/blog/ferretdb-illacloud/14-display-table.gif)

## Start working with FerretDB data on Illa Cloud

Anyone building on Illa Cloud can start leveraging FerretDB for their internal apps and admin panels.
With Illa Cloud and FerretDB, you can quickly build your apps in minutes using its numerous out-of-the-box components, drag-and-drop features, and database integrations.

Beyond its open source benefits, FerretDB has a growing community and amazing team ready to help with bugs, feature requests, questions – whatever you need.
You can learn more about [FerretDB here](https://docs.ferretdb.io/understanding-ferretdb/).

Want to know more about monitoring and visualization of data on FerretDB?
See [Grafana Monitoring and Visualization for FerretDB](https://blog.ferretdb.io/grafana-monitoring-ferretdb/).
