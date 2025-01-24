---
sidebar_position: 11
---

# Known differences

<!--
   Each numbered point should have a corresponding, numbered test file https://github.com/FerretDB/FerretDB/tree/main/integration/diff_*_test.go
   Bullet subpoints should be in the same file as the parent point.
-->

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. Collection name must be valid UTF-8.

If you encounter some other difference in behavior, please [join our community](/#community) to report a problem.
