CFv2 RabbitMQ Service Broker
=====================================

This RabbitMQ service broker is a fork based on the great work by [Michal Jemala](https://github.com/michaljemala).

It adds the following further functionality:

 * Uses JSON configuration file instead of command line parameters
 * Sets up [RabbitMQ Federation](http://www.rabbitmq.com/federation.html) Upstream and Policies at Provision stage between arbitrary number of RMQ clusters (see config-example.json) 

Installation
============

To use this service broker, first install it:

```
mkdir ~/go
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH

go get -v github.com/FreightTrain/cf-rabbitmq-broker
```

Configuration
=============


Look at the config-example.json, copy and customize it as appropriate into a location of your choosing.

```
{
    "broker": {
        "host": "0.0.0.0",						Interface to run broker on
        "port": 9998,							Port to run broker on
        "username": "xxx",						Broker auth username
        "password": "yyy"						Broker auth password
        "debug": true,							Enable debug on stdout
        "logFile": "",							File to log output to
        "trace": false,							Enable HTTP API trace output
        "pidFile": ""							Location of broker pid file
    },
    "rabbitmq": {
        "catalog": "",							Not used yet
        "zones": [								Array of Rabbit MQ Clusters
           {
                "name": "dc1",
                "host": "xxx.xxx.xxx.xxx",
                "port": 5672,
                "mgmtHost": "xxx.xxx.xxx.xxx",
                "mgmtPort": 15672,
                "mgmtUser": "xxx",
                "mgmtPass": "xxx",
                "trace": false
            },
            {
                "name": "dc2",
                "host": "yyy.yyy.yyy.yyy",
                "port": 5672,
                "mgmtHost": "yyy.yyy.yyy.yyy",
                "mgmtPort": 15672,
                "mgmtUser": "yyy",
                "mgmtPass": "yyy",
                "trace": false
            }
        ]
    }
}
```

We organized our Rabbit MQ deployment into clusters; one cluster per datacenter. Enabling Federation allows messages to be relayed between clusters for good HA and load balancing. Also, apps running in Cloud Foundry can connect to the RMQ endpoint local to the app, as VCAP_SERVICES will contain a hash of RMQ endpoints, using zone name as the key.


Broker Deployment
=================

You can start the broker locally like this:

```
cf-rabbitmq-broker /path/to/config.json
```

For production use, we deploy the broker into Cloud Foundry itself using [cloudfoundry-buildpack-go](https://github.com/michaljemala/cloudfoundry-buildpack-go). 

Check out this broker's ```cloudfoundrify``` branch and ```cf push``` it somewhere nice.


Usage
=====

Now deploy an app into Cloud Foundry (try this [sample app](http://github.com/michaljemala/rabbitmq-cloudfoundry-samples.git))

```
cf push rabbit-test
cf create-service-broker rabbitmq <broker auth user> <broker auth pass> http://<broker ip>:9999
cf create-service RabbitMQ "Simple RabbitMQ Plan" rabbitmq-simple
```

Now you need to make your newly created service instance public:

```
cf curl /v2/service_plans -X 'GET'
```
```
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "41a1b0c4-3101-4502-bac6-7cddd625d8a5",
        "url": "/v2/service_plans/41a1b0c4-3101-4502-bac6-7cddd625d8a5",
        "created_at": "2014-05-07T20:21:44+00:00",
        "updated_at": "2014-05-07T20:24:56+00:00"
      },
      "entity": {
        "name": "Simple RabbitMQ Plan",
        "free": true,
        "description": "Simple RabbitMQ plan represented as a unique broker's vhost.",
        "service_guid": "ab104ab8-7cc6-4d03-b228-dbf960e09599",
        "extra": null,
        "unique_id": "simple",
        "public": false,
        "service_url": "/v2/services/ab104ab8-7cc6-4d03-b228-dbf960e09599",
        "service_instances_url": "/v2/service_plans/41a1b0c4-3101-4502-bac6-7cddd625d8a5/service_instances"
      }
    }
  ]
}
```

Note the resources.metadata.url and use it in a subsequent cf curl request:

```
cf curl /v2/service_plans/41a1b0c4-3101-4502-bac6-7cddd625d8a5 -X 'PUT' -d '{"public":true}'
```
```
{
    ...
    "public": true,
    ...
}
```

The 'public' property is now set to true. Now you can bind the service instance to your app:

```
cf bind-service rabbit-test rabbitmq-simple
cf push rabbit-test
```
