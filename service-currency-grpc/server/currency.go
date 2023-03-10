package server

import (
	"context"

	"github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/data"
	protos "github.com/CharlesSchiavinato/go-microservices/service-currency-grpc/protos/currency"
	"github.com/hashicorp/go-hclog"
)

// Currency is a gRPC server it implements the methods defined by the CurrencyServer interface
type Currency struct {
	log   hclog.Logger
	rates *data.ExchangeRates
}

// mustEmbedUnimplementedCurrencyServer implements currency.CurrencyServer
func (*Currency) mustEmbedUnimplementedCurrencyServer() {
	panic("unimplemented")
}

// NewCurrency creates a new Currency server
func NewCurrency(l hclog.Logger, r *data.ExchangeRates) *Currency {
	return &Currency{l, r}
}

// GetRate implements the CurrencyServer GetRate method and returns the currency exchange rate
// for the two given currencies.
func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle request for GetRate", "base", rr.GetBase(), "dest", rr.GetDestination())

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())

	if err != nil {
		return nil, err
	}

	return &protos.RateResponse{Rate: rate}, nil
}
