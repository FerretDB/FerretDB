---
sidebar_position: 6
slug: /diff/
---

# Known differences

<!--
   Each numbered point should have a corresponding test file in https://github.com/FerretDB/dance/tree/main/tests/diff
   Bullet subpoints should be in the same file as the parent point.
-->

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. FerretDB does not support NUL (`\0`) characters in strings.
3. FerretDB does not support nested arrays.
4. Document keys must not contain `$` or `.` signs.
5. Database and collection names restrictions:
   * name cannot start with the reserved prefix `_ferretdb_`;
   * database name must not include non-latin letters, spaces, dots, dollars or dashes;
   * collection name must not include non-latin letters, spaces, dots or dollars;
   * name must not start with a number;
   * database name cannot contain capital letters;
   * database name length cannot be more than 63 characters;
   * collection name length cannot be more than 120 characters.
6. For Tigris, FerretDB requires Tigris schema validation for `create` command: validator must be set as `$tigrisSchemaString`.
   The value must be a JSON string representing JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).

If you encounter some other difference in behavior,
please [join our community](https://github.com/FerretDB/FerretDB#community) to report a problem.
