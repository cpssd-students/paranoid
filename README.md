Paranoid
========
[![GoDoc](https://godoc.org/github.com/pp2p/paranoid?status.svg)](https://godoc.org/github.com/pp2p/paranoid)
[![CircleCI](https://circleci.com/gh/pp2p/paranoid/tree/master.svg?style=shield)](https://circleci.com/gh/pp2p/paranoid/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/pp2p/paranoid)](https://goreportcard.com/report/github.com/pp2p/paranoid)
[![GitHub release](https://img.shields.io/github/release/pp2p/paranoid.svg)](https://github.com/pp2p/paranoid/releases)

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
Main list can be found on the [issues](https://github.com/pp2p/issues) page.
Other than that we are missing:

- OSX Compatibility
- Windows Compatibility
- Modularity (ability to use your own components)
- Hard FUSE depencency - main reason for Linux-only Compatibility
