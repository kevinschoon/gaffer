# Gaffer

[![CircleCI](https://img.shields.io/circleci/project/github/mesanine/gaffer.svg)]()

Gaffer is the init system and process manager for [Mesanine](https://github.com/mesanine). It is designed to control important system and user processes by launching 
containers with [Runc](https://github.com/opencontainers/runc). Gaffer aims to be compatible with [Linuxkit](https://github.com/linuxkit/linuxkit) which is used to 
build Mesanine. While components in Linuxkit have generally very discrete functions, Gaffer takes a more encompassing and opinionated approach to system configuration.


## Building

    go get github.com/mesanine/gaffer
    cd $GOPATH/src/github.com/mesanine/gaffer
    make docker

## Extending

Everything in Gaffer is a plugin. Plugins emit and respond to events across a shared `EventBus`. To add new features you Gaffer you must implement the `Plugin` interface and register it when Gaffer is initialized.
