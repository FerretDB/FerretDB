FerretDB should handle the client's authentication commands and use provided credentials to authenticate in PostgreSQL.

Instead of using one global PostgreSQL connection pool, separate pools should be shared by all client connections with the same username and password.

createUser, dropUser, etc

Update listCommands to include requiresAuth.
Update connectionStatus to include authInfo.

## Auth Commands

| command      | Description                  |
|--------------|------------------------------|
| authenticate | Authenticate with PostgreSQL |


## Users Management Commands

| command                  | Description                            |
|--------------------------|----------------------------------------|
| createUser               | Creates user.                          |
| dropAllUsersFromDatabase | Deletes all users from a database.     |
| dropUser                 | Remove a single user from database.    |
| grantRolesToUser         | Grants roles and privileges to a user. |
| revokeRolesFromUser      | Removes roles from a user.             |
| updateUser               | Updates a user's data.                 |
| usersInfo                | Returns information about the users.   |

## Roles Management Commands

| command                  | Description                                                                           |
|--------------------------|---------------------------------------------------------------------------------------|
| createRole               | Creates a role with specified privileges.                                             |
| dropRole                 | Deletes role.                                                                         |
| dropAllRolesFromDatabase | Deletes all user's roles from a database.                                             |
| grantPrivilegesToRole    | Grants privileges to a role.                                                          |
| grantRolesToRole         | Grants selected roles to a role.                                                      |
| invalidateUserCache      | Invalidates the in-memory cache of user information, including credentials and roles. |
| revokePrivilegesFromRole | Revokes privileges from a role.                                                       |
| revokeRolesFromRole      | Revokes selected roles from a role.                                                   |
| rolesInfo                | Returns roles info.                                                                   |
| updateRole               | Updates a role.                                                                       |

## Commands that require authentication

| command          | Description                 |
|------------------|-----------------------------|
| listCommands     | Returns a list of commands. |
| connectionStatus | Returns connection status.  |
