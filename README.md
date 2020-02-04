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

## Example Configurations

# Single node manager cluster

rm -rf .raft/manager-01 && go run ./cmd/manager --config-dir .raft/manager-01

go run ./cmd/cli --socket ".raft/manager-01/control.sock" init

# Three node cluster

rm -rf .raft/manager-01 && go run ./cmd/manager --config-dir .raft/manager-01

rm -rf .raft/manager-02 && go run ./cmd/manager --config-dir .raft/manager-02 \
--listen-raft-addr 127.0.0.1:15761 \
--listen-remote-addr 127.0.0.1:27243

rm -rf .raft/manager-03 && go run ./cmd/manager --config-dir .raft/manager-03 \
--listen-raft-addr 127.0.0.1:15762 \
--listen-remote-addr 127.0.0.1:27244

go run ./cmd/cli --socket ".raft/manager-01/control.sock" init --peers "http://127.0.0.1:15761,http://127.0.0.1:15762"
go run ./cmd/cli --socket ".raft/manager-02/control.sock" init --join-addr 127.0.0.1:27242
go run ./cmd/cli --socket ".raft/manager-03/control.sock" init --join-addr 127.0.0.1:27242

# Add an additional node to the three node cluster

rm -rf .raft/manager-04 && go run ./cmd/manager --config-dir .raft/manager-04 \
--listen-raft-addr 127.0.0.1:15763 \
--listen-remote-addr 127.0.0.1:27245

rm -rf .raft/manager-05 && go run ./cmd/manager --config-dir .raft/manager-05 \
--listen-raft-addr 127.0.0.1:15764 \
--listen-remote-addr 127.0.0.1:27246

rm -rf .raft/manager-06 && go run ./cmd/manager --config-dir .raft/manager-06 \
--listen-raft-addr 127.0.0.1:15765 \
--listen-remote-addr 127.0.0.1:27247

go run ./cmd/cli --socket ".raft/manager-04/control.sock" join 127.0.0.1:27242
go run ./cmd/cli --socket ".raft/manager-05/control.sock" join 127.0.0.1:27242
go run ./cmd/cli --socket ".raft/manager-06/control.sock" join 127.0.0.1:27242

go run ./cmd/cli --socket ".raft/manager-01/control.sock" init
go run ./cmd/cli --socket ".raft/manager-02/control.sock" join 127.0.0.1:27242
go run ./cmd/cli --socket ".raft/manager-03/control.sock" join 127.0.0.1:27242
