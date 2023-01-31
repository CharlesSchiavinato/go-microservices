package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var bindAddress = ":9090" // env.String("BIND_ADDRESS", false, ":9090", "Bind address or the server")
var grpcCurrencyTarget = "localhost:9092"
var allowedOrigins = []string{"http://localhost:3000"}

func main() {
	// env.Parse()

	l := log.New(os.Stdout, "working", log.LstdFlags)

	// create the handlers
	hp := handlers.NewProducts(l)

	conn, err := grpc.Dial()

	// create a new serve mux and register the handlers
	sm := mux.NewRouter()

	// handlers for API
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", hp.ProductList)
	getRouter.HandleFunc("/products/{id:[0-9]+}", hp.ProductGet)

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
		ErrorLog:     l,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start the server
	go func() {
		l.Println("Starting server on port 9090")
		err := s.ListenAndServe()

		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	l.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
