package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	protos "github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/protos/currency"
	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/data"
	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

var bindAddress = ":9090" // env.String("BIND_ADDRESS", false, ":9090", "Bind address or the server")
var grpcCurrencyTarget = "localhost:9092"
var allowedOrigins = []string{"http://localhost:3000"}

func main() {
	// env.Parse()

	l := hclog.Default()

	// create currency grpc client
	conn, err := grpc.Dial(grpcCurrencyTarget, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// create client
	cc := protos.NewCurrencyClient(conn)

	// create database instance
	pdb := data.NewProductDB(l, cc)

	// create the handlers
	hp := handlers.NewProducts(l, cc, pdb)

	// create a new serve mux and register the handlers
	sm := mux.NewRouter()

	// handlers for API
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", hp.ProductList)
	getRouter.HandleFunc("/products", hp.ProductList).Queries("currency", "{[A-Z]{3}}")
	getRouter.HandleFunc("/products/{id:[0-9]+}", hp.ProductGet)
	getRouter.HandleFunc("/products/{id:[0-9]+}", hp.ProductGet).Queries("currency", "{[A-Z]{3}}")

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/products/{id:[0-9]+}", hp.ProductUpdate)
	putRouter.Use(hp.ProductMiddlewareValidation)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/products", hp.ProductCreate)
	postRouter.Use(hp.ProductMiddlewareValidation)

	deleteRouter := sm.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/products/{id:[0-9]+}", hp.ProductDelete)

	// handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)

	getRouter.Handle("/docs", sh)
	getRouter.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	//CORS
	ch := gohandlers.CORS(gohandlers.AllowedOrigins(allowedOrigins))

	// create a new server
	s := &http.Server{
		Addr:         bindAddress,
		Handler:      ch(sm),
		ErrorLog:     l.StandardLogger(&hclog.StandardLoggerOptions{}),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start the server
	go func() {
		l.Info("Starting server on port 9090")
		err := s.ListenAndServe()

		if err != nil {
			l.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	l.Info("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
