package main

import (
	"net"
	"os"

	"github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/data"
	protos "github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/protos/currency"
	"github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/server"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var bindAddress = ":9092"

func main() {
	log := hclog.Default()

	rates, err := data.NewRates(log)

	if err != nil {
		log.Error("Unable to generate rates", "error", err)
		os.Exit(1)
	}

	// create a new gRPC server, use WithInsecure to allow http connections
	gs := grpc.NewServer()

	// create an instance of the Currency server
	c := server.NewCurrency(log, rates)

	// register the currency server
	protos.RegisterCurrencyServer(gs, c)

	// register the reflection service which allows clients to determine the methods
	// for this gRPC service
	reflection.Register(gs)

	// create a TCP socket for inbound server connections
	l, err := net.Listen("tcp", bindAddress)

	if err != nil {
		log.Error("Unable to create listener", "error", err)
		os.Exit(1)
	}

	log.Info("Starting server", "bind_address", bindAddress)

	// listen for requests
	gs.Serve(l)
}
