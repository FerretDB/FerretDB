---
sidebar_position: 2
---

# Known differences

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. FerretDB does not support NUL (`\0`) characters in strings.
3. Document keys must not contain `$` sign.
4. Database and collection names restrictions:
   * name cannot start with the reserved prefix `_ferretdb_`;
   * name must not include non-latin letters, spaces, dots, dollars or dashes;
   * collection name length must be less or equal than 120 symbols, database name length limit is 63 symbols;
   * name must not start with a number;
   * database name cannot contain capital letters.
5. For Tigris, FerretDB requires Tigris schema validation for `msg_create`: validator must be set as `$tigrisSchemaString`.
   The value must be a JSON string representing JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).

If you encounter some other difference in behavior,
please [join our community](https://github.com/FerretDB/FerretDB#community) to report a problem.
