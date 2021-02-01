# symphony

An open-source cloud platform heavily inspired from existing orchestrators (Docker Swarm, Kubernetes, OpenStack, etc).

Developed weekly live at [The Alt-F4 Stream](https://www.twitch.tv/thealtf4stream "The Alt-F4 Stream") on Twitch.

> IMPORTANT: everything below is a work-in-progress and subject to change at any time

## Architecture

Below is an overiew of the current architecture design of Symphony.

Services in orange implemented and gray are unimplemented.

![Overiew - Archecture][preview]

## Concepts

Below describes the basic requirements of a Symphony environment.

- Consul: service discovery and replicated key-value storage of state

- Manager: all internal state changes and messages to cloud services

Below describes the optional cloud services and their functions.

- Block: cloud volume management

## Development

Steps on how to setup a local development environment and perform various tasks.

### Start environment

> NOTE: this does require docker and docker-compose to work

```bash
docker-compose up --build -d
```

### Add a cloud image

Download the cloud image of your choice (Ubuntu 20.04 as an example).

> NOTE: you may execute `image` commands on any host via CLI

```bash
wget https://cloud-images.ubuntu.com/focal/current/focal-server-cloudimg-amd64-disk-kvm.img
```

Build the command-line tools to use locally.

```bash
go build ./cmd/cli
```

Run the `image new` command pointed to your local manager address.

```bash
./cli --manager-addr="localhost:15760" image new "Ubuntu Server 20.04 LTS (Focal Fossa) daily builds" "Ubuntu 20.04" "focal-server-cloudimg-amd64-disk-kvm.img"
```

The image should be registered in Consul at `http://localhost:8500/ui/dc1/kv/image/`.

### Add a cloud service

Verify the new cloud service has the proper `--manager-addr` configuration and start it (block used as example).

> NOTE: you don't need to do this if running the stack locally with docker-compose as it's started already

```bash
block --bind-interface="eth0" --config-dir="/config" --manager-addr="manager:15760" --verbose
```

Then run the `service new` command pointed at the socket path of the service.

> NOTE: you may execute `service` commands only on the service host CLI.

```bash
docker-compose exec block ./cli --socket-path="/config/control.sock" service new
```

The service should be registered healty in Consul at `http://localhost:8500`.

[preview]: overview-architecture.png "Overview - Archecture"
