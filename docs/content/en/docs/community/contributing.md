---
title: "Contributing Guidelines"
weight: 20
description: >
  How to best contribute to the project
---

Thank you for your interest in contributing to our project. Whether it's a bug report, new feature, correction, or additional
documentation, we greatly value feedback and contributions from our community.

Please read through this document before submitting any issues or pull requests to ensure we have all the necessary
information to effectively respond to your bug report or contribution.

## General Guidelines

### Pull Requests

Make sure to keep Pull Requests small and functional to make them easier to review, understand, and look up in commit history.
This repository uses "Squash and Commit" to keep our history clean and make it easier to revert changes based on PR.

Adding the appropriate documentation, unit tests and e2e tests as part of a feature is the responsibility of the
feature owner, whether it is done in the same Pull Request or not.

Pull Requests should follow the "subject: message" format, where the subject describes what part of the code is being
modified.

Refer to the [template](https://github.com/aws/eks-anywhere/blob/main/.github/PULL_REQUEST_TEMPLATE.md) for more information on what goes into a PR description.

### Design Docs

A contributor proposes a design with a PR on the repository to allow for revisions and discussions.
If a design needs to be discussed before formulating a document for it, make use of GitHub Discussions to
involve the community on the discussion.

### GitHub Discussions

GitHub Discussions are used for feature requests (that don't have actionable items/issues), questions, and anything else
the community would like to share.

Categories:
* Q/A - Questions
* Proposals - Feature requests and other suggestions
* Show and tell - Anything that the community would like to share
* General - Everything else (possibly announcements as well)

### GitHub Issues

GitHub Issues are used to file bugs, work items, and feature requests with actionable items/issues (Please refer to the
"Reporting Bugs/Feature Requests" section below for more information).

Labels:
* "\<area\>" - area of project that issue is related to (create, upgrade, flux, test, etc.)
* "priority/p\<n\>" - priority of task based on following numbers
  * p0: need to do right away
  * p1: don't have a set time but need to do
  * p2: not currently being tracked (backlog)
* "status/\<status\>" - status of the issue (notstarted, implementation, etc.)
* "kind/\<kind\>" - type of issue (bug, feature, enhancement, docs, etc.)

Refer to the [template](https://github.com/aws/eks-anywhere/tree/main/.github/ISSUE_TEMPLATE) for more information on
what goes into an issue description.

### GitHub Milestones

GitHub Milestones are used to plan work that is currently being tracked.

* next: changes for next release
* next+1: won't make next release but the following
* techdebt: used to keep track of techdebt items, separate ongoing effort from release action items
* oncall: used to keep track of issues needing active follow-up
* backlog: items that don't have a home in the others

### GitHub Projects (or tasks within a GitHub Issue)

GitHub Projects are used to keep track of bigger features that are made up of a collection of issues.
Certain features can also have a tracking issue that contains a checklist of tasks that
link to other issues.

## Reporting Bugs/Feature Requests

We welcome you to use the GitHub issue tracker to report bugs or suggest features that have actionable items/issues
(as opposed to introducing a feature request on GitHub Discussions).

When filing an issue, please check existing open, or recently closed, issues to make sure somebody else hasn't already
reported the issue. Please try to include as much information as you can. Details like these are incredibly useful:

* A reproducible test case or series of steps
* The version of the code being used
* Any modifications you've made relevant to the bug
* Anything unusual about your environment or deployment

## Contributing via Pull Requests
Contributions via pull requests are much appreciated. Before sending us a pull request, please ensure that:

1. You are working against the latest source on the *main* branch.
1. You check existing open, and recently merged, pull requests to make sure someone else hasn't addressed the problem already.
1. You open an issue to discuss any significant work - we would hate for your time to be wasted.

To send us a pull request, please:

1. Fork the repository.
1. Modify the source; please focus on the specific change you are contributing. If you also reformat all the code, it
   will be hard for us to focus on your change.
1. Ensure local tests pass.
1. Commit to your fork using clear commit messages.
1. Send us a pull request, answering any default questions in the pull request interface.
1. Pay attention to any automated CI failures reported in the pull request, and stay involved in the conversation.

GitHub provides additional document on [forking a repository](https://help.github.com/articles/fork-a-repo/) and
[creating a pull request](https://help.github.com/articles/creating-a-pull-request/).


## Finding contributions to work on
Looking at the existing issues is a great way to find something to contribute on. As our projects, by default, use the
default GitHub issue labels (enhancement/bug/duplicate/help wanted/invalid/question/wontfix), looking at any 'help wanted'
and 'good first issue' issues are a great place to start.


## Code of Conduct
This project has adopted the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct).
For more information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq) or contact
opensource-codeofconduct@amazon.com with any additional questions or comments.


## Security issue notifications
If you discover a potential security issue in this project we ask that you notify AWS/Amazon Security via our
[vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/). Please do **not** create a
public GitHub issue.


## Licensing

See the [LICENSE](https://github.com/aws/eks-anywhere/blob/main/LICENSE) file for our project's licensing. We will ask you to confirm the licensing of your contribution.
