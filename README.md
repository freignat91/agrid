# AGRID

Agrid v0.1.4 experimental

# Purpose

Agrid is a high available file storage design to be easy to use and scale. Agrid can handle from 1 to several hundred storage nodes. Each file are cut and spread on nodes to be stored.

Agrid provide a CLI to store and retrieve files and a Go API, this one allows also direct file manipulations create/open, read/write, seek, close,...

Agrid is a docker service. It enough to pull its image `freignat91/agrid:latest` or build it using `make build` and create the docker service to use it on a swarm cluster. It can be scaled using docker service scale command.

Agrid uses grpc protocol for communication between nodes and between nodes and clients. Under 20 nodes, all nodes are completely connected one to each other, up to 20 nodes Agrid create a grid and the communication between nodes are not direct anymore. 

Agrid use Ant like behavior to found the shortest path between two nodes. The path are dynamically adapted regarding the nodes workload to stay the shortest in term of time. The grid become more efficient while it is used. see: 

- ./docs/Agrid-grid-building.pptx
- ./docs/Agrid-Ant-net.pptx


# Configuration using System Variables:


- GRPCPORT:               grpc server port used by nodes
- NB_DUPLICATE:           number of time a file is replicated in the cluster when it is stored: default 3
- NB_DUPLICATE_ACK:       number of acknowledge replications before a file is considered stored and acknowledge the client: default 1
- NB_LINE_CONNECT:        number of "line" type connection in grid: default 0 means computed automatically
- NB_CROSS_CONNECT:       number of "cross" type connection in grid: default 0 means computed automatically
- BUFFER_SIZE:            messages waiting to be treated buffer size: default 10000
- PARALLEL_SENDER:        maximum number of messages to be sent treated in parallel: default 100
- PARALLEL_RECEIVER:      maximum number of received messages treated in parallel: default 100
- DATA_PATH:              path in container where file data is stored: default: /data (should be mapped on host using mount docker argument (--mount type=bind,source=/[hostpath],target=/data)


# Install


- Docker 1.12.3 min should be installed 
- clone this project
- execute `make install` to build the agrid command line executable
- execute `make build` to create a image freignat91/agrid:latest
- start the service: `make start` to create a service agrid using 3 nodes (update Makafile or execute directly the "docker service create" command to change the number of nodes or orher parameters)

for instance with 5 nodes, using a publish port 30103 and network aNetwork

```
docker service create --network aNetwork --name agrid \
        --publish 30103:30103 \
        --mount type=bind,source=/home/freignat/data,target=/data \
        --replicas=5 \
        freignat91/agrid:latest
```

# Resilience

For resilience reason, it's better to have a separated disk file system for each node (each node on its own VM), but for test reason it's possible to use nodes on the same file system or have architecture with several nodes on several VMs.

## Node crash

If a node crash (agrid itself, or disk file system failure or VM failure), docker will restart the node. When the new node restart, it will try to get it's previous file system or ask the other nodes to resend the blocks he handles (this last part is targeted for 0.1.5 version)

## Scale out

To scale out the number of nodes, it's enough to use `docker service scale agrid=xxx` command. Agrid will recompute its grid recreating all the node connections accordingly to the number of nodes

## Scale in

To scale in the number of nodes, it's enough to use `docker service scale agrid=xxx` command. Agrid will recompute its grid recreating all the node connections accordingly to the number of nodes. Warning, to do not lose files, it's important to scale in with a difference between the number of nodes lower than the NB_DUPLICATE parameter and let time to Agrid to reorganize the files blocks between each scale command

## Grid simulation

To simulate nodes connections using different parameters as, node number, line connections, cross connections, use the cli command:

`agrid grid simul [nodes] <--line> <--cross>`
- [nodes] the number of nodes
- <--line> optionally: the number of line connections 
- <--cross> optionally: the number of cross connections 

this command as not effect on the real cluster grid connections.see (./docs/Agrid-grid-building.pptx, ./docs/Agrid-Ant-net.pptx)

## Users

Agrid use a "common" file space by default, everyone can access to this space. Even if files can be encrypted with a specific key, it's possible to create a user. A user create a dedicated file space that no one can access except the user. A user can see and act only on its own file space.
To authenticate a user a token a given at user creation by the cluster, this token should be provided for all commands used with a user.


# CLI

Agrid command lines is implemented using the Agrid Go API

### common options

- --verbose: display more informations messages
- --server: format addr1:port,addr2:port, ...   list of the cluster servers (can list less servers than really in the cluster)

### create a user
`agrid user create [username] <--token>`

Create a user with its own file space in the cluster. This command return a token used to authenticate the user when executing any other command
- [username] the user name to create
- <--token> set the token for this user, without the token is computed by the server

### remove a user
`agrid user remove [username] <--force>`

Remove a user. All files in its file space should have been removed first

- [username] the user name to remove, format userName:token
- <--force> if this option exist then the user is removed with all its associated files, if not the user is removed only if its file space is empty.

### store a file on cluster:

`agrid file store [source] [target] <--thread> <--key> <--user>`
- [source]: the full pathname of the local file to store
- [target]: the full pathname of the file in the cluster
- <--thread number>: optionally: number of threads used to store the file (default 1), each thread open a grpc connection on a distinct node.
- <--key>: optionally: AES key to encrypt the file
- <--user userName:token>: to store on the usee file space, token is given at token creation

by default, the file is stored 3 times (config NB_DUPLICATE parameter default value)


### retrieve a file from cluster

Retrieve a file from cluster using duplicated blocks if some are missing

`agrid file retrieve [source] [target] <--thread> <--key> <--user>`
- [source]: the full pathname of the file to get in cluster
- [target]: the full pathname of the file to write locally
- <--thread>: optionally: number of threads used to retrieve the file (default 1), each thread open a grpc connection on a distinct node.
- <--key>: optionally: AES key to encrypt the file
- <--user userName:token>: to store on the usee file space, token is given at user creation

### list the files on the cluster

`agrid file ls [path] <--user>`
- [path]: path name on the cluster to list, default /
- <--user userName:token>: to store on the usee file space, token is given at user creation


### remove a file on the cluster

`agrid file rm [pathname] <-r> <--user>`
- [pathname]: full pathname of the file to remove on the cluster
- <-r>: to remove a folder recursively
- <--user userName:token>: to store on the usee file space, token is given at user creation

### list the cluster nodes

`agrid node ls`

### ping a cluster node

`agrid node ping |node]`
- [node] the node name to ping


# API

Agrid is usable using Go api API github.com/freignat91/agrid/agridapi

### Usage

```
        import "github.com/freignat91/agrid/agridapi"
        ...
        api := agridapi.New("localhost:10315")
        err, fileList := api.FileLs("/")
        ...
```

### func (api *AgridAPI) userCreate(name string, token string) (string, error)

Create a new user, return a token to authenticate the user
Argument
- name: the user name to create
- token: if equal to "", the token is computed by server, if not it'sused as the user token.

### func (api *AgridAPI) userRemove(name string, force bool) error

Remove a user
Argument
- name: the user name to remove, format userName:token
- force : if true the user is removed with all its associated files, if false the user is removed only if its file space is empty.

### func (api *AgridAPI) SetUser(user string)

Set the current user and authenticate it with the token, then every api function will be executed with this user
Arguements:
- user: format userName:token, or "" to set the common user

### func (api *AgridAPI) FileLs(folder string) ([]string, error)

List the file stored on the cluster
Argument:
- folder: Folder under which the files are listed

### func (api *AgridAPI) FileStore(localFile string, clusterPathname string, meta *[]string, nbThread int, key string) error 

Store a file on the cluster
Arguments:
- localFile: pathname of the local file to store
- clusterPathName: pathname of the file on the cluster
- metadata associated to the file and stored with the file
- nbThread: number of threads used to store the file (each thread open a distinc grpc connection)
- key: AES key to encrypt the file on the cluster 

### func (api *AgridAPI) FileRetrieve(clusterPathname string, localFile string, nbThread int, key string) error

Get a file from the cluster
Arguments:
- clusterPathname: pathname of the file to get on the cluster
- localFile: pathname of the file to write locally
- nbThread: number of threads used to store the file (each thread open a distinc grpc connection)
- key: AES key to decrypt the file

### func (api *AgridAPI) FileRm(clusterPathname string, recursive bool) (error, bool) 

Remove a file on the cluster
Arguments:
- clusterPathname: pathname of the file to remove on the cluster
- recusive: if true remove all files under the clusterPathname

### func (api *AgridAPI) NodePing(node string, debugTrace bool) (string, error)

Ping a node
Arguments:
- node: node name to ping
- debugTrace: if true, trace the message especially in the node logs.

### func (api *AgridAPI) NodePingFromTo(node1 string, node2 string, debugTrace bool) (string, error)
Ping a node from another node
Arguments:
- node1: node name of the node which execute the ping
- node2: targetted node name
- debugTrace: if true, trace the message especially in the node logs.

### func (api *AgridAPI) NodeSetLogLevel(node string, logLevel string) error

Set the logLevel on a node(s)
Arguments:
- node: targetted node
- logLevel: error, warn, info, debug

### func (api *AgridAPI) NodeLs() ([]string, error)

List the node of the cluster

### func (api *AgridAPI) CreateFile(name string, key string) *AFile, error

Create a new file on cluster and return an AFile instance
if key != "", encrypte the data with the key (AES256)

### func (api *AgridAPI) OpenFile(name string, key string) *AFile, error

Open an existing file on cluster, move the current position to the end of file and return an AFile instance
if key != "", encrypte the data with the key (AES256)

## Functions on AFile 

### func (a *AFile) Write(data []byte) (int, error)

Write data in file at the current position

### func (a *AFile) WriteAt(data []byte, at int64) (int, error)

Write data in file at the position 'at'

### func (a *AFile) WriteString(data []byte) (int, error)

Write the string data in file at the current position

### func (a *AFile) WriteStringAt(data []byte, at int64) (int, error)

Write the string data in file at the position 'at'

### func (a *AFile) Read(data []byte) (int, error)

Read data in file at the current position

### func (a *AFile) ReadAt(data []byte, at int64) (int, error)

Read data in file at the position 'at'

### func (a *AFile) Seek(pos int64,  whence int) (int64, error)

Move the current possition to 'pos', relative to whence:
- 0: from the begining of the file
- 1: from the current position
- 2: from the end of the file

### func (a *AFile) Sync() error

Save the memory data on the cluster

### func (a *AFile) Close() error

Save the memory data on the cluster and free all AFile instance ressources

# tests

execute: make test

## License

Agrid is licensed under the Apache License, Version 2.0. See https://github.com/freignat91/agrid/blob/master/LICENSE
for the full license text.
