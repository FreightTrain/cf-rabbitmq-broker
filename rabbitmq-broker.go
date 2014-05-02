package main

import (
	"github.com/nimbus-cloud/cf-service-broker/broker"
	"github.com/nimbus-cloud/cf-service-broker/rabbitmq"
	"flag"
	"fmt"
	"log"
	"os"
)

const version = "1.0.0"

func init() {
	flag.BoolVar(&showHelp, "help", false, "")
	flag.BoolVar(&showVersion, "version", false, "")
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if showHelp {
		Usage()
	}
	if showVersion {
		Version()
	}

	brokerService, err := rabbitmq.New(rabbitmq.Opts)
	if err != nil {
		log.Fatal(err)
	}

	broker := broker.New(broker.Opts, brokerService)
	broker.Start()
}

func Usage() {
	fmt.Print(versionStr)
	fmt.Print(broker.UsageStr)
	fmt.Print(rabbitmq.UsageStr)
	fmt.Print(usageStr)
	os.Exit(0)
}
func Version() {
	fmt.Print(versionStr)
	os.Exit(0)
}

var (
	showHelp, showVersion bool
	versionStr            = fmt.Sprintf(`
RabbitMQ Service Broker v%v
`, version)
	usageStr = `
Common Options:
        --help                         Show this message
        --version                      Show service broker version
`
)
