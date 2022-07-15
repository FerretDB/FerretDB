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

| #    | command          | Description                                           |
|------|------------------|-------------------------------------------------------|
| 1    | listCommands     | Returns a list of commands.                           |
| 2    | connectionStatus | Returns connection status.                            |
| 3    | getParameter     | Response should be extended with authentication data. |
| 4    | serverStatus |                                                       |

## MongoDB's authentication mechanisms

* PLAIN (Relatively easy to implement)
* SCRAM-SHA-256 (Could be tricky because we need to deal with the user's password hash)
* SCRAM-SHA-1 (*not sure about this one as PostgreSQL doesn't support it*)
* X509 (It would be big and hard to implement task)
* GSSAPI Kerberos (Needs more research on that one)



# Tasks
## Add support for separate connection pools for each user 

We will add support for Tigris authentication once it would be clear enough.

*We will try to not store the username and password*.
We should use PostgreSQL's authentication mechanism.

As the first step we will implement PLAIN authentication mechanism.

When FerretDB starts up, it should not connect to PostgreSQL.
Parse the connection string and strip the username and password if they are present.

When a client connects with credentials specified in the connection string (username and password), authentication message would be sent to the server (`saslStart`).
We should handle that message:
* Check if we already have a connection pool for the user.
* If we don't, create a new connection pool.
* Connect to the database with provided credentials.
* If the connection is successful, send authentication status message to the client
* If the connection is not successful, send that message to the client.

`saslContinue` would not be supported for now.

When a client disconnects, we should close the connection pool.
We should close all user pools when the server is shut down.

### Tests:
* Connect to the server with username and password in connection string twice.
* Connect to the server with different usernames and passwords.
* Connect to the server without username and password in connection string.

## Add support for `createUser` command

Add support for the `createUser` command with the following parameters:
* createUser
* pwd
* mechanisms

We will support only PLAIN mechanism for now.

Not in the scope of this task:
* customData
* roles
* writeConcern
* authenticationRestrictions
* digestPassword
* comment

User data should be stored in the PostgreSQL database (`admin` database?).
Create a PostgreSQL user with provided credentials and auth mechanisms.
Grant user access for database specified in command.

### Tests:

* Create user with username and password.
* Create user with username and password and mechanisms.
* Create user with username and password and mechanism that not supported or not exists.
