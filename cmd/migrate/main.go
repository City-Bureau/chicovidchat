package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
)

func handler(request events.CloudWatchEvent) error {
	db, err := gorm.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s",
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_PORT"),
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_DB_NAME"),
		os.Getenv("RDS_PASSWORD"),
	))
	if err != nil {
		return err
	}
	// db.DropTable(&chat.Conversation{})
	db.AutoMigrate(&chat.Conversation{})
	defer db.Close()

	return nil
}

func main() {
	lambda.Start(handler)
}
