package scrapers

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/City-Bureau/chicovidchat/pkg/directory"
)

// ScrapeFoodBanks loads resources from the food bank list
func ScrapeFoodBanks(r io.Reader) ([]directory.Resource, error) {
	var resources []directory.Resource
	document, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return resources, err
	}

	document.Find(".list--row").Each(func(index int, element *goquery.Selection) {
		name, _ := element.Attr("data-location-name")
		zipCode, _ := element.Attr("data-location-zip")
		id, _ := element.Attr("data-location-id")
		// TODO: Include category?
		addrStr := strings.TrimSpace(element.Find(".location--address").First().Text())
		locInfo := strings.TrimSpace(element.Find(".location--info").First().Text())

		phoneRe := regexp.MustCompile(`\(\d{3}\)\s?\d{3}-\d{4}`)
		phoneNumber := phoneRe.FindString(locInfo)

		hoursRe := regexp.MustCompile(`HOURS\:\s+.*?\n`)
		// TODO: Multiple hours in some
		hours := strings.TrimSpace(strings.Replace(hoursRe.FindString(locInfo), "HOURS:", "", 1))

		// TODO: Description should pull from info
		// TODO: Level should be based on service area
		resource := directory.Resource{
			Name:        name,
			Phone:       phoneNumber,
			Description: locInfo,
			Hours:       hours,
			Address:     addrStr,
			ZIP:         zipCode,
			Category:    []string{"Food"},
			// Level:
			Type:       "Nonprofit",
			ExternalID: fmt.Sprintf("gfdc-%s", id),
		}

		log.Println(locInfo)
		log.Println(resource)
		resources = append(resources, resource)
	})
	return resources, nil
}
