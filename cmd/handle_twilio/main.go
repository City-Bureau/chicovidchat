package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/schema"
	"github.com/sfreiberg/gotwilio"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	values, err := url.ParseQuery(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	client := gotwilio.NewTwilioClient(
		os.Getenv("TWILIO_ACCOUNT_SID"),
		os.Getenv("TWILIO_AUTH_TOKEN"),
	)
	twilioChat := svc.NewTwilioChat(client, os.Getenv("TWILIO_FROM"), "")
	isValid, signatureErr := twilioChat.CheckSignature(
		fmt.Sprintf("%s%s", os.Getenv("GW_ENDPOINT"), request.Path),
		request.Headers["X-Twilio-Signature"],
		values,
	)
	if signatureErr != nil {
		return events.APIGatewayProxyResponse{}, signatureErr
	}
	if !isValid {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("Twilio signature is not valid")
	}

	var smsWebhook gotwilio.SMSWebhook

	// Create custom decoder ignoring keys not in webhook struct
	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)
	formDecoder.SetAliasTag("form")
	err = formDecoder.Decode(&smsWebhook, values)

	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	snsClient := svc.NewSNSClient()
	createdAt := time.Now()
	message := chat.Message{
		ID:        smsWebhook.MessageSid,
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
		Body:       `<?xml version="1.0" encoding="UTF-8"?><Response></Response>`,
		Headers:    map[string]string{"content-type": "text/html"},
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
