Writing Code for Paranoid
=========================

## Code Formatting ##

All `go` code checked into the repository must be formatted according to `go fmt`.
The imports are using `goimport` to manage order and to separate stdlib from
custom packages. To lint your code for style errors, use [golint](https://github.com/golang/lint). All code must also be
checked by the `go vet` command, which will report valid but poor code. This is separate
from `golint`, which report style errors.

## Code Style ##

Go doesn't have an official style guide as such, but a list of common mistakes can be
found [here](https://github.com/golang/go/wiki/CodeReviewComments). Also worth a look is
[Effective Go](https://golang.org/doc/effective_go.html), a slightly more advanced guide
to writing Go.

## Commenting ##

Commenting will be done in accordance with the standard Go style used for generating Godocs.
You should write comments to explain any code you write that does something in a way that may not be obvious.
An explaination of standard Go commenting practice and Godoc can be found [here](https://blog.golang.org/godoc-documenting-go-code).

## Automated Testing ##

Automated Testing will be done using the [testing](https://golang.org/pkg/testing/) Go package.
Unit tests and integration tests should be kept in separate files that are run in separate Jenkins projects.
Files containing integration tests should be marked by starting with the comment `// +build integration`
Files containing unit tests should be marked by starting with the comment `// +build !integration`
Tests will be automatically ran for each pull request and pull requests will failling tests will not be merged.

## Forking ##

In order to contribute to the project you must create a fork. The easiest way
is to do the following:

1. Fork the project using Github's web UI.
2. Clone your fork into `$GOPATH/src/github.com/cpssd-students/paranoid`
3. Add `upstream` remote (`git remote add upstream git@github.com:cpssd-students/paranoid`)

## Branching ##

Branches should have a short prefix describing what type of changes are contained in it.
These prefixes should be one of the following:

* **feature/** -- for changes which add/affect a feature.
* **doc/** -- for changes to the documentation.
* **hotfix/** -- for quick bugfixes.

Make sure to `git fetch` and merge into `master` often to prevent unnecessary
commits.

## Version Numbering ##

We will be using [semantic versioning](http://semver.org/). Every binary will have a separate
version number (server, client, etc.). This will begin at 0.1.0 and will be incremented strictly
according to the semantic versioning guidelines.

## Code Review ##

All code must be submitted via pull requests, *not* checked straight into the repo.
A pull request will be from a single feature branch, which will not be used for any other
features after merging. Ideally, it will be deleted after a merge.

All pull requests must be reviewed by at least one person who is not the submitter. You can
ask in Hipchat for a review, or use Github's assign feature to assign the pull request to a
specific person.
