<!--
    This is a custom pull request template for FerretDB Inc. engineers.
    It is more complex than the default template, but it contains some additional
    information that is useful for team members.
-->

# Description

This PR closes #{issue_number}.

<!--
    Write a short description to explain changes that are not mentioned in the initial issue.
    What were the reasons for those changes?
    Which decisions did you make and why?
    What else should reviewers know about your changes?
-->

## Readiness checklist

<!--
    If you want your changes to be merged quickly,
    please follow CONTRIBUTING.md.
-->

* [ ] I set assignee, reviewers, labels, project and sprint.
* [ ] I added tests for new functionality or bugfixes.
* [ ] I ran `task all`, and it passed.
* [ ] I added/updated documentation for exported and unexported functions, variables, types, etc.
* [ ] I checked complex documentation rendering with `bin/godoc -index -play -timestamps -http=127.0.0.1:6060`.
