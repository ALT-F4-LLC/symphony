# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Replaced `etcd` with `consul` for service discovery and state storage.
- Replaced `APIServer` gRPC endpoints with better naming conventions.

## [0.2.0-alpha.0] - 2020-07-10

### Added

- Support for `etcd` in place of own distributed state via custom raft implementation
- Support for block service manager endpoints and CLI commands

### Changed

- Gossip protocol is now used on all services for up-time monitoring

### Removed

- Removed custom raft implementation and integrated with `etcd` for distributed state

## [0.1.0-alpha.0] - 2020-02-07

### Added

- Raft Consensus: managers can now initialize a raft cluster and add or remove existing nodes. Nodes can also fail, restart, and regain state from other members automatically on startup. Manager nodes are also the only nodes in the cluster that can update state in the raft.

- Block Service: block services are able to provision logical volumes using LVM for hosts they run on. Block services are currently able to join the cluster and create physical resources.

- Initial release.
