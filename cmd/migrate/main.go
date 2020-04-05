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
	db, _ := gorm.Open("postgres", fmt.Sprintf(
		"%s:%s@tcp(%s:5432)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	// db.DropTable(&chat.Conversation{})
	db.AutoMigrate(&chat.Conversation{})
	defer db.Close()

	return nil
}

func main() {
	lambda.Start(handler)
}
