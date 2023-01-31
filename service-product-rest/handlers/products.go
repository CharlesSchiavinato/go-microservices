// Package classification of Product API
//
// Documentation for Product API
//
//	Schemes: http
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/data"
	"github.com/gorilla/mux"
)

// Products is a http.Handler
type Products struct {
	l *log.Logger
}

// NewProducts creates a products handler with the given logger
func NewProducts(l *log.Logger) *Products {
	return &Products{l}
}

// swagger:route GET /products products ListProducts
// Returns a list of products from the data store
// responses:
// 	200: productsResponse

// ProductList returns all products from the data store
func (p *Products) ProductList(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("[DEGUB] Handle ProductList")

	rw.Header().Add("Content-Type", "application/json")

	// fetch the products from the datastore
	pl := data.ProductList()

	// serialize the list to JSON
	err := data.ToJSON(pl, rw)

	if err != nil {
		p.l.Println("[ERROR] Handle ProductList - Internal error", err)
		http.Error(rw, "Unable to serializing json", http.StatusInternalServerError)
		return
	}
}

// swagger:route GET /products/{id} products GetProduct
// Returns the product from the data store
// responses:
// 	200: productResponse
//  404: errorResponse

// ProductGet returns the product from the data store
func (p *Products) ProductGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	p.l.Println("[DEGUB] Handle ProductGet", id)

	rw.Header().Add("Content-Type", "application/json")

	if err != nil {
		p.l.Println("[ERROR] Handle ProductGet - Invalid id", id)
		http.Error(rw, "Invalid id", http.StatusBadRequest)
		return
	}

	pg, _, err := data.ProductFind(id)

	if err == data.ErrProductNotFound {
		p.l.Println("[ERROR] Handle ProductGet - Product not found", id)
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		p.l.Println("[ERROR] Handle ProductGet - Internal error", err)
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}

	err = data.ToJSON(pg, rw)

	if err != nil {
		p.l.Println("[ERROR] Handle ProductList - Internal error", err)
		http.Error(rw, "Unable to serializing json", http.StatusInternalServerError)
		return
	}
}

// swagger:route POST /products products createProduct
// Create a new product
//
// responses:
//	200: productResponse
//  422: errorValidation
//  501: errorResponse

// ProductCreate requests to add new products
func (p *Products) ProductCreate(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("[DEGUB] Handle ProductCreate")

	prb := r.Context().Value(KeyProduct{}).(*data.Product)

	// p.l.Printf("Product: %#v\n", pa)
	data.ProductAdd(prb)
}

// swagger:route PUT /products products updateProduct
// Update a products details
//
// responses:
//
//	201: noContentResponse
//	404: errorResponse
//	422: errorValidation
func (p *Products) ProductUpdate(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	p.l.Println("Handle PUT products", id)

	if err != nil {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
		return
	}

	rw.Header().Add("Content-Type", "application/json")

	prb := r.Context().Value(KeyProduct{}).(*data.Product)

	prb.ID = id

	// p.l.Printf("Product: %#v\n", pa)
	err = data.ProductUpdate(prb)

	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}
}

// swagger:route DELETE /products/{id} products DeleteProduct
// Delete product
//
// responses:
// 	201: noContentResponse
//  404: errorResponse
// 	501: errorResponse

// Handle ProductDelete delete a product from the data store
func (p *Products) ProductDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	p.l.Println("[DEBUG] Handle ProductDelete", id)

	if err != nil {
		p.l.Println("[ERROR] Handle ProductDelete - Invalid id", id)
		http.Error(rw, "Invalid id", http.StatusBadRequest)
		return
	}

	err = data.ProductDelete(id)

	if err == data.ErrProductNotFound {
		p.l.Println("[ERROR] Handle ProductDelete - Product not found", id)
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		p.l.Println("[ERROR] Handle ProductDelete - Internal error", err)
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

type KeyProduct struct{}

func (p Products) ProductMiddlewareValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prb := &data.Product{}

		err := data.FromJSON(prb, r.Body)

		if err != nil {
			p.l.Println("[ERROR] deserializing product")
			http.Error(rw, "Error reading product", http.StatusBadRequest)
			return
		}

		// validate de product
		err = prb.Validate()

		if err != nil {
			p.l.Println("[ERROR] validating product")
			http.Error(
				rw,
				fmt.Sprintf("Error validating product: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		// p.l.Printf("Product: %#v\n", prb)

		// add the product to the context
		ctx := context.WithValue(r.Context(), KeyProduct{}, prb)
		req := r.WithContext(ctx)

		// call the next handler, which can be another middleware in the chain, or the final handler
		next.ServeHTTP(rw, req)
	})
}
