package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

func main() {
	// Azure Cosmos DB connection string
	connectionString := "AccountEndpoint=https://cdbazewp-admiral-core.documents.azure.com:443/;AccountKey=ZZ5NJmqjAobGMI8w1MHbO4LN4oigpQnM0xwn0q40B9uoHYFqlgRvM6sEhdGv2UxpnJ0nqI1yLRuw2jc5Jx0s4w==;"

	// Parse the connection string and initialize Cosmos DB client
	client, err := azcosmos.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Fatalf("Failed to create Cosmos client from connection string: %v", err)
	}

	// Database and container details
	databaseName := "Admiral"
	containerName := "Support"

	// Get the container reference
	container, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		log.Fatalf("Failed to get container client: %v", err)
	}

	// Query to fetch ticket details created in the last 90 days
	detailsQuery := `
		SELECT * FROM c 
		WHERE c.CreatedDate <= GetCurrentDateTime() 
		AND c.CreatedDate >= DateTimeAdd("day", -90, GetCurrentDateTime())
	`

	// Execute the query
	queryPager := container.NewQueryItemsPager(detailsQuery, azcosmos.PartitionKey{}, nil)

	var tickets []map[string]interface{}
	ctx := context.TODO()

	for queryPager.More() {
		resp, err := queryPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("Failed to query items: %v", err)
		}

		for _, item := range resp.Items {
			var ticket map[string]interface{}
			if err := json.Unmarshal(item, &ticket); err != nil {
				log.Printf("Failed to unmarshal item: %v", err)
				continue
			}
			tickets = append(tickets, ticket)
		}
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found created in the last 90 days.")
		return
	}

	// Save to JSON file
	jsonOutputFile := "support_tickets_last_90_days.json"
	if err := saveToJSON(jsonOutputFile, tickets); err != nil {
		log.Fatalf("Failed to save tickets to JSON file: %v", err)
	}
	fmt.Printf("Ticket details saved to %s\n", jsonOutputFile)

	// Save to Markdown file
	mdOutputFile := "support_tickets_last_90_days.md"
	if err := saveToMarkdown(mdOutputFile, tickets); err != nil {
		log.Fatalf("Failed to save tickets to Markdown file: %v", err)
	}
	fmt.Printf("Ticket details saved to %s\n", mdOutputFile)
}

// saveToJSON saves data to a JSON file
func saveToJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// saveToMarkdown saves data to a Markdown file
func saveToMarkdown(filename string, tickets []map[string]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write Markdown content
	file.WriteString("# Support Tickets (Last 90 Days)\n\n")
	for _, ticket := range tickets {
		file.WriteString(fmt.Sprintf("## Ticket ID: %v\n", ticket["id"]))
		file.WriteString(fmt.Sprintf("- **Created Date:** %v\n", ticket["CreatedDate"]))
		file.WriteString(fmt.Sprintf("- **Details:** %v\n\n", ticket["Data"]))
	}
	return nil
}
