package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mafewo/mcp-card-kingdom-go/scraper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer("mcp-card-kingdom-go", "1.0.0")

	// Add search_prices tool
	searchTool := mcp.NewTool("search_prices",
		mcp.WithDescription("Search for card prices on Card Kingdom with advanced filters"),
		mcp.WithString("query",
			mcp.Description("The name of the card to search for"),
			mcp.Required(),
		),
		mcp.WithString("set_code",
			mcp.Description("Optional: The set code or edition name to filter by (e.g. 'MH2', 'Commander')"),
		),
		mcp.WithBoolean("is_foil",
			mcp.Description("Optional: Set to true to filter for foil cards only. If omitted, returns both."),
		),
		mcp.WithString("variant",
			mcp.Description("Optional: Filter by variant type (e.g. 'Borderless', 'Etched', 'Showcase', 'Extended Art')"),
		),
	)

	s.AddTool(searchTool, searchPricesHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func searchPricesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid query: %v", err)), nil
	}

	// Extract arguments map
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		// Handle the case where arguments are not a map or are missing
		// For robustness, treat as empty map
		args = make(map[string]interface{})
	}

	var setCode string
	if v, ok := args["set_code"]; ok {
		if s, ok := v.(string); ok {
			setCode = s
		}
	}

	var isFoil bool
	if v, ok := args["is_foil"]; ok {
		if b, ok := v.(bool); ok {
			isFoil = b
		}
	}

	var variant string
	if v, ok := args["variant"]; ok {
		if s, ok := v.(string); ok {
			variant = s
		}
	}

	opts := scraper.SearchOptions{
		Query:   query,
		SetCode: setCode,
		IsFoil:  isFoil,
		Variant: variant,
	}
	
	prices, err := scraper.SearchPrices(opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search prices: %v", err)), nil
	}

	if len(prices) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No prices found for %s", query)), nil
	}

	result := fmt.Sprintf("Prices for %s:\n", query)
	for _, p := range prices {
		foilStatus := ""
		if p.IsFoil {
			foilStatus = " (Foil)"
		}
		result += fmt.Sprintf("- Name: %s%s, Set: %s, Condition: %s, Price: $%.2f, Stock: %d\n", p.Name, foilStatus, p.Edition, p.Condition, p.Price, p.Stock)
	}

	return mcp.NewToolResultText(result), nil
}
