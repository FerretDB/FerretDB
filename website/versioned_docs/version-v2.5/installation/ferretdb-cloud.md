---
sidebar_position: 4
description: Learn how to create, deploy, and manage FerretDB instances using FerretDB Cloud.
---

# FerretDB Cloud

FerretDB Cloud is a managed service that lets you create, deploy, and manage FerretDB instances in the cloud.

## Create an account

1. Visit [FerretDB Cloud](https://cloud.ferretdb.com/signup) to create an account.

2. Fill in the required information, including your name, email address, company name and URL (if applicable), and preferred password.

3. Submit the form to create your account.
   You will receive a verification email; click the link in the email to activate your account.
   Be sure to check your spam folder if you don't see the email within a few minutes.

   Once activated, sign in to your FerretDB Cloud account.

## Create a FerretDB Cloud instance

Ensure that you are logged in to your FerretDB Cloud account.

1. From the main dashboard, click on "Deployments" in the left-hand sidebar, then click "Create".
   You will be prompted to choose a subscription plan.
   To select a plan, click its "Subscribe" button.
   You will be notified by email upon approval.

   :::info
   FerretDB Cloud offers four different subscription plans:
   - **Free Tier:** Ideal for developers, students, and small projects getting started with FerretDB.
   - **Pro Tier:** Designed for professional teams and growing businesses requiring predictable pricing and enterprise-grade features.
   - **Enterprise Tier:** Ideal for enterprises requiring dedicated support and premium service levels.
   - **Bring Your Own Account - Enterprise Tier:** Perfect for large enterprises requiring maximum flexibility and control.
     :::

2. Configure your instance by selecting the cloud provider, region, network type, compute size, among other settings.
   The default settings are a great starting point, but you can customize them based on your requirements.

3. Review your configuration and click "Create" to deploy the instance.
   The deployment process may take a few minutes with the instance "Lifecycle Status" showing 'Deploying'.

4. Once the deployment is complete, the status will change to 'Running'.
   That means the instance is now active and ready for use.
   Click on the instance to view its connection details.

## Connect to a deployed instance

1. In the instance dashboard, find the username, password, and endpoint for your deployed instance.

   Using these details, set up your connection string as follows:

   :::caution
   TLS/SSL connections are not supported in Free tier instances - Ensure to connect without adding (`?tls=true`) to the connection string.
   Also, keep in mind Free Tier instances are temporary and may be stopped or deleted if inactive.

   Paid tiers (Pro, Enterprise, and Bring Your Own Account - Enterprise) require TLS/SSL connections to be enabled by adding (`?tls=true`) to the connection string.
   :::

   For Paid tiers (Pro, Enterprise, and Bring Your Own Account - Enterprise):

   ```text
   mongodb://<username>:<password>@<endpoint>/?tls=true
   ```

   For Free Tier instances:

   ```text
   mongodb://<username>:<password>@<endpoint>
   ```

   Replace `<username>`, `<password>`, and `<endpoint>` with the actual values provided in your instance dashboard.
   You can find the `<endpoint>` under the "Connectivity" tab.

2. To interact with your database, you need a client application on your computer (e.g., MongoDB Compass, `mongosh`, or any other MongoDB-compatible client).

   Use the connection string to connect to the instance.
   For example, if you are using `mongosh`, you can run:

   For Paid tiers (Pro, Enterprise, and Bring Your Own Account - Enterprise):

   ```sh
   mongosh mongodb://<username>:<password>@<endpoint>/?tls=true
   ```

   For Free Tier instances:

   ```sh
   mongosh mongodb://<username>:<password>@<endpoint>
   ```

   Replace `<username>`, `<password>`, and `<endpoint>` with the actual values for your instance.
