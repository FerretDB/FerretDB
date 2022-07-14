# Task description

* Instead of using one global PostgreSQL connection pool, separate pools should be shared by all client connections with the same username and password.

* FerretDB should handle the client's authentication commands and use provided credentials to authenticate in PostgreSQL.

* Implement createUser, dropUser, etc. commands.
Update listCommands to include requiresAuth.
Update connectionStatus to include authInfo.

# Overview of the authentication commands

## Useful links

[MongoDB auth docs](https://github.com/mongodb/mongo/blob/a06bc8bbced8f0c60b94ed784f5f105f2f01ed5d/src/mongo/db/auth/README.md)

[Authentication reference](https://www.mongodb.com/docs/manual/core/authentication/)

[SCRAM authentication](https://www.mongodb.com/docs/manual/core/security-scram/)

## Auth Commands

| #   | command      | Description                        |
|-----|--------------|------------------------------------|
| 1   | authenticate | Authenticate with x.509 mechanism. |
| 2   | saslStart    | Start SASL authentication.         |
| 3   | saslContinue | Continue SASL authentication.      |


## Users Management Commands

| #   | command                  | Description                            |
|-----|--------------------------|----------------------------------------|
| 1   | createUser               | Creates user.                          |
| 2   | dropAllUsersFromDatabase | Deletes all users from a database.     |
| 3   | dropUser                 | Remove a single user from database.    |
| 4   | updateUser               | Updates a user's data.                 |
| 5   | usersInfo                | Returns information about the users.   |

## Commands that require authentication

| #   | command          | Description                                 |
|-----|------------------|---------------------------------------------|
| 1   | listCommands     | Returns a list of commands.                 |
| 2   | connectionStatus | Returns connection status.                  |
| 3   | adminCommand     | `getParameter` response should be extended. |


# Tasks
## Add support for separate connection pools for each user 

*We will try to not store the username and password*.
We should try to use PostgreSQL's authentication mechanism.

When a client connects with credentials specified in the connection string (username and password), authentication message would be sent to the server (`saslStart`).
In order to create a new connection pool, we should use the username from the authentication message.

When a client disconnects, we should close the connection pool.
We should also close the system pool and all user pools when the server is shut down.

To separate connection pools we should:
* Use current connection pool as the `system` pool (we will use it later to query the database for user's credentials).
* Handle `saslStart` message.
* To distinguish between users we should add to the Handler instance a map of connection pools, where the key is the username.

### Questions:
* How to distinguish between users connected without authentication?
* Should we use a different connection pool for each user (for example for each host+port pair)?

### Tests:
* Connect to the server with username and password in connection string twice.
* Connect to the server with different usernames and passwords.

## Add support for `createUser` command

Add support for the `createUser` command with the following parameters:
* createUser
* pwd
* mechanisms

Not in the scope of this task:
* customData
* roles
* writeConcern
* authenticationRestrictions
* digestPassword
* comment

User data should be stored in the PostgreSQL database (`admin` database?).
Create a PostgreSQL user with provided credentials and auth mechanisms.
Grant user access for all databases.

### Tests:

* Create user with username and password.
* Create user with username and password and mechanisms.
* Create user with username and password and mechanism that not supported or not exists.

## Add flag `authenticationMechanisms`

**Note:** This one should be discussed with the team.

Add flag `authenticationMechanisms` to the FerretDB server.
It should be an array of strings, containing the names of the supported authentication mechanisms.
Possible values are:
* SCRAM-SHA-256 (Most used)
* PLAIN (Relatively easy to implement)

Not in the scope of this task:
* SCRAM-SHA-1 (*not sure about this one as PostgreSQL doesn't support it*)
* X509 (It would be big and hard to implement task)
* GSSAPI Kerberos (Not sure would someone use this type of authentication)

This one will allow to restrict authentication mechanisms that could be used with `createUser`.
