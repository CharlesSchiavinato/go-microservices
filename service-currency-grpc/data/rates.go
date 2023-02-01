package data

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/go-hclog"
)

type ExchangeRates struct {
	log   hclog.Logger
	rates map[string]float64
}

func NewRates(l hclog.Logger) (*ExchangeRates, error) {
	er := &ExchangeRates{log: l, rates: map[string]float64{}}

	er.getRatesECB()

	return er, nil
}

func (e *ExchangeRates) GetRate(base, dest string) (float64, error) {
	br, ok := e.rates[base]

	if !ok {
		return 0, fmt.Errorf("Rate not found for base currency %s", base)
	}

	dr, ok := e.rates[dest]

	if !ok {
		return 0, fmt.Errorf("Rate not found for destination currency %s", dest)
	}

	return dr / br, nil
}

func (e *ExchangeRates) getRatesECB() error {
	resp, err := http.DefaultClient.Get("https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")

	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid Status Code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	md := &Cubes{}
	xml.NewDecoder(resp.Body).Decode(md)

	for _, cd := range md.CubeData {
		r, err := strconv.ParseFloat(cd.Rate, 64)

		if err != nil {
			return err
		}

		e.rates[cd.Currency] = r
	}

	e.rates["EUR"] = 1

	return nil
}

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
