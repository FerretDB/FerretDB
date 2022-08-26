# Known differences

_TODO:_ These differences need to be documented properly.

## Tigris

### Validator

FerretDB requires Tigris schema validation for `msg_create`: validator must be set as `$tigrisSchemaString`.
The value must be a JSON string representing JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).
