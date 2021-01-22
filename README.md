# symphony

An open-source cloud platform heavily inspired from existing orchestrators (Docker Swarm, Kubernetes, OpenStack, etc).

Developed weekly live at [The Alt-F4 Stream](https://www.twitch.tv/thealtf4stream "The Alt-F4 Stream") on Twitch.

> NOTE: Everything below is a work-in-progress and subject to change at any time.

## Concepts

Below describes basic concepts of a Symphony environment.

### Consul

Consul is used for service discovery and replicated key-value storage of state.

### APIServer

APIServer handles all internal state changes and processes messages to services.

### Block

Block service handles provisioning of logical volumes to be used for other resources.

## Development

Steps on how to setup a local development environment and perform various tasks.

### Start environment

> NOTE: This does require docker and docker-compose to work.

```bash
docker-compose up --build -d
```

### Add a cloud service (block used as example)

When adding a cloud service you must execute the commands on the host via CLI.

```bash
docker-compose exec block ./cli --socket-path="/config/control.sock" service new
```

The service should be registered healty in Consul at `http://localhost:8500`.
