Cloud Foundry RabbitMQ Service Broker
=====================================

To use this broker, first install it:

```
mkdir ~/go
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH

go get -v github.com/cloudfoundry-incubator/spiff
```

Start it somewhere, paying attention to the mandatory CLI parameters:

```
cf-rabbitmq-broker -bu <broker auth user> \
                   -bp <broker auth pwd> \
                   -rh <consumer RMQ IP> \
                   -rmh <admin RMQ IP> \
                   -rmp <admin RMQ pass> \
                   -rmu <admin RMQ user> \
                   -D \
                   -V \
                   -R
```

Now deploy an app into Cloud Foundry (try the sample one here: http://github.com/michaljemala/rabbitmq-cloudfoundry-samples.git)

```
cf push rabbit-test
cf create-service-broker rabbitmq <broker auth user> <broker auth pass> http://<broker ip>:9999
cf create-service RabbitMQ "Simple RabbitMQ Plan" rabbitmq-simple
cf bind-service rabbit-test rabbitmq-simple
cf push
```
