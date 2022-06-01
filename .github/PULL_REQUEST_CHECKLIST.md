
Pull Request checklist:
- linters passed.
- all funcs and variables should be documented.
- comments should answer on the question of why.
- code similarity.
- code complexity.
- no copy from licensed projects documentation.
- no copy from licensed projects source code.
- prefer "atomic" tests for separate methods. It's easier to read and understand.
- code generation functions must not break protocols (stringer comments for string value constants in protocols).
- use spell-checkers to check grammar and spelling.
- code coverage.
- a reference to variable names and descriptions.
- no extra test coverage - see the principle of having less code for simplicity and removing the artificial complexity.
- remove named parameters that we don't use.
- code consistency and similarity.
- must not treat fluky tests as lightly, ok
- sort alphabetically linters and bulks of variables
- for specific types, use one order everywhere
