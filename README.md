# Credential Management for [oras-go](https://github.com/oras-project/oras-go)

> **Warning** This project is currently under initial development. APIs may and will be changed incompatibly from one commit to another.

[![Build Status](https://github.com/oras-project/oras-credentials-go/actions/workflows/build.yml/badge.svg?event=push&branch=main)](https://github.com/oras-project/oras-credentials-go/actions/workflows/build.yml?query=workflow%3Abuild+event%3Apush+branch%3Amain)
[![codecov](https://codecov.io/gh/oras-project/oras-credentials-go/branch/main/graph/badge.svg)](https://codecov.io/gh/oras-project/oras-credentials-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/oras-project/oras-credentials-go)](https://goreportcard.com/report/github.com/oras-project/oras-credentials-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/oras-project/oras-credentials-go.svg)](https://pkg.go.dev/github.com/oras-project/oras-credentials-go)

<p align="left">
<a href="https://oras.land/"><img src="https://oras.land/img/oras.svg" alt="banner" width="100px"></a>
</p>

`oras-credentials-go` is a credential management library designed for [`oras-go`](https://github.com/oras-project/oras-go). It supports reading, saving, and removing credentials from Docker configuration files and external credential stores that follow the [Docker credential helper protocol](https://docs.docker.com/engine/reference/commandline/login/#credential-helper-protocol), while not handling credential encryption and decryption. 

Once stable, this library will be merged into `oras-go`.

## Versioning

The `oras-credentials-go` library follows [Semantic Versioning](https://semver.org/), where breaking changes are reserved for MAJOR releases, and MINOR and PATCH releases must be 100% backwards compatible.

## Docs

- [oras-go](https://github.com/oras-project/oras-go): Source code of the ORAS Go library
- [oras.land/docs/Client_Libraries/go](https://oras.land/docs/Client_Libraries/go): Documentation for the ORAS Go library
- [Reviewing guide](https://github.com/oras-project/community/blob/main/REVIEWING.md): All reviewers must read the reviewing guide and agree to follow the project review guidelines.

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
