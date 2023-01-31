package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductMissingNameReturnsErr(t *testing.T) {
	p := Product{
		Name:  "",
		Price: 1.22,
		SKU:   "aaa-aaa-aaa",
	}

	err := p.Validate()

	assert.Len(t, err, 1)
}

func TestProductMinNameReturnsErr(t *testing.T) {
	p := Product{
		Name:  "aa",
		Price: 1.22,
		SKU:   "aaa-aaa-aaa",
	}

	err := p.Validate()

	assert.Len(t, err, 1)
}
