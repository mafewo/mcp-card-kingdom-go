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
		mcp.WithDescription("Search for card prices on Card Kingdom"),
		mcp.WithString("query",
			mcp.Description("The name of the card to search for"),
			mcp.Required(),
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
	
	prices, err := scraper.SearchPrices(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search prices: %v", err)), nil
	}

	if len(prices) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No prices found for %s", query)), nil
	}

	result := fmt.Sprintf("Prices for %s:\n", query)
	for _, p := range prices {
		result += fmt.Sprintf("- Name: %s, Condition: %s, Price: $%.2f, Stock: %d\n", p.Name, p.Condition, p.Price, p.Stock)
	}

	return mcp.NewToolResultText(result), nil
}
