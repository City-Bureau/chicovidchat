package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfreiberg/gotwilio"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	values, err := url.ParseQuery(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var smsWebhook gotwilio.SMSWebhook
	err = gotwilio.DecodeWebhook(values, &smsWebhook)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	snsClient := svc.NewSNSClient()
	createdAt := time.Now()
	message := chat.Message{
		ID:        smsWebhook.MessageSid, // TODO: Is it this or SmsSid
		Sender:    smsWebhook.From,
		Recipient: smsWebhook.To,
		Body:      smsWebhook.Body,
		CreatedAt: &createdAt,
	}
	messageJSON, _ := json.Marshal(message)
	log.Println(string(messageJSON))
	err = snsClient.Publish(string(messageJSON), os.Getenv("SNS_TOPIC_ARN"), svc.ReceivedMessageFeed)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       `<?xml verson="1.0" encoding="UTF-8"?><Response></Response>`,
		Headers:    map[string]string{"content-type": "text/xml"},
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
