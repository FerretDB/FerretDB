FerretDB should handle the client's authentication commands and use provided credentials to authenticate in PostgreSQL.

Instead of using one global PostgreSQL connection pool, separate pools should be shared by all client connections with the same username and password.

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

