# symphony

An open-source cloud platform written in Go, heavily inspired from Docker Swarm and developed weekly live at The Alt-F4 Stream on Twitch.

> NOTE: All documentation below is a WIP (work-in-progress) which is subject to change at any time and is mostly conceptual for development purposes.

## Concepts

Below describes basic concepts of a Symphony cluster.

### I. Manager

Manager nodes maintain all raft state as well as node discovery.

#### Manager Initialization

The following steps happen when a cluster is initialized:

- First manager generates `key.json` to store `raft_node_id`
- First manager generates `manager` and `worker` join tokens

### II. Worker

Worker nodes maintain their resources from the raft state and join/leave the cluster via manager nodes.

> NOTE: Worker nodes are not able to directly affect state outside of node discovery.
