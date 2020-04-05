package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/City-Bureau/chicovidchat/pkg/directory"
)

func handler(request events.CloudWatchEvent) error {
	airtableBase := os.Getenv("AIRTABLE_BASE")
	airtableTable := os.Getenv("AIRTABLE_TABLE")
	airtableKey := os.Getenv("AIRTABLE_KEY")
	records, err := directory.LoadAirtableResources(airtableBase, airtableTable, airtableKey)
	if err != nil {
		return err
	}

	recordsJSON, jsonErr := json.Marshal(records)
	if jsonErr != nil {
		return err
	}

	client := session.New()
	_, err = s3.New(client).PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET")),
		Key:         aws.String("latest.json"),
		ACL:         aws.String("private"),
		Body:        bytes.NewReader(recordsJSON),
		ContentType: aws.String("application/json"),
	})

	return err
}

func main() {
	lambda.Start(handler)
}
