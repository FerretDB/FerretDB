---
slug: dynamic-database-credentials-with-hashicorp-vault-and-ferretdb
title: 'Dynamic Database Credentials with HashiCorp Vault and FerretDB'
authors: [alex]
description: >
  Learn how to integrate HashiCorp Vault's database secrets engine with FerretDB to generate dynamic credentials for enhanced security and management.
image: /img/blog/ferretdb-vault.jpg
tags: [compatible applications, community, tutorial]
---

![Dynamic Database Credentials with HashiCorp Vault and FerretDB](/img/blog/ferretdb-vault.jpg)

No application can be truly secure without managing its database credentials or secrets effectively.
Hardcoding secrets, using long-lived credentials, or manually rotating them introduces significant security risks and operational overhead.

<!--truncate-->

[HashiCorp Vault](https://www.hashicorp.com/products/vault) addresses these challenges by providing a centralized interface for managing secrets, including the ability to generate dynamic database credentials.

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.
Through HashiCorp Vault's database secrets engine, you can automate the management of database credentials for FerretDB.

This blog post explores how HashiCorp Vault's database secrets engine seamlessly integrates with FerretDB, enabling you to generate dynamic credentials for your PostgreSQL-backed document databases.

## What is HashiCorp Vault?

HashiCorp Vault is a comprehensive secrets management tool that provides:

- **Secure secret storage:** Store sensitive data like API keys, passwords, and certificates.
- **Dynamic secrets:** Generate on-demand, time-limited credentials for databases, cloud services, and more.
- **Data encryption:** Encrypt data in transit and at rest.
- **Leasing & revocation:** Automatically revoke secrets after a specified time or on demand.
- **Audit logging:** Track all secret access and changes.
- **Policy-based access control:** Define granular permissions for who can access what secrets.

Vault's database secrets engine is a powerful feature that integrates directly with various database types, including MongoDB, to automate the lifecycle of database credentials.

## Why use HashiCorp Vault with FerretDB?

HashiCorp Vault's database secrets engine provides native support for MongoDB, allowing it to connect and generate dynamic credentials.
Since FerretDB is designed as an open-source alternative to MongoDB, Vault can also connect to FerretDB and serve as a dynamic credential manager for your FerretDB instances.
This powerful combination offers several compelling advantages:

- **Enhanced security:** Generate credentials that expire automatically, drastically reducing the risk of compromised long-lived secrets.
- **Automated rotation:** Vault can automatically rotate the root credentials it uses to create dynamic users, further bolstering security.
- **Centralized management:** Manage database credentials from a single, secure Vault instance.

## Connecting HashiCorp Vault to FerretDB

Connecting HashiCorp Vault's database secrets engine to your FerretDB instance involves configuring Vault to use the MongoDB plugin and providing the FerretDB connection details.

Here's a step-by-step guide to get you started with a local Vault and FerretDB setup:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible, with authentication.
   If you haven't set it up yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
   You need a user with sufficient privileges to create, update, and revoke other users.
   This will be the "root user" Vault uses for its operations.

2. **Set up and initialize HashiCorp Vault:**
   For local testing, you can run Vault in dev mode.
   This will start a single-node Vault server that is unsealed and initialized.
   You can run Vault in a Docker container:

   ```sh
   docker run -d --name vault -p 8200:8200 \
   --cap-add=IPC_LOCK \
   -e 'VAULT_DEV_ROOT_TOKEN_ID=<root_token>' \
   -e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200' \
   hashicorp/vault:latest
   ```

   `<root_token>` should have the value for your development root token ID – replace it with any string you prefer for easy access during development.

   You need the Vault CLI to interact with the Vault server.
   Ensure that you have the Vault CLI installed and configured to PATH on your system.
   If you haven't installed it yet, you can follow the [Vault installation guide](https://developer.hashicorp.com/vault/tutorials/get-started/install-binary) to set it up.

   Set the `VAULT_ADDR` environment variable:

   ```sh
   export VAULT_ADDR='http://127.0.0.1:8200'
   ```

   Log in to Vault using the root token you set during the `docker run` command:

   ```sh
   vault login <root_token>
   ```

   You should see a success message confirming your login.

3. **Enable the database secrets engine:**
   Enable the database secrets engine at a desired path, for example, `ferretdb-creds`.
   This path will serve as the base for all database-related operations in Vault.

   ```sh
   vault secrets enable -path=ferretdb-creds database
   ```

4. **Configure the MongoDB plugin for FerretDB:**
   Configure the `ferretdb-creds` secrets engine to use the mongodb-database-plugin and provide the connection details for your FerretDB instance.
   The `connection_url` will use a templated format for enhanced security and to enable root credential rotation.
   `allowed_roles` specifies which Vault roles are permitted to use the connection to create or manage database users.
   In this guide, we will use `my-app-role` as the role name and specify it later when creating/managing dynamic credentials.

   ```sh
   vault write ferretdb-creds/config/ferretdb-conn \
   plugin_name=mongodb-database-plugin \
   allowed_roles="my-app-role" \
   connection_url="mongodb://{{username}}:{{password}}@host.docker.internal:27017/" \
   username="<ferretdb_username>" \
   password="<ferretdb_password>"
   ```

   Replace `<ferretdb_username>` and `<ferretdb_password>` with your FerretDB authentication credentials.

5. **Create a database role:**
   Create a role within Vault that defines the dynamic credentials' properties, such as the roles they will have in FerretDB and their time-to-live (TTL).

   :::note
   At the time of writing, FerretDB's `createUser` supports a specific set of roles for dynamic user creation.
   To successfully create a user via Vault's database secrets engine, the `creation_statements` must specify one of the allowed role combinations that FerretDB recognizes:
   - `[{role: "readAnyDatabase", db: "admin"}]` or
   - `[{role: "clusterAdmin", db: "admin"}, {role: "readWriteAnyDatabase", db: "admin"}]`

   Although Vault allows specifying `database_name` in the creation statement, FerretDB does not currently scope users to a particular database.

   All users are created globally.
   The `db` field in roles and `database_name` are accepted for compatibility but do not enforce database-specific access in FerretDB.
   :::

   For a user with `clusterAdmin` and `readWriteAnyDatabase` roles, you can create a role in Vault like this:

   ```sh
   vault write ferretdb-creds/roles/my-app-role \
   db_name=ferretdb-conn \
   creation_statements='{"roles": [{"role": "clusterAdmin", "db": "admin"}, {"role": "readWriteAnyDatabase", "db": "admin"}]}' \
   default_ttl="1h" \
   max_ttl="24h"
   ```

   - `db_name`: References the connection configuration created in the previous step (ferretdb-conn).
   - `creation_statements`: This specifies the conditions (roles, databases, etc.) for creating the new dynamic user.
     Here, it creates a user with `clusterAdmin` and `readWriteAnyDatabase` roles.
   - `default_ttl`: The default lease duration for generated credentials.
   - `max_ttl`: The maximum lease duration.

   If you require users with more granular access (e.g. `readWrite` on a specific database), monitor FerretDB's documentation for updates on broader role/authorization support or consider [contributing to the project](https://github.com/FerretDB/FerretDB).

6. **Generate dynamic credentials:**
   Now, an application or user can request dynamic credentials from Vault by reading from the role's path:

   ```sh
   vault read ferretdb-creds/creds/my-app-role
   ```

   Vault will dynamically generate a unique username and password, create that user in FerretDB, and return the credentials along with a lease ID.

   ```text
   Key                Value
   ---                -----
   lease_id           ferretdb-creds/creds/my-app-role/3OQtYhih1BAjiGHBgjfK3Nue
   lease_duration     1h
   lease_renewable    true
   password           <generated_password>
   username           <generated_username>
   ```

   You can then use these username and password to connect your application to FerretDB.

   After the `lease_duration` expires, Vault will automatically revoke these credentials from FerretDB, enhancing your security posture.

## Exploring dynamic credentials in FerretDB

After running `vault read ferretdb-creds/creds/my-app-role`, Vault creates a temporary user in FerretDB.
You can verify this by connecting directly to FerretDB (using your generated credentials).

```sh
mongosh mongodb://<generated_username>:<generated_password>@localhost:27017/admin
```

Then, list the users by running `db.getUsers()`.
A typical output will look like this:

```js
{
users: [
    {
    _id: 'admin.v-token-my-app-role-MmtCUcTXkDifvCwUBOpI-1753129631',
    userId: 'admin.v-token-my-app-role-MmtCUcTXkDifvCwUBOpI-1753129631',
    user: 'v-token-my-app-role-MmtCUcTXkDifvCwUBOpI-1753129631',
    db: 'admin',
    roles: [
        { role: 'readWriteAnyDatabase', db: 'admin' },
        { role: 'clusterAdmin', db: 'admin' }
    ]
    }
],
ok: 1
}
```

You'll see all current users and the dynamically generated user created by Vault.
This output demonstrates that HashiCorp Vault successfully creates users within FerretDB and assigns them specific roles, allowing for automated, secure management of database access.

## Conclusion

The integration of HashiCorp Vault's database secrets engine with FerretDB provides a robust and scalable solution for dynamic database credential management.
By leveraging FerretDB, you can seamlessly integrate it into your existing security and secrets management workflows with Vault, providing your document database operations with enhanced, automated security.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Explore HashiCorp Vault Documentation](https://developer.hashicorp.com/vault/docs)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
