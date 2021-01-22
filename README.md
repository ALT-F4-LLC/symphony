# symphony

An open-source cloud platform heavily inspired from existing orchestrators (Docker Swarm, Kubernetes, OpenStack, etc).

Developed weekly live at [The Alt-F4 Stream](https://www.twitch.tv/thealtf4stream "The Alt-F4 Stream") on Twitch.

> NOTE: Everything below is a work-in-progress and subject to change at any time.

## Architecture

Below is an overiew of the current architecture design of Symphony.

Services in orange implemented and gray are unimplemented.

![Overiew - Archecture][preview]

## Concepts

Below describes the basic requirements of a Symphony environment.

- Consul: used for service discovery and replicated key-value storage of state.

- APIServer: used for all internal state changes and messages to services.

Below describes the optional cloud services and their functions.

- Block: used for provisioning logical volumes to be used with other resources.

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

[preview]: overview-architecture.png "Overview - Archecture"
