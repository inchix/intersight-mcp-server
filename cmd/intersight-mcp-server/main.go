package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/inchix/intersight-mcp-server/internal/intersight"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	keyID := os.Getenv("INTERSIGHT_API_KEY_ID")
	keyFile := os.Getenv("INTERSIGHT_API_KEY_FILE")
	host := os.Getenv("INTERSIGHT_API_HOST")

	if keyID == "" || keyFile == "" {
		fmt.Fprintf(os.Stderr, "Error: INTERSIGHT_API_KEY_ID and INTERSIGHT_API_KEY_FILE environment variables are required\n")
		os.Exit(1)
	}

	client, err := intersight.NewClient(keyID, keyFile, host)
	if err != nil {
		log.Fatalf("Failed to create Intersight client: %v", err)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "intersight-mcp-server", Version: version},
		nil,
	)

	intersight.RegisterTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
