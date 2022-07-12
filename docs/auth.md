FerretDB should handle the client's authentication commands and use provided credentials to authenticate in PostgreSQL.

createUser, dropUser, etc

Update listCommands to include requiresAuth.
Update connectionStatus to include authInfo.

## Auth Commands

| #   | command      | Description                  |
|-----|--------------|------------------------------|
| 1   | authenticate | Authenticate with PostgreSQL |


## Users Management Commands

| #   | command                  | Description                            |
|-----|--------------------------|----------------------------------------|
| 1   | createUser               | Creates user.                          |
| 2   | dropAllUsersFromDatabase | Deletes all users from a database.     |
| 3   | dropUser                 | Remove a single user from database.    |
| 4   | grantRolesToUser         | Grants roles and privileges to a user. |
| 5   | revokeRolesFromUser      | Removes roles from a user.             |
| 6   | updateUser               | Updates a user's data.                 |
| 7   | usersInfo                | Returns information about the users.   |

## Commands that require authentication

| #   | command          | Description                 |
|-----|------------------|-----------------------------|
| 1   | listCommands     | Returns a list of commands. |
| 2   | connectionStatus | Returns connection status.  |


## Add support for separate connection pools for each user 

When a client connects with credentials specified in the connection string (username and password), authentication message would be sent to the server (`saslStart`).
In order to create a new connection pool, we should use the username from the authentication message.

When a client disconnects, we should close the connection pool.
We should also close the system pool and all user pools when the server is shut down.

To separate connection pools we should:
* Use current connection pool as the `system` pool (we will use it later to query the database for user's credentials).
* Handle `saslStart` message.
* To distinguish between users we should add to the Handler instance a map of connection pools, where the key is the username.

Questions:
* How to distinguish between users connected without authentication?
* Should we use a different connection pool for each user (for example for each host+port pair)?

Tests:
* Connect to the server with username and password in connection string twice.
* Connect to the server with different usernames and passwords.
