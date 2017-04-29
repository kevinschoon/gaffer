# Gaffer

Gaffer is a tool for provisioning and orchestrating an Apache Mesos cluster.

## Installation

    go get github.com/vektorlab/gaffer
    cd $GOPATH/src/github.com/vektorlab/gaffer
    make

## Usage

    gaffer [OPTIONS] -mode MODE

### Modes of Operation

#### server

Server mode runs an HTTP process for serving cluster configuration and the
Gaffer management UI.

#### supervisor

Supervisor mode launches one or more of the core Mesos cluster components.
When running in supervisor mode you must provide an API token and endpoint
of a Gaffer server.

##### zookeeper

Launch an instance of Zookeeper and configure it based on an existing
cluster configuration.

##### master

Launch a Mesos Master process and configure it based on an existing
cluster configuration.

##### agent

Launch a Mesos Agent process and configure it based on an existing
cluster configuration.
