
Pull Request checklist:
- [ ] Clean code
  - [ ] linters passed (`task all`, or simply `task`).
  - [ ] for specific types use one order everywhere.
  - [ ] sort alphabetically linters and bulks of variables,
        * example https://pkg.go.dev/github.com/FerretDB/FerretDB/internal/types#hdr-Mapping.
        * linter https://github.com/FerretDB/FerretDB/pull/654
  - [ ] all funcs and variables should be documented
       * `revive` linter is enabled.
  - [ ] code similarity https://github.com/FerretDB/FerretDB/issues/694.
  - [ ] variable names, function names etc should tell us themselves what do they do, be clear and not misleading (Code Complete by Steve McConnel)
  - [ ] code generation functions must not break protocols (stringer comments for string value constants in protocols)
  - [ ] code smells removed (long varible names, long parameter list, repeating code etc) https://refactoring.guru/
  - [ ] remove named parameters that we don't use
  - [ ] watch and remove code cyclomatic complexity, do early returns.
  - [ ] code complexity: don't add artificial complexity.
  - [ ] big functions also should be extracted into smaller ones
- [ ] Tests
  - [ ] prefer "atomic" tests for separate methods. It's easier to read and understand.
  - [ ] should not treat fluky tests as lightly ok: change tests in separate PR or research
  - [ ] check code coverage
  - [ ] no extra test coverage - see the principle of having less code for simplicity and removing the artificial complexity.
        * https://rbcs-us.com/documents/Why-Most-Unit-Testing-is-Waste.pdf
        * https://rbcs-us.com/documents/Segue.pdf
- [ ] Documentation
  - [ ] documentation must be updated to corresponded changes with examples when possible.
  - [ ] comments should answer on question why and what for.
       * `checkConsumed` function for fuzzing
  - [ ] a reference to variable names and descriptions
  - [ ] use spell-checkers to check grammar and spelling.
  - [ ] godoc should look OK, check with `godoc -index -play -timestamps -http=127.0.0.1:6060`
- [ ] Legal
  - [ ] no copy of licensed projects' source code
  - [ ] no copy of licensed projects documentation