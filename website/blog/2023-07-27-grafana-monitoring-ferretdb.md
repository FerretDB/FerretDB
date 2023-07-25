---
slug: grafana-monitoring-ferretdb
title: 'Grafana Monitoring for FerretDB'
description: >
  FerretDB is well suited to leverage the MongoDB data source plugin for Grafana to monitor and analyze your data.
image: /img/blog/ferretdb-grafana.jpg
tags: [compatible applications, tutorial]
---

![Grafana Monitoring for FerretDB](/img/blog/ferretdb-grafana.jpg)

Monitoring tools play an essential role in overseeing the operations of a database, and [Grafana](https://grafana.com/) is one of the most popular monitoring tools.

<!--truncate-->

As the truly open-source alternative to MongoDB, [FerretDB](https://www.ferretdb.io/) helps you remain in the open source ecosystem and [avoid vendor lock-in](https://blog.ferretdb.io/5-ways-to-avoid-database-vendor-lock-in/).
It also allows you to use the same commands and queries you're used to, along with many applications that MongoDB users already know.

In this regard, FerretDB is well suited to leverage the MongoDB data source plugin for Grafana.
This blog post will cover Grafana and how you can use it to analyze and monitor your FerretDB data.

## What is Grafana?

Grafana is a comprehensive visualization and monitoring tool for managing and tracking data from multiple data sources such as MongoDB, ElasticSearch, PostgreSQL, MySQL, and Graphite, among others.

With insight into large data sets with complex infrastructures, Grafana provides you with configurable dashboards that you can use to tailor the different visualizations (or panels) to monitor your data.
Dashboards in Grafana consists of panels containing several views such as graph, table, and heatmap.

Besides this, Grafana offers many data source plugins – including the MongoDB plugin for Grafana – that can plug into your existing data.
This gives you direct access to your data in real time and enables you to visualize and build dashboards for your contextual metrics.

And since there's a MongoDB plugin for Grafana, you can also set up Grafana to work with FerretDB databases, without any hassles.

## How to Set Up Grafana Dashboards for your FerretDB data

### Prerequisites

- Grafana Enterprise (note that the MongoDB plugin for Grafana is only available for Grafana Enterprise and Grafana Cloud users)
- FerretDB connection URI

In this tutorial, we'll be making use of the FerretDB `issues` collection.
This is a typical document in this collection:

```json5
  {
    _id: ObjectId("64b7e557921f991a8697d6fd"),
    number: 179,
    url: 'https://github.com/FerretDB/FerretDB/issues/179',
    title: 'Support field types validation',
    state: 'OPEN',
    stateReason: '',
    createdAt: ISODate("2021-12-14T18:13:01.000Z"),
    closedAt: ISODate("0001-01-01T00:00:00.000Z"),
    labels: [ 'code/feature', 'not ready' ],
    milestone: '',
    author: 'AlekSi',
    assignees: [],
    typename: 'Issue'
  }
```

### How to connect Grafana with FerretDB

This section of the tutorial assumes that you have Grafana Enterprise and have your FerretDB Connection URI.

Once you're signed in, go to the navigation section; under "Administration", select "Plugins" to download the MongoDB plugin for Grafana.

![Install MongoDB plugin for Grafana](/img/blog/ferretdb-grafana/mongodb-plugin.png)

After installing the plugin, go to "Connections", select "Data sources" and add the connection string URI and authentication credentials to connect to the FerretDB database.

![Grafana Connection](/img/blog/ferretdb-grafana/grafana-connection.png)

Through Grafana, we will visualize and analyze the data to see the performance over time.
We'll also create several panels on the dashboard.

On the Grafana dashboard, you need to select the right data source; here we will select the `issues` data source and then run `db.issues.find()` in the query tab.
Then change the visualization to a table to see the list of all documents.

![Grafana table view of database](/img/blog/ferretdb-grafana/complete-table-view.png)

Now that the data is visible in Grafana, let's further scrutinize the data and gain some granular insight from it.

#### Panel 1: Total number of issues in the last 10 milestones

First, we want to know the number of issues the team has worked no leading to every milestone (release).
To do this, we'll be using the following aggregation pipeline operation, and setting the visualization in Grafana as bar chart:

```js
db.issues.aggregate([
  {
    $match: {
      milestone: { $ne: ['', 'Next'] }
    }
  },
  {
    $group: {
      _id: '$milestone',
      count: { $sum: 1 }
    }
  },
  { $sort: { _id: -1 } },
  { $limit: 10 }
])
```

To start, it will filter out the issues with an empty "milestone" value or not labeled as "Next".
Then it group these selected issues based on their milestones.
For each group, it then calculates the number of issues per milestone and stores the result in "count".

![Issues per Milestone](/img/blog/ferretdb-grafana/issues-per-milestone.png)

Instantly, through the chart, we can see that there were 38 issues leading up to the "milestone" of v1.1.0, compared to just 8 for v0.9.2.

Brilliant, right?
Just like that we've already gleaned insight into the total number of issues per milestone.

#### Panel 2: Total number of issues marked as "not ready"

The next thing we want to know is the total number of issues that are marked as "not ready" and we'll be using the gauge visualization for this.

The query for setting up this operation is:

```js
db.issues.aggregate([
  { $match: { state: 'OPEN' } },
  { $unwind: '$labels' },
  { $match: { labels: { $eq: 'not ready' } } },
  { $group: { _id: '$labels', count: { $sum: 1 } } }
])
```

![Issues marked as "not ready"](/img/blog/ferretdb-grafana/issues-not-ready.png)

Once again, we can track the total number of OPEN issues that are not yet ready to be worked on, and this could shed light on the current backlog.

#### Panel 3: Total number of "good first issue"

To make it easier for first time contributors, having a decent number of "good first issue" can be tremendously vital to their open source contribution.

The aggregation operation for generating this visualization will group all `OPEN` issues, and then unwind the labels array so we can work directly with the elements in the array; then we match, group, and count all labels with "good first issue".

```js
db.issues.aggregate([
  { $match: { state: 'OPEN' } },
  { $unwind: '$labels' },
  { $match: { labels: { $eq: 'good first issue' } } },
  { $group: { _id: '$labels', count: { $sum: 1 } } }
])
```

Similar to the previous panel, we'll be using the gauge visualization to show the total number of "good first issues"

![Total number of "good first issue"](/img/blog/ferretdb-grafana/good-first-issue.png)

#### Panel 4: Group all issues by their state

Here, we want group all the issues by their state and set the visualization as a pie chart.

```js
db.issues.aggregate([
  {
    $group: {
      _id: '$state',
      count: { $sum: 1 }
    }
  }
])
```

![Group issues by state](/img/blog/ferretdb-grafana/issues-per-state.png)

In the Pie chart, we can instantly see all the percentage of open issues vs closed issues.

#### Panel 5: Group all issues by labels

We can also group all the issues by their different labels.
Since the labels are in array, we will need to unwind the array elements and group them into their separate labels.

```js
db.issues.aggregate([
  { $unwind: '$labels' },
  {
    $match: {
      labels: {
        $in: [
          'code/feature',
          'code/chore',
          'code/bug',
          'code/enhancement',
          'code/bug-regression',
          'documentation'
        ]
      }
    }
  },
  { $group: { _id: '$labels', count: { $sum: 1 } } }
])
```

On Grafana, we will the pie chart visualization to showcase the labels

![Grouping of issues by label](/img/blog/ferretdb-grafana/issues-per-label.png)

#### Panel 6: Total number of issues created by authors

In this panel, we want to group and count all issues by "author" and sort them from highest count to the lowest.
The aggregation operation to use in the query:

```js
db.issues.aggregate([
  {
    $group: {
      _id: '$author',
      count: { $sum: 1 }
    }
  },
  {
    $sort: { count: -1 }
  },
  {
    $limit: 8
  }
])
```

We can then set the visualization to be a table.

![Total number of issues per author](/img/blog/ferretdb-grafana/issues-per-author.png)

Below, we can see all of these panels in one dashboard.

![Complete Dashboard View](/img/blog/ferretdb-grafana/complete-dashboard.png)

## Start Monitoring FerretDB Data with Grafana

So far, we have explored how you can seamlessly integrate Grafana with FerretDB using the Grafana MongoDB plugin.
With a user-friendly dashboard and several informative panels, Grafana enables you to visualize data and gain insights into the current state of your database.

Of course, there's so much more that we can do with Grafana monitoring and FerretDB, and the examples in this blog post only gives you a glimpse of what's possible, and we can't wait to see what you come up with.

Check out our installation guide to [get started](https://docs.ferretdb.io/quickstart-guide/).
