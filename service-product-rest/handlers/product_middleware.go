package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CharlesSchiavinato/go-microservices/service-product-rest/data"
)

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
