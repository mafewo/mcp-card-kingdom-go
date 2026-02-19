package scraper

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type CardPrice struct {
	Name      string
	Condition string // NM, EX, VG, G
	Price     float64
	Stock     int
}

func SearchPrices(query string) ([]CardPrice, error) {
	var results []CardPrice
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Selector for each product result
	c.OnHTML(".product-row", func(e *colly.HTMLElement) {
		name := e.ChildText(".productDetailTitle")
		
		// In Card Kingdom, conditions are listed in a <ul> within .itemDetails
		// and the corresponding prices/stocks are in another <ul> or same.
		// Usually: 
		// <ul class="itemDetails"><li>NM</li><li>EX</li>...</ul>
		// <ul class="itemDetails"><li>$10.00 5 available</li>...</ul>
		
		conditions := []string{}
		e.ForEach(".itemDetails:first-child li", func(_ int, el *colly.HTMLElement) {
			conditions = append(conditions, strings.TrimSpace(el.Text))
		})

		e.ForEach(".itemDetails:nth-child(2) li", func(i int, el *colly.HTMLElement) {
			if i >= len(conditions) {
				return
			}
			
			text := strings.TrimSpace(el.Text)
			if text == "" || strings.Contains(text, "Out of stock") {
				return
			}

			// Example text: "$7,999.99 1 available"
			parts := strings.Split(text, " ")
			if len(parts) < 2 {
				return
			}

			priceStr := strings.ReplaceAll(parts[0], "$", "")
			priceStr = strings.ReplaceAll(priceStr, ",", "")
			price, _ := strconv.ParseFloat(priceStr, 64)

			stockStr := parts[1]
			stock, _ := strconv.Atoi(stockStr)

			results = append(results, CardPrice{
				Name:      name,
				Condition: conditions[i],
				Price:     price,
				Stock:     stock,
			})
		})
	})

	searchURL := fmt.Sprintf("https://www.cardkingdom.com/catalog/search?filter%%5Bname%%5D=%s", url.QueryEscape(query))
	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	return results, nil
}
