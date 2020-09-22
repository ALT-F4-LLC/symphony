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

How To Achieve Simple Tasks:

1. How will manager service maintain state:

   - We will use etcd (distributed key/value store) that saves all cloud state and replicates it across hosts.
   - We will distribute the cloud state across multiple hosts and direct requests to current leader.

1. Create a logical volume:

   - Provision block service
   - Configure block service to have a static variable for the manager host address
   - Create a new logical volume on manager service
     - Scheduler notices the unassigned volume and schedules it for provisioning
     - Scheduler notifies manager of where to place volume in cloud
     - Manager assigns the volume to the block host
     - Block host notices the changes and updates the resources accordingly

1. Copy/write the OS to that volume:

   - Provision image service
   - Configure image service to have a static variable for the manager host address
   - Create a new logical root volume on manager service
     - Do all previous steps before
     - Include a OS value to also copy/write to volume
     - Prompt block service to download image and "dd" image to volume

1. Create the virtual machine:

   - Provision compute service
   - Configure compute service to have a static variable for the manager host address
   - Create a new virtual machine on manager service
     - Create logical root volume on manager (as per 2)
     - Create logical additional volume(s) on manager (as per 1)
     * Create virtual networks for your machine
     - Assign the logical volumes to the compute host
     - Assign the virtual networks to the compute host
     - Compute host notice the changes and connect/mount the volume
