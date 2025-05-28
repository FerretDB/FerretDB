---
slug: powering-notification-infrastructure-ferretdb-meteor
title: 'Powering Your Notification Infrastructure with Novu and FerretDB'
authors: [alex]
description: >
  Discover how to leverage Novu's powerful notification capabilities with FerretDB as your robust, PostgreSQL-backed database.
image: /img/blog/ferretdb-meteor.jpg
tags: [compatible applications, open source, community]
---

![Meteor.js and FerretDB](/img/blog/ferretdb-meteor.jpg)

Effective communication with your users is critical for all applications.
From transactional emails to real-time in-app alerts, a robust notification system is crucial for user engagement and satisfaction.

<!--truncate-->

At FerretDB, we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.
This means you can get the best of both worlds: a document database experience with the stability of a relational database.

We're thrilled to explore how Novu, the open-source notification infrastructure, seamlessly integrates with FerretDB, offering a powerful and flexible solution for your notification needs.

## What is Novu?

Novu is an open-source notification infrastructure, designed to simplify and streamline multi-channel notification delivery.
It provides a unified API, a customizable in-app Inbox component, and a drag-and-drop workflow builder, allowing developers and product teams to manage and deliver notifications across channels like:

- In-app
- Email
- Push
- SMS
- Chat

With Novu, you can build comprehensive notification workflows, manage user preferences, and ensure timely and relevant communication without the complexities of building a notification system from scratch.

![A screenshot of Novu's dashboard showing its features, such as the workflow builder and notification channels]()

## Why Use Novu with FerretDB?

Novu uses MongoDB as its primary database for storing notification data, user preferences, and workflow configurations.
Because FerretDB is designed to be a wire-protocol compatible alternative to MongoDB, it can serve as a drop-in replacement for Novu's database.
This combination offers several compelling advantages:

- **Open-Source Harmony:** Both Novu and FerretDB are open-source, providing transparency, flexibility, and a strong community backing.
- **PostgreSQL Reliability for Notifications:** By using FerretDB as Novu's backend, you gain the battle-tested reliability, ACID compliance, and advanced features of PostgreSQL for your critical notification data.
- **Simplified Infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating Novu with FerretDB can simplify your database management and reduce operational overhead.
- **Scalability and Performance:** Leverage the scalability and performance characteristics of both Novu's efficient notification delivery and FerretDB's robust data handling.

## Getting Started: Connecting Novu to FerretDB

Connecting Novu to your FerretDB instance is straightforward, as Novu expects a MongoDB-compatible database.
Here's a general outline:

1.  **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
    If you haven't set it up yet, refer to our [FerretDB Quickstart Guide](link to FerretDB quickstart).
2.  **Self-Host Novu:** Novu offers self-hosting options, typically via Docker or by running individual services.
    You'll need to set up the Novu backend. \* [Link to Novu Self-Hosting Documentation - Example: `https://docs.novu.co/self-hosting/overview`]
3.  **Configure Novu's Database Connection:** When configuring Novu, you'll specify the MongoDB connection string.
    Instead of pointing to a MongoDB instance, you'll point it to your FerretDB instance.

          * **Environment Variable Example:**
              When running Novu via Docker or through environment variables, you'll typically set a `MONGO_URL` or similar variable.

              ```bash
              # Example for Novu's Docker setup (adjust as per Novu's latest docs)
              # Assuming FerretDB is running on localhost:27017
              docker run -e MONGO_URL="mongodb://127.0.0.1:27017/novu" \
                          -e NODE_ENV=production \
                          # ... other Novu environment variables ...
                          novu/novu
              ```

              Ensure to replace `127.0.0.1:27017` with your FerretDB host and port, and `novu` with your desired database name.

4.  **Launch Novu and Test:** Once configured, launch Novu.
    It should connect to FerretDB, create the necessary collections, and be ready to start sending notifications!
    You can then proceed with Novu's setup, creating workflows and sending test notifications.

        ![A screenshot of the Novu dashboard showing the database connection with FerretDB highlighted as the MongoDB-compatible backend.]()

## Conclusion

The integration of Novu and FerretDB provides a robust, scalable, and fully open-source solution for all your notification infrastructure needs.
By leveraging FerretDB's MongoDB compatibility, you can seamlessly integrate Novu into your PostgreSQL-backed stack, ensuring reliable and efficient communication with your users.

This synergy exemplifies FerretDB's mission: to empower developers with the tools they love, backed by the database they trust.

- Learn more about FerretDB: [Link to FerretDB website]
- Explore Novu's capabilities: [Link to Novu website]
- Join the FerretDB community: [Link to FerretDB community channels, e.g., Slack, GitHub]
- Check out more FerretDB integrations: [Link to FerretDB integrations page (if available)]
