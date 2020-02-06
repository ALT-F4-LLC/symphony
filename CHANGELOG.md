# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Raft Consensus: managers can now initialize a raft cluster and add or remove existing nodes. Nodes can also fail, restart, and regain state from other members automtaically on startup. Manager nodes are also the only nodes in the cluster that can update state in the raft.
