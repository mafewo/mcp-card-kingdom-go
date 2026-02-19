package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
)

type CardPrice struct {
	Name      string
	Condition string // NM, EX, VG, G
	Price     float64
	Stock     int
}

func SearchPrices(query string) ([]CardPrice, error) {
	// Scraper implementation using colly
	return nil, fmt.Errorf("Not implemented")
}
