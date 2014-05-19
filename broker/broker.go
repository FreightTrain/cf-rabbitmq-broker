package broker

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

type broker struct {
	opts   Options
	router *router
}

func New(o Options, bs []BrokerService) *broker {
	return &broker{o, newRouter(o, newHandler(bs))}
}

func (b *broker) Start() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error, 1)
	go func() {
		//appEnv, _ := cfenv.Current()
		//var appHost string
		//var appPort int
		//if appEnv.Host != "" {
		//	appHost = appEnv.Host
		//} else {
		//		appHost = b.opts.Host
		//	}
		//		if appEnv.Port != 0 {
		//			appPort = appEnv.Port
		//		} else {
		//			appPort = b.opts.Port
		//		}
		//		addr := fmt.Sprintf("%v:%v", os.Getenv("HOST")Host, appEnv.Port)
		//		log.Printf("Broker started: Listening at [%v]", addr)
		log.Printf("Broker started: " + os.Getenv("HOST") + ":" + os.Getenv("PORT"))
		errCh <- http.ListenAndServe(":"+os.Getenv("PORT"), b.router)
	}()

	select {
	case err := <-errCh:
		log.Printf("Broker shutdown with error: %v", err)
	case sig := <-sigCh:
		var _ = sig
		log.Print("Broker shutdown gracefully")
	}
}
