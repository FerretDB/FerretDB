# The Process

This file contains some process notes for FerretDB Inc. engineers and long-term contributors.
If you are a community contributor, feel free to ignore all this!
We will do all that for you so you can focus on what is really interesting for you.

This guide tries to be short and does not mention things that are completely automated by our tooling.

## Team performance and planning

We want to deliver fast and make our deliverables predictable.
To achieve that, we measure the flow of work and estimate tasks.
The flow of work is measured automatically based on GitHub's pull request workflow.
Tasks are estimated manually by engineers before or during Sprint planning.

Task estimation depends on the following parameters:

* Scope.
  Are the changes required in one or multiple files, packages, components?
* Difficulty.
  How hard is the task?
  Does it require to cover a lot of test cases?
  edge cases?
* Clarity.
  Are the description and definition of done clear?
  Or that's more of a research task?

We use the following T-Shirt Sizes to estimate tasks:

* **S**: Small simple clear task.
* **M**: Only one parameter can be changed compared to **S** (e.g. a small but not completely clear task).
* **L**: Two parameters can be changed compared to **S** (e.g. clear but big and somewhat complex task).

If the team thinks that the task is bigger than **L**, it should be decomposed into smaller tasks.

Unless the issue explicitly states otherwise, the following things are always in the scope:

* All handlers.
* Tests.
  See contributing documentation for general discussion about unit and integration tests.
* Small spot refactorings.
  Larger refactorings should be in a separate issue with minimal behavior changes.
* Minor updates to existing documentation.
  Completely new documentation should be in a separate issue.

Words "small", "spot", "larger", "minimal", "minor" and similar are defined to be not clearly defined.
As always, use your best judgment and communicate (preferably on the planning).

Engineering tasks could be added to the Sprint only if there are no other pre-planned tasks that are not done
or could not benefit from more help.
Those tasks should not be estimated.

After every Sprint, we calculate how many tasks of each size we were able to complete
and discuss what went well and what could be improved.
We also look at the flows' metrics to gather more information about the team's dynamic.

Estimation mistakes should be an exception, not a norm.
Try to estimate on the planning as best as you can.

## Pull requests

For FerretDB Inc. engineers and long-term contributors, besides the guides in [CONTRIBUTING.md](../CONTRIBUTING.md), please follow these notes below when you working with pull request:

1. Send pull requests from forks; do not make personal branches in the main repository.
   This way, we are similar to community members and could notice similar problems that we could fix for everyone,
   not just for us.
2. Pull request **title** should be accurate and descriptive as it is used in the generated changelog.
   It generally should start with imperative verb ("Fix …", not "Fixing …", "Fixed …" or other forms).
   It should not mention the issue number.
3. We provide a pull requests template that includes suggestions and readiness checklist.
   Please use it.
4. It is fine to send several sequential pull requests for one issue to make them easier to review.
   In that case, please still use the `Closes` word as described above
   (because [words like "refs" do not link PR to the issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword)),
   but don't forget to reopen the issue once PR is merged,
   but the issue as a whole is not done.
5. (This point is for FerretDB Inc. engineers only.)
   Please create a draft pull request as soon as you start working on an issue; that is needed for our process tooling.
   Once it is ready for review, please mark it as such.
   After that, there is no need to switch between draft and non-draft states.
6. Pull requests should be merged by an auto-merge that the author should enable.
   If the pull request is not ready to be merged for some reason (but ready to be reviewed as usual),
   the description should explain why, and the `do not merge` label should be applied.
   Even in that case, auto-merge should still be enabled by the author.
7. Ultimately, the author is responsible for doing everything to ensure that the pull request is merged and done.

One can see all pull requests that await review [there](https://github.com/pulls/review-requested?q=user%3AFerretDB+is%3Aopen).
We also have [#devlog channel](https://ferretdb.slack.com/archives/C02P0MR7VJS)
in our community Slack with periodic reminders.
To make them useful,
talk with [@GitHub bot in direct messages](https://ferretdb.slack.com/archives/D02P4EJPFGV)
to link your Slack and GitHub identities,
and then configure notification only for your name.
