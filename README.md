Paranoid
========
[![GoDoc](https://godoc.org/github.com/cpssd-students/paranoid?status.svg)](https://godoc.org/github.com/cpssd-students/paranoid)
[![Go](https://github.com/cpssd-students/paranoid/actions/workflows/go.yaml/badge.svg)](https://github.com/cpssd-students/paranoid/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cpssd-students/paranoid)](https://goreportcard.com/report/github.com/cpssd-students/paranoid)
[![GitHub release](https://img.shields.io/github/release/cpssd-students/paranoid.svg)](https://github.com/cpssd-students/paranoid/releases)

Distributed Secure Peer-to-Peer filesystem

---

__WARNING: This repository was not maintained for a while. It has a lot of
problems, probably doesn't work in lot of places. If you want to help out
please consult CONTRIBUTING.md and/or create an issue.__

## Building
Simplest way to get everything is `go get ./...`. This should download all
Go dependencies. The project also requires `fuse`, so make sure you have it
installed and you are in the `fuse` group.

To build a specific binary, consult the README file for that directory.

---

## Testing
To run the unit tests recursively for the entire project, run `go test ./...` from this directory.

To run the integration tests recursively for the entire project run `go test ./... -tags=integration` from this directory.

---

## Known Issues:
Main list can be found on the [issues](https://github.com/cpssd-students/issues) page.
Other than that we are missing:

- __Doesn't even build right now.__
- OSX Compatibility
- Windows Compatibility
- Modularity (ability to use your own components)
- Hard FUSE depencency - main reason for Linux-only Compatibility
