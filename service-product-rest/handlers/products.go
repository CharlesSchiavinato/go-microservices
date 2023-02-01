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
	"net/http"
	"strconv"

	protos "github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/protos/currency"
	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/data"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

// Products is a http.Handler
type Products struct {
	l         hclog.Logger
	cc        protos.CurrencyClient
	productDB *data.ProductDB
}

// NewProducts creates a products handler with the given logger
func NewProducts(l hclog.Logger, cc protos.CurrencyClient, pdb *data.ProductDB) *Products {
	return &Products{l, cc, pdb}
}

// swagger:route GET /products products ListProducts
// Returns a list of products from the data store
// responses:
// 	200: productsResponse

// ProductList returns all products from the data store
func (p *Products) ProductList(rw http.ResponseWriter, r *http.Request) {
	p.l.Debug("Handle ProductList")

	rw.Header().Add("Content-Type", "application/json")

	// fetch the products from the datastore
	pl, err := p.productDB.ProductList("")

	if err != nil {
		p.l.Error("Handle ProductList - Unable to get currency rate", "error", err)
		http.Error(rw, "Unable to get currency rate", http.StatusInternalServerError)
		return
	}

	// serialize the list to JSON
	err = data.ToJSON(pl, rw)

	if err != nil {
		p.l.Error("Handle ProductList - Unable to serializing product", "error", err)
		http.Error(rw, "Unable to serializing product", http.StatusInternalServerError)
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

	p.l.Debug("Handle ProductGet", "id", id)

	rw.Header().Add("Content-Type", "application/json")

	if err != nil {
		p.l.Error("Handle ProductGet - Invalid id", "id", id, "error", err)
		http.Error(rw, "Invalid id", http.StatusBadRequest)
		return
	}

	pg, err := p.productDB.ProductGetByID(id, "")

	if err == data.ErrProductNotFound {
		p.l.Error("Handle ProductGet - Product not found", "id", id, "error", err)
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		p.l.Error("Handle ProductGet - Internal error", "error", err)
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}

	// get exchange rate
	rr := &protos.RateRequest{
		Base:        protos.Currencies_EUR,
		Destination: protos.Currencies_BRL,
	}
	resp, err := p.cc.GetRate(context.Background(), rr)

	if err != nil {
		p.l.Error("Handle ProductGet - Error getting new rate", "error", err)
		http.Error(rw, "Error getting new rate", http.StatusInternalServerError)
		return
	}

	pg.Price = pg.Price * resp.Rate

	err = data.ToJSON(pg, rw)

	if err != nil {
		p.l.Error("Handle ProductList - Internal error", "error", err)
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
	p.l.Debug("Handle ProductCreate")

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

	p.l.Debug("Handle PUT products", "id", id)

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
		p.l.Error("Handle PUT - Product not found", "id", id, "error", err)
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		p.l.Error("Handle PUT - Internal error", "id", id, "error", err)
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

	p.l.Debug("Handle ProductDelete", "id", id)

	if err != nil {
		p.l.Error("Handle ProductDelete - Invalid id", "id", id, "error", err)
		http.Error(rw, "Invalid id", http.StatusBadRequest)
		return
	}

	err = data.ProductDelete(id)

	if err == data.ErrProductNotFound {
		p.l.Error("Handle ProductDelete - Product not found", "id", id, "error", err)
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		p.l.Error("Handle ProductDelete - Internal error", "id", id, "error", err)
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
			p.l.Error("Handle ProductMiddleware - Deserializing product", "error", err)
			http.Error(rw, "Error reading product", http.StatusBadRequest)
			return
		}

		// validate de product
		err = prb.Validate()

		if err != nil {
			p.l.Error("Handle ProductMiddleware - Validating product", "error", err)
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
