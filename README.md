# Gaffer 

Gaffer is a supervisor process and service discovery tool for running an [Apache Mesos](https://mesos.apache.org) cluster. It's job is to manage the configuration of the `mesos-master`, `mesos-agent`, and `zookeeper` process and ensure they continue running. Since Mesos is dependent on Zookeeper Gaffer provides similar functionality to [Exhibitor](https://github.com/soabase/exhibitor/wiki).

## Gaffer-http

`gaffer-http` is an HTTP server which stores configuration data for running an Apache Mesos cluster. It's design is similar to etcd's [discovery service protocol](https://github.com/coreos/etcd/blob/master/Documentation/dev-internal/discovery_protocol.md). It can be deployed internally or you may consider the hosted [mesos.co](http://mesos.co) discovery service.

### Endpoints

#### /1/<user-id>/clusters

      [
        {
          "name": "my-cluster",
          "initialized": true,
          "zookeeper_ready": true,
          "mesos_ready": true,
          "size": 3,
          "zookeeper_config": [],
          "master_options": [
            {
              "name": "MESOS_LOGGING_LEVEL",
              "value": "WARNING"
            },
            {
              "name": "--credentials",
              "value": "file:///etc/credentials.json"
            },  
              -- or --
            {
              "name": "--credentials",
              "value": null,
              "data": {  -- arbitrary JSON data will be written to a temporary file and passed to the flag
                "credentials": [
                  {
                    "principal": "foo",
                    "secret": "bar"
                  }
                ]
              }
            }
          ],
          "agent_config": [...],
          "zookeepers": [
            {
              "hostname": "master-1",
              "ip": "192.168.0.10",
              "running": true,
              "options": [
                {
                  "name": "server.1",
                  "value": "master-1:2888,3888"
                },
                {
                  "name": "syncLimit",
                  "value": "2"
                }
              ]
            }
            {
              "hostname": "master-2",
              "ip": "192.168.0.11",
              "running": true
            }
            {
              "hostname": "master-3",
              "ip": "192.168.0.12",
              "running": true
            }
          ]
        },
        "masters": [
            {
              "hostname": "master-1",
              "ip": "192.168.0.10",
              "running": true,
              "config": {
                "options": [
                  {
                    "name": "MESOS_ZK",
                    "value": "zk://master-1:2181,master-2:2181,master-3:2181/mesos"
                  },
                  {
                    "name": "MESOS_LOGGING_LEVEL",
                    "value": "WARNING"
                  },
                ]
              }
            },
            {
              "hostname": "master-2",
              "ip": "192.168.0.11",
              "running": true
            },
            {
              "hostname": "master-3",
              "ip": "192.168.0.12",
              "running": true
            },
          ]
        "agents": [
          {
            "hostname": "agent-1",
            "ip": "192.168.1.10",
            "running": true
            "config": {...}
          },
          {
            "hostname": "agent-2",
            "ip": "192.168.1.11",
            "running": true
            "config": {...}
          }
        ]
      ]

#### /1/<user-id>/create

The create endpoint creates a new uninitialized cluster with sensible defaults and returns a unique cluster token. The configuration may be modified with any additional configuration once created.

      {
        "size" 3
      }

      {
        "token": "55518bbf-cd7d-4f98-8488-faedcbdfa5a4"
      }

#### /1/<user-id>/<token>/initialize

Until the cluster is initialized all Gaffer CLI processes will poll the endpoint and refuse to load any configuration. Once initialized the cluster must go through two distinct stages before agents
can join. 1) Zookeepers must join one another 2) Mesos Masters must join together.


## Gaffer

`gaffer` is the supervisor which launches and maintains the Apache Mesos process. Gaffer supports all of the normal `MESOS_` style environment variables. Flags can be passed to the mesos process by specifying `-mesos-flags="--zk ..."` or by setting `GAFFER_MESOS_FLAGS="--zk ..."`. If `-endpoint` and `-token` are specified Gaffer will configure Mesos based on the remote configuration. Locally specified configuration will always take precedence over the remote configuration.

### Usage

    gaffer -endpoint mesos.co -user-id 1234 -token 55518bbf-cd7d-4f98-8488-faedcbdfa5a4 -mode master
    ...
