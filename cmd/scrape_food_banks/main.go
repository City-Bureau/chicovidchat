package main

import (
	"net/http"
	"os"

	"github.com/City-Bureau/chicovidchat/pkg/directory"
	"github.com/City-Bureau/chicovidchat/pkg/scrapers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.CloudWatchEvent) error {
	response, err := http.Get("https://www.chicagosfoodbank.org/find-food/")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	resources, scraperErr := scrapers.ScrapeFoodBanks(response.Body)
	if scraperErr != nil {
		panic(scraperErr)
	}

	airtableBase := os.Getenv("AIRTABLE_BASE")
	airtableTable := os.Getenv("AIRTABLE_TABLE")
	airtableKey := os.Getenv("AIRTABLE_KEY")
	return directory.SyncAirtableResources(resources, airtableBase, airtableTable, airtableKey)
}

func main() {
	lambda.Start(handler)
}
