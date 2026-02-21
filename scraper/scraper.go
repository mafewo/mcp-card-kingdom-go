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
	Edition   string
	Condition string // NM, EX, VG, G
	IsFoil    bool
	Price     float64
	Stock     int
}

type SearchOptions struct {
	Query   string
	SetCode string // Optional: e.g. "MH2"
	IsFoil  bool   // Optional: true to filter only foils
	Variant string // Optional: "Borderless", "Etched", "Showcase", "Extended Art"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SearchPrices(opts SearchOptions) ([]CardPrice, error) {
	var results []CardPrice
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Resolve Set Code to Name if provided
	targetSet := ""
	if opts.SetCode != "" {
		name, err := GetSetName(strings.ToUpper(opts.SetCode))
		if err == nil {
			targetSet = name
		} else {
			// Fallback: use code as is, maybe user passed name?
			targetSet = opts.SetCode
		}
	}
	// fmt.Printf("Target Set: %s\n", targetSet)

	c.OnResponse(func(r *colly.Response) {
		// fmt.Printf("Received response from: %s\n", r.Request.URL)
		// fmt.Printf("Status: %d\n", r.StatusCode)
		// fmt.Printf("Body length: %d\n", len(r.Body))
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		// fmt.Printf("HTML content snippets: %s\n", e.Text[:min(100, len(e.Text))])
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// fmt.Printf("Body class: %s\n", e.Attr("class"))
		// fmt.Printf("Title: %s\n", e.ChildText("title"))
		if strings.Contains(e.Text, "Just a moment") {
			fmt.Println("Cloudflare detected!")
		}
	})

	c.OnHTML(".product-row", func(e *colly.HTMLElement) {
		name := strings.TrimSpace(e.ChildText(".productDetailTitle"))
		edition := strings.TrimSpace(e.ChildText(".productDetailSet"))
		
		// Filter by Set
		if targetSet != "" {
			// Fuzzy match: CK edition names might vary slightly
			if !strings.Contains(strings.ToLower(edition), strings.ToLower(targetSet)) {
				return 
			}
		}

		// Detect Foil
		isFoil := strings.Contains(strings.ToLower(name), "foil") || strings.Contains(strings.ToLower(edition), "foil")
		if opts.IsFoil && !isFoil {
			return
		}
		// If user specifically DOES NOT want foil, we should filter foils out?
		// Usually search implies "any" unless specified. 
		// If strict non-foil is needed, we'd need another flag. 
		// For now, let's assume IsFoil=false means "don't care" or "prefer non-foil but show all".
		// Actually, standard behavior is usually mixed unless filtered.

		// Filter by Variant
		if opts.Variant != "" {
			// Check if name contains variant text (e.g. "Borderless")
			if !strings.Contains(strings.ToLower(name), strings.ToLower(opts.Variant)) {
				return
			}
		}

		// Parse Conditions & Prices
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
				Edition:   edition,
				Condition: conditions[i],
				IsFoil:    isFoil,
				Price:     price,
				Stock:     stock,
			})
		})
	})

	// Build URL with filters
	baseURL := "https://www.cardkingdom.com/catalog/search"
	v := url.Values{}
	v.Set("filter[name]", opts.Query)
	
	if opts.IsFoil {
		v.Set("filter[foil]", "1")
	}

	searchURL := fmt.Sprintf("%s?%s", baseURL, v.Encode())
	// fmt.Printf("Visiting: %s\n", searchURL)
	
	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	return results, nil
}
