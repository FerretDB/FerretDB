---
slug: building-project-management-stack-wekan-ferretdb
title: 'Build Your Project Management Stack with WeKan and FerretDB'
authors: [alex]
description: >
  Learn how to use WeKan, the open-source Trello-like Kanban board, with FerretDB, leveraging a reliable PostgreSQL-backed database for your project data.
image: /img/blog/ferretdb-wekan.jpg
tags: [compatible applications, open source, community]
---

![Build Your Project Management Stack with WeKan and FerretDB](/img/blog/ferretdb-wekan.jpg)

Effective project management and task organization are crucial for teams of all sizes.
[WeKan](https://wekan.fi/), an open-source Trello-like Kanban board, provides a flexible and collaborative platform for visualizing workflows and tracking progress.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this blog post, we're excited to explore how WeKan, the open-source Kanban board, seamlessly integrates with FerretDB, offering a robust, open-source, and self-hostable solution for your project management needs.

## What is WeKan?

WeKan is a free and open-source Kanban board application.
It's a popular choice for teams looking for a self-hostable alternative to proprietary project management tools.
Key features include:

- **Kanban boards:** Visualize tasks as cards on boards, moving them through various stages of a workflow.
- **Real-time collaboration:** Multiple users can collaborate on boards, cards, and comments in real-time.
- **Customizable:** Create custom fields, labels, and swimlanes to adapt to any workflow.
- **Attachments & checklists:** Add files, images, and sub-tasks to cards.
- **User management:** Manage users, teams, and permissions within the application.
- **Open source & self-hostable:** Full control over your data and deployment.

WeKan empowers teams to improve their organization, communication, and productivity by providing a clear visual overview of their projects.

## Why use WeKan with FerretDB?

WeKan uses MongoDB as its primary database backend for storing board data, cards, users, and settings.
Since FerretDB is designed to be an open-source alternative to MongoDB, it can serve as a drop-in replacement for WeKan's database.
This combination offers several advantages:

- **Full open-source stack:** Create a complete, transparent, and controllable project management stack, from your Kanban board (WeKan) to your database (FerretDB + PostgreSQL), eliminating proprietary lock-in.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating WeKan with FerretDB can streamline your database management and reduce operational overhead.
- **Community-driven:** Both WeKan and FerretDB are vibrant open-source projects with active communities, providing robust support and continuous development.

## Connecting WeKan to FerretDB

Connecting a self-hosted WeKan instance to your FerretDB instance is straightforward, as WeKan expects a MongoDB-compatible database.

Here's a step-by-step guide to get you started with a self-hosted WeKan instance using Docker Compose:

### Set up a Docker Compose file

Create a `docker-compose.yml` file with the following content to define the services for WeKan and FerretDB:

```yaml
services:
  ferretdb:
    image: ghcr.io/ferretdb/ferretdb-eval:2
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - POSTGRES_USER=<username>
      - POSTGRES_PASSWORD=<password>
      - POSTGRES_DB=postgres
    volumes:
      - ./ferretdb_data:/var/lib/postgresql/data
  wekan:
    image: ghcr.io/wekan/wekan:latest
    restart: always
    ports:
      - 80:8080
    environment:
      - MONGO_URL=mongodb://<username>:<password>@ferretdb:27017/wekan
      - ROOT_URL=http://localhost
      - WRITABLE_PATH=/data
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - ./wekan-files:/data:rw

networks:
  default:
    name: ferretdb-network
```

This setup defines two services: `ferretdb` for the FerretDB instance and `wekan` for the WeKan application.
The `ferretdb` service uses the [FerretDB evaluation image](https://docs.ferretdb.io/installation/evaluation/), which is designed for quick testing and experiments.

Replace `<username>` and `<password>` with your desired credentials.
`WRITABLE_PATH` is set to `/data`, which is where WeKan will store its files and attachments.

### Launch services and access FerretDB and WeKan

Run the following command in the directory where your `docker-compose.yml` file is located:

```sh
docker compose up -d
```

Once the services are up and running, you can access WeKan by navigating to `http://localhost:80` in your web browser.
You can now sign up, create boards, and start managing your projects.
All data will be stored in FerretDB.

## Exploring WeKan project data in FerretDB

The WeKan interface allows you to create and manage tasks visually, while FerretDB handles the underlying data storage.
In the image below, you can see a WeKan board with several lists and cards set up for managing tasks.

![WeKan Board](/img/blog/wekan-board.png)

Let's inspect how WeKan stores its data within FerretDB.
Connect to your FerretDB instance using `mongosh` or a GUI tool (like MongoDB Compass or DBeaver).

```sh
mongosh mongodb://<username>:<password>@localhost:27017/wekan
```

Ensure to replace `<username>` and `<password>` with the credentials you set in the `docker-compose.yml` file.

WeKan creates numerous collections to manage all aspects of a Kanban board.
Now, let's query the cards collection by running `db.cards.find()` to see how WeKan stores its cards:

```js
{
  _id: 'EHZTgExJbq4CBE544',
  title: 'New content on FerretDB and WeKan',
  members: [],
  labelIds: [],
  customFields: [],
  listId: 'tFNtC3ACHGmbAuEyQ',
  boardId: 'ixLgNAAYsuEfWZ5wJ',
  sort: 0,
  swimlaneId: 'AqHDB9p8HiL5NpCRg',
  type: 'cardType-card',
  cardNumber: 1,
  archived: false,
  parentId: '',
  coverId: '68823f43e307057ebb2e3f95',
  createdAt: ISODate('2025-07-24T14:11:54.730Z'),
  modifiedAt: ISODate('2025-07-28T04:35:13.844Z'),
  dateLastActivity: ISODate('2025-07-28T04:35:13.844Z'),
  description: '',
  requestedBy: '',
  assignedBy: '',
  assignees: [],
  spentTime: 0,
  isOvertime: false,
  userId: 'n7znsRbkQbri9TpKj',
  subtaskSort: -1,
  linkedId: '',
  vote: {
    question: '',
    positive: [],
    negative: [],
    end: null,
    public: false,
    allowNonBoardMembers: false
  },
  poker: {
    question: false,
    one: [],
    two: [],
    three: [],
    five: [],
    eight: [],
    thirteen: [],
    twenty: [],
    forty: [],
    oneHundred: [],
    unsure: [],
    end: null,
    allowNonBoardMembers: false
  },
  targetId_gantt: [],
  linkType_gantt: [],
  linkId_gantt: []
}
```

This output demonstrates that WeKan successfully writes and reads its document-based data into FerretDB, which in turn stores it efficiently in PostgreSQL, providing a reliable backend for your collaborative project boards.
You can also add other metadata to your cards, such as attachments, comments, checklists, and subtasks.

## Conclusion

The integration of WeKan and FerretDB provides a robust, scalable, and fully open-source solution for self-hosting your project management and Kanban boards.
By leveraging FerretDB, you can seamlessly build a complete project management stack that is fully open-source, eliminating vendor lock-in and proprietary licensing concerns.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
