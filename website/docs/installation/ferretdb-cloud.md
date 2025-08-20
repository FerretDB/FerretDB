---
sidebar_position: 4
description: Learn how to create, deploy, and manage FerretDB instances using FerretDB Cloud.
---

# FerretDB Cloud

FerretDB Cloud is a managed service that lets you create, deploy, and manage FerretDB instances in the cloud.

## Create an account

1. Visit [FerretDB Cloud](https://cloud.ferretdb.com/signup) to create an account.

2. Fill in the required information, including your name, email address, company name, company URL, and preferred password.

3. Submit the form to create your account.
   You will receive a verification email; click the link in the email to activate your account.
   Once activated, sign in to your FerretDB Cloud account.

## Create a FerretDB Cloud instance

Ensure that you are logged in to your FerretDB Cloud account.

1. Navigate to the "Deployments" section and click "Create".
   You will be prompted to choose a subscription plan.
   To select a plan, click its "Subscribe" button.
   You will be notified by email upon approval.

   :::note
   FerretDB Cloud offers four different subscription plans:
   - Free Tier: Ideal for developers, students, and small projects getting started with FerretDB.
   - Pro Tier: Designed for professional teams and growing businesses requiring predictable pricing and enterprise-grade features.
   - Enterprise Tier: Ideal for enterprises requiring dedicated support and premium service levels.
   - Bring Your Own Account - Enterprise Tier: Perfect for large enterprises requiring maximum flexibility and control.
     :::

2. Configure your instance by selecting the cloud provider, region, network type, compute size, among other settings.

3. Review your configuration and click "Create" to deploy your FerretDB Cloud instance.
   The deployment process may take a few minutes.

4. Once the instance is created, you will see it listed in your instance dashboard with its connection details.

## Connect to deployed instance

1. Open your MongoDB-compatible client (e.g., MongoDB Compass, `mongosh`).
2. Use the connection string provided in your FerretDB Cloud instance dashboard.
   The connection string will typically look like this:

   ```text
   mongodb://<username>:<password>@<endpoint>
   ```

   :::tip
   You can find the `<endpoint>` under the "Connectivity" tab in your FerretDB instance dashboard.
   :::

3. Replace `<username>`, `<password>`, and `<endpoint>` with the actual values provided in your FerretDB instance dashboard.
   Use the connection string to connect to your FerretDB instance.
   For example, if you are using `mongosh`, you can run:

   ```sh
   mongosh mongodb://<username>:<password>@<endpoint>
   ```
