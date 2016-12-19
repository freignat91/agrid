# AGRID

Agrid v0.1.0 experimental

# Purpose

Agrid is a high available file storage design to be easy to use and scale. Agrid can handle from 1 to several hundred storage nodes. Each file are cut and spread on nodes to be stored.

Agrid is a docker service. It enough to pull its image or build it using `make build` and create the docker service to use it on a swarm cluster. It can be scale using docker service scale command.

Agrid uses grpc protocol for communication between nodes and between nodes and clients. Under 20 nodes, all nodes are completly connected one to eachother, up to 20 nodes agrid create a grid and the communication between nodes are not direct anymore. 

Agrid use Ant like behavior to found the shortest path between two nodes. The path are dynamically adapted regarding the nodes workload to stay the shortest in term of time. The grid become more efficiente while it is used. see: 

- ./docs/Agrid-grid-building.pptx
- ./docs/Agrid-Ant-net.pptx


# Configuration using System Variables:


- GRPCPORT:               grpc server port used by nodes
- NB_DUPLICATE:           number of time a file is replicated in the cluster when it is stored.
- NB_DUPLICATE_ACK:       number of acknoledged replications before a file is concidered stored and acknoledge the client


# Install


- Docker 1.12.3 min should be installed 
- clone this project
- execute `make build` to create a image freignat91/agrid:latest
- start the service: `make start` to create a service agrid using 5 nodes (change make start to exectue direclty the docker create service command to modify startup parameters)


# CLI


- agrid file store [source file pathname] [file pathname in the cluster] <--thread> <--key>
- agrid file get [file pathname in the cluster] [file pathname to write] <--thread>
- agrid file ls [path]
- agrid file rm [pathname] <-r>
- agrid node ls
- agrid node ping


## License

Agrid is licensed under the Apache License, Version 2.0. See https://github.com/freignat91/agrid/blob/master/LICENSE
for the full license text.

