# symphony

An open-source cloud platform written in Go, heavily inspired from existing orchestrators (Docker Swarm, Kubernetes, OpenStack, tc) and developed weekly live at [The Alt-F4 Stream](https://www.google.com "The Alt-F4 Stream") on Twitch.

> NOTE: All documentation below is a WIP (work-in-progress) which is subject to change at any time and is mostly conceptual for development purposes.

## Development

This repository uses `go mod` for versioned modules. You will also need to install the `protoc-gen-go@v1.3.2` package to build grpc files properly with the version of grpc used in this project.

> SEE: https://github.com/grpc/grpc-go/issues/3347

You will also need `docker` and `docker-compose` if you would like to run the local development `etcd` cluster.

## Concepts

Below describes basic concepts of a Symphony environment.

### Manager

Manager nodes maintain all environment state as well as node discovery.

#### Initialization

The following steps happen when a cluster is initialized:

- Manager initalizes in `etcd` for discovery
- If not already a node - will automatically create/join the cluster
- If already a node - will fail request

### Service

Service nodes maintain their resources and state from managers. Worker nodes are not able to directly affect state outside of node discovery.

## Examples

Below shows a simple environment setup process.

#### Start local etcd cluster:

> NOTE: This does require docker and docker-compose to work.

```
$ docker-compose up -d
```

#### Initialize the managers:

```
$ docker-compose exec manager_01 cli --socket="/config/control.sock" manager init
$ docker-compose exec manager_02 cli --socket="/config/control.sock" manager init
$ docker-compose exec manager_03 cli --socket="/config/control.sock" manager init
```

#### Initialize the block service:

```
$ docker-compose exec block_01 cli --socket="/config/control.sock" block init localhost:15760
```

#### Check for services:

```
docker-compose exec manager_01 cli --socket="/config/control.sock" manager service list
```

#### Remove a cloud service:

```
docker-compose exec manager_01 cli --socket="/config/control.sock" manager service remove <service-id>
```
