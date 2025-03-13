---
sidebar_position: 11
---

# Known differences

We don't plan to address those known differences in behavior:

1. FerretDB uses the same protocol error names and codes as MongoDB,
   but the exact error messages may sometimes be different.
2. FerretDB collection names must be valid UTF-8; MongoDB allows invalid UTF-8 sequences.

We consider all other differences in behavior to be problems and want to address them.
Please [join our community](/#community) to report them.

<!--
   Each numbered point should have a corresponding, numbered test file https://github.com/FerretDB/FerretDB/tree/main/integration/diff_*_test.go
   Bullet subpoints should be in the same file as the parent point.

   This comment should not be on top to avoid showing on a DocCard: https://github.com/facebook/docusaurus/issues/10589.
-->
