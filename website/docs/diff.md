---
sidebar_position: 8
slug: /diff/ # referenced in README.md and beacon
---

# Known differences

<!--
   Each numbered point should have a corresponding test file in https://github.com/FerretDB/dance/tree/main/tests/diff
   Bullet subpoints should be in the same file as the parent point.
-->

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. FerretDB does not support NUL (`\0`) characters in strings.
3. FerretDB does not support nested arrays.
4. FerretDB converts `-0` (negative zero) to `0` (positive zero).
5. Document restrictions:
   * document keys must not contain `.` sign;
   * document keys must not start with `$` sign;
   * document fields of double type must not contain `Infinity`, `-Infinity`, or `NaN` values.
6. When insert command is called, insert documents must not have duplicate keys.
7. Update command restrictions:
   * update operations producing `Infinity`, `-Infinity`, or `NaN` are not supported.
8. Database and collection names restrictions:
   * name cannot start with the reserved prefix `_ferretdb_`;
   * database name must not include non-latin letters;
   * collection name must be valid UTF-8 characters;
   * database name must not start with a number;
   * database name cannot contain capital letters;
9. For Tigris, FerretDB requires Tigris schema validation for `create` command: validator must be set as `$tigrisSchemaString`.
    The value must be a JSON string representing JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).

If you encounter some other difference in behavior,
please [join our community](/#community) to report a problem.
