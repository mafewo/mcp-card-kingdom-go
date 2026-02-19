package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mafewo/mcp-card-kingdom-go/scraper"
	"github.com/mark3labs/mcp-go"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer("mcp-card-kingdom-go", "1.0.0")

	// Add search_prices tool
	searchTool := mcp.NewTool("search_prices",
		mcp.WithDescription("Search for card prices on Card Kingdom"),
		mcp.WithArgument("query",
			mcp.SetDescription("The name of the card to search for"),
			mcp.SetRequired(true),
		),
	)

	s.AddTool(searchTool, searchPricesHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func searchPricesHandler(args map[string]interface{}) (*mcp.CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return mcp.NewToolResultError("Invalid query"), nil
	}
	
	prices, err := scraper.SearchPrices(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search prices: %v", err)), nil
	}

	result := fmt.Sprintf("Prices for %s:\n", query)
	for _, p := range prices {
		result += fmt.Sprintf("- Condition: %s, Price: $%.2f, Stock: %d\n", p.Condition, p.Price, p.Stock)
	}

	return mcp.NewToolResultText(result), nil
}
