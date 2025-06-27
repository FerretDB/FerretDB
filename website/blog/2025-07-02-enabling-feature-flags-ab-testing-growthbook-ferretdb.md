---
slug: enabling-feature-flags-ab-testing-growthbook-ferretdb
title: 'Enabling Feature Flags and A/B Testing with GrowthBook and FerretDB'
authors: [alex]
description: >
  Learn how to seamlessly integrate GrowthBook's powerful feature flagging and A/B testing capabilities with FerretDB.
image: /img/blog/ferretdb-growthbook.jpg
tags: [compatible applications, open source, community]
---

![Enabling Feature Flags and A/B Testing with GrowthBook and FerretDB](/img/blog/ferretdb-growthbook.jpg)

In modern software development, agile methodologies and data-driven decisions are key.
Feature flags and A/B testing allow teams to release features safely, test ideas, and personalize user experiences without full-scale deployments.
GrowthBook offers a powerful open-source solution for these critical tasks.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

This blog post explores how GrowthBook, an open-source feature flagging and A/B testing platform, seamlessly integrates with FerretDB, offering a robust and high-performing solution for your growth initiatives.

## **What is GrowthBook?**

[GrowthBook](https://www.growthbook.io/) is an open-source platform designed to help product teams and developers manage feature flags and run A/B tests.
It empowers organizations to:

- **Roll out features safely:** Control who sees new features and when, reducing risk.
- **Run A/B tests:** Scientifically test different versions of features to optimize for impact.
- **Personalize experiences:** Deliver tailored content and functionality to specific user segments.
- **Manage experiments:** Centralize experiment definitions, metrics, and results.
- **Integrate easily:** Provides SDKs for various platforms and languages.

GrowthBook focuses on providing a powerful, developer-friendly platform that integrates seamlessly into your existing workflow, enabling continuous experimentation and rapid iteration.

## Why Use GrowthBook with FerretDB?

GrowthBook uses MongoDB as the supported database backend for its self-hosted deployments, enabling it to store login credentials, cached experiment results, and metadata.
Given that FerretDB is designed to be a true open source alternative to MongoDB, it can serve as a drop-in replacement in GrowthBook.
This powerful combination offers several compelling advantages:

- **Open-source:** Both GrowthBook and FerretDB are open-source projects, providing transparency, flexibility, and strong community backing.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating GrowthBook with FerretDB can streamline your database management and reduce operational overhead.
- **No vendor lock-in:** Enjoy the freedom of truly open-source solutions without concerns about proprietary licensing or vendor lock-in.

## Connecting GrowthBook to FerretDB

Connecting GrowthBook to your FerretDB instance is straightforward.
Here's a step-by-step guide to get you started with a self-hosted GrowthBook instance:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
2. **Set up GrowthBook:** You can self-host GrowthBook using Docker – see [GrowthBook Self-Hosting Documentation for more details](https://docs.growthbook.io/self-host).
   You can use the following Docker command to run GrowthBook with FerretDB:

   ```sh
   docker run -d \
     --name growthbook \
     -p 3000:3000 \
     -p 3100:3100 \
     -v growthbook_uploads:/usr/local/src/app/packages/back-end/uploads \
     -e MONGODB_URI="mongodb://<username>:<password>@<host-address>:27017/growthbook/" \
     growthbook/growthbook:latest
   ```

   Replace `<username>`, `<password>`, and `<host>` with your FerretDB connection details.
   The `growthbook` database will be created automatically if it doesn't exist.
   Ensure that the FerretDB instance is accessible to the GrowthBook container.
   If you're running FerretDB locally, set the host address as `host.docker.internal` to enable the GrowthBook container to connect to your local FerretDB instance.

3. **Launch GrowthBook:**
   GrowthBook should connect to FerretDB, initialize its database, and be accessible via its web interface.
   You can access it at `http://localhost:3000` (or whatever port you specified in the Docker command).

You can now log into GrowthBook, define feature flags and create experiments, with all data seamlessly stored in FerretDB.

## Example of GrowthBook data in FerretDB

After setting up GrowthBook, create a feature flag or an experiment to see how GrowthBook interacts with FerretDB.
In the image below,, an experiment named `ferretdb-experiment` is created.

![A screenshot of the GrowthBook dashboard showing an experiment named "ferretdb-experiment" with a description and status.](/img/blog/growthbook-experiment.png)

GrowthBook creates various collections to manage its configuration, user data, feature flags, and experiment results; you can see these collections in the FerretDB instance.
Connect to your FerretDB instance using a MongoDB shell or GUI tool (like MongoDB Compass or Mongo Express) and switch to the GrowthBook database (default growthbook):

![A screenshot of a MongoDB GUI tool connected to FerretDB, showing the GrowthBook database and its collections.](/img/blog/growthbook-collections.png)

Now, let's query the `experiments` collection to see the feature flags and experiments defined in GrowthBook:

![A screenshot of a MongoDB shell or GUI tool showing the results of a query on the experiments collection, displaying feature flags with their names, descriptions, and statuses.](/img/blog/growthbook-data.png)

This output demonstrates that GrowthBook successfully writes and reads its document-based data into FerretDB, which in turn stores it efficiently in PostgreSQL, providing a reliable backend for your feature flags and experiments.

## Conclusion

The integration of GrowthBook and FerretDB provides a robust, scalable, and fully open-source solution for managing your feature flags and A/B tests.
By leveraging FerretDB, you can run their entire workloads in open source, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
