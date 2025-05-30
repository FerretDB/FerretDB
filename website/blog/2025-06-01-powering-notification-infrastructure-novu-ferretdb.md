---
slug: powering-notification-infrastructure-novu-ferretdb
title: 'Powering Your Notification Infrastructure with Novu and FerretDB'
authors: [alex]
description: >
  Discover how to leverage Novu's powerful notification capabilities with FerretDB as your robust, PostgreSQL-backed database.
image: /img/blog/ferretdb-novu.jpg
tags: [compatible applications, open source, community]
---

![Powering Your Notification Infrastructure with Novu and FerretDB](/img/blog/ferretdb-novu.jpg)

Effective communication with your users is critical for all applications.
From transactional emails to real-time in-app alerts, a robust notification system is needed for user engagement and satisfaction.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.
This means you can get the best of both worlds: a document database experience with the stability of a relational database.

We're thrilled to explore how Novu, the open-source notification infrastructure, seamlessly integrates with FerretDB, offering a powerful and flexible solution for your notification needs.

## What is Novu?

[Novu](https://novu.co/) is an open-source notification infrastructure, designed to simplify and streamline multi-channel notification delivery.
It provides a unified API, a customizable in-app Inbox component, and a drag-and-drop workflow builder, allowing developers and product teams to manage and deliver notifications across channels like:

- In-app
- Email
- Push
- SMS
- Chat

With Novu, you can build comprehensive notification workflows, manage user preferences, and ensure timely and relevant communication without the complexities of building a notification system from scratch.

## Why use Novu with FerretDB?

Novu uses MongoDB as its primary database for storing notification data, user preferences, and workflow configurations.
Because FerretDB is designed to be a compatible alternative to MongoDB, it can serve as a drop-in replacement for Novu's database.
This combination offers several compelling advantages:

- **Open-source:** Both Novu and FerretDB are open-source, providing transparency, flexibility, and strong community support.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating Novu with FerretDB can simplify your database management and reduce operational overhead.
- **Scalability and lack of vendor lock-in:** Leverage the scalability and performance characteristics of both Novu's efficient notification delivery and FerretDB without fear of vendor lock-in or licensing issues.

## Connecting Novu to FerretDB

Connecting Novu to your FerretDB instance is straightforward, as Novu expects a MongoDB-compatible database.
Here's a step-by-step guide to get you started:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
    If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
2. **Self-host Novu:** Novu offers self-hosting options, typically via Docker or by running individual services.
    You'll need to set up Novu — [refer to Novu's documentation](https://docs.novu.co/community/self-hosting-novu/overview) for detailed instructions.
3. **Configure Novu's database connection:** When configuring Novu, instead of pointing to a MongoDB instance, update the `.env` file to point to your FerretDB instance.

    You'll need to update the `MONGO_URL` environment variable, as shown below (assuming FerretDB is running on `127.0.0.1:27017`):

    ```text
    MONGO_INITDB_ROOT_USERNAME=<username>
    MONGO_INITDB_ROOT_PASSWORD=<password>
    MONGO_URL=mongodb://$MONGO_INITDB_ROOT_USERNAME:$MONGO_INITDB_ROOT_PASSWORD@localhost:27017/novu-db
    MONGO_AUTO_CREATE_INDEXES=true
    ```

    Ensure to replace `127.0.0.1:27017` with your FerretDB host and port.

4.  **Launch Novu and test:** Once configured, launch Novu.
    It should connect to FerretDB, create the necessary collections, and be ready to start sending notifications!

    You can then proceed with Novu's setup, creating workflows and sending test notifications.

### Example of Novu with FerretDB as the backend

Below is a screenshot of the Novu dashboard with a notification workflow created to send an in-app notification to a user:

![A screenshot of a notification workflow in the Novu dashboard with FerretDB as the backend](/img/blog/novu-workflow-dashboard.png)

The next image shows the in-app notification received by a user:

![A screenshot of a Novu in-app notification received by a user](/img/blog/novu-notification.png)

Here is a view of all the collections created in FerretDB for Novu as well as the workflow and notifications created:

```text
> use novu-db
switched to db novu-db
> show collections
changes
controls
environments
executiondetails
feeds
integrations
jobs
layouts
members
messages
messagetemplates
notificationgroups
notifications
notificationtemplates
organizations
preferences
subscribers
tenants
topics
topicsubscribers
users
workflowoverrides
```

The following command shows the most recent message sent by Novu, which is stored in the `messages` collection in FerretDB:

```text
> db.messages.find().sort({ createdAt: -1 }).limit(1)
[
  {
    _id: ObjectId('6839dcaf213a25d911bae32e'),
    _templateId: ObjectId('6839db04662b0a0da8fcd8a7'),
    _environmentId: ObjectId('6839c07d33b29b25befa6d03'),
    _messageTemplateId: ObjectId('6839dc53fc5f8f606848f8d2'),
    _notificationId: ObjectId('6839dcae21b36bad120921f9'),
    _organizationId: ObjectId('6839c07d33b29b25befa6cfc'),
    _subscriberId: ObjectId('6839dafffc5f8f606848f87c'),
    _jobId: ObjectId('6839dcae21b36bad120921fd'),
    templateIdentifier: 'onboarding-demo-workflow',
    subject: 'Message from FerretDB',
    cta: { type: 'redirect', action: { buttons: [] } },
    _feedId: null,
    channel: 'in_app',
    content: 'Hi,\n\nThis is a notification with FerretDB!',
    providerId: 'novu',
    deviceTokens: [],
    seen: false,
    read: false,
    archived: false,
    status: 'sent',
    transactionId: '411e35f7-2de7-4eef-ac46-0e91f5972223',
    payload: {
      subject: 'subject',
      body: 'body',
      primaryActionLabel: 'primaryActionLabel',
      secondaryActionLabel: 'secondaryActionLabel',
      __source: 'dashboard'
    },
    tags: [],
    avatar: 'https://dashboard-v2.novu.co/images/info.svg',
    createdAt: ISODate('2025-05-30T16:28:31.377Z'),
    updatedAt: ISODate('2025-05-30T16:28:31.377Z'),
    __v: 0
  }
]
```

That's it! You now have a fully functional notification system powered by Novu and FerretDB.

## Conclusion

The integration of Novu and FerretDB provides a robust, scalable, and fully open-source solution for all your notification infrastructure needs.
By leveraging FerretDB's MongoDB compatibility, you can seamlessly integrate Novu into your PostgreSQL-backed stack, ensuring reliable and efficient communication with your users.

This synergy exemplifies FerretDB's mission: to enable developers to run their MongoDB workloads in open source, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
