---
sidebar_position: 4
---

# FerretDB Cloud

FerretDB Cloud is a managed service that lets you create, deploy, and manage FerretDB instances in the cloud.

## Prerequisites

Before you begin, ensure you have the following:

- A [FerretDB Cloud account](https://cloud.ferretdb.com/).
- MongoDB-compatible client installed (e.g., MongoDB Compass, mongosh).

## Creating a FerretDB Instance

1. Log in to your FerretDB Cloud account.

2. Navigate to the "Deployments" section to create a new FerretDB instance.

3. Click on the "Create" button to create a new instance and you will be prompted to choose a plan.

   - Select the subscription plan that best suits your needs.

        :::note
        FerretDB Cloud offers four different plans:
        - Free Tier: Ideal for developers, students, and small projects getting started with FerretDB.
        - Pro Tier: Designed for professional teams and growing businesses requiring predictable pricing and enterprise-grade features.
        - Enterprise Tier: Ideal for enterprises requiring dedicated support and premium service levels.
        - Bring Your Own Account - Enterprise Tier: Perfect for large enterprises requiring maximum flexibility and control
        :::

   - Select the cloud provider and region where you want to deploy your instance.
   - Configure the instance settings, such as instance name, storage size, and other options if applicable.

4. Review your configuration and click "Create" to deploy your FerretDB instance.
   The deployment process may take a few minutes.
5. Once the instance is created, you will see it listed in your instance dashboard with its connection details.

## Connecting to FerretDB Instance

1. Open your MongoDB-compatible client (e.g., MongoDB Compass, mongosh).
2. Use the connection details from your FerretDB instance dashboard to connect to the database.
   The connection string will typically look like this:

     ```text
     mongodb://<username>:<password>@<endpoint>
     ```

     :::tip
     You can find the <endpoint> under "Connectivity" in your FerretDB instance dashboard.
     :::

    Replace `<username>`, `<password>`, and `<endpoint>` with the actual values provided in your FerretDB instance dashboard.

3. Use the connection string to connect to your FerretDB instance.
   For example, if you are using `mongosh`, you can run:

   ```sh
   mongosh mongodb://<username>:<password>@<endpoint>
   ```