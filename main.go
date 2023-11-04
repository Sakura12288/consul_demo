package main

import (
	"context"
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"os"
	"os/signal"
	"service/config"
	"service/discover"
	"service/endpoint"
	"service/service"
	"service/transport"
	"strconv"
	"syscall"
)

func main() {
	var (
		servicePort = flag.Int("service.port", 10086, "service port")
		serviceHost = flag.String("service.host", "192.168.187.1", "service host")
		serviceName = flag.String("service.name", "SayHello", "service name")

		consulPort = flag.Int("consul.port", 8500, "consul port")
		consulHost = flag.String("consul.host", "127.0.0.1", "consul host")
	)
	flag.Parse()
	ctx := context.Background()

	errChan := make(chan error)

	var discoveryClient discover.DiscoveryClient

	//discoveryClient, err := discover.NewMyDiscoverClient(*consulHost, *consulPort)
	discoveryClient, err := discover.NewKitDiscoverClient(*consulHost, *consulPort)
	if err != nil {
		config.Logger.Println("Get consul client failed" + err.Error())
		os.Exit(-1)
	}

	var svc = service.NewDiscoveryServiceImpl(discoveryClient)

	sayHelloEndpoint := endpoint.MakeSayHelloEndpoint(svc)

	discoverEndpoint := endpoint.MakeDiscoveryEndpoint(svc)

	healthCheckEndpoint := endpoint.MakeHealthCheckEndpoint(svc)

	endpts := endpoint.DiscoveryEndpoints{
		SayHelloEndpoint:    sayHelloEndpoint,
		HealthCheckEndpoint: healthCheckEndpoint,
		DiscoveryEndpoint:   discoverEndpoint,
	}
	r := transport.MakeHttpHandler(ctx, endpts, config.KitLogger)
	instanceID := *serviceName + "-" + uuid.NewV4().String()

	go func() {
		config.Logger.Println("Http Server start at port:" + strconv.Itoa(*servicePort))

		if !discoveryClient.Register(*serviceName, instanceID, "/health",
			*serviceHost, *servicePort, nil, config.Logger) {
			config.Logger.Printf("string-service for service %s failed.", *serviceName)
			os.Exit(1)
		}
		handler := r
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), handler)
	}()

	//监控信号关闭
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	err = <-errChan
	discoveryClient.DeRegister(instanceID, config.Logger)
	config.Logger.Println(err)

}
