package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/schema"
	"github.com/jinzhu/gorm"
	"github.com/sfreiberg/gotwilio"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

const optInStr string = "covid"

func handleChatSMS(smsWebhook gotwilio.SMSWebhook) error {
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

	return snsClient.Publish(string(messageJSON), os.Getenv("SNS_TOPIC_ARN"), svc.ReceivedMessageFeed)
}

func proxyTwilioRequest(request events.APIGatewayProxyRequest, values url.Values) error {
	client := &http.Client{}
	r, _ := http.NewRequest("POST", os.Getenv("SPOKE_ENDPOINT"), strings.NewReader(values.Encode()))

	r.Header.Add("X-Spoke-Proxy-Endpoint", fmt.Sprintf("%s%s", os.Getenv("GW_ENDPOINT"), request.Path))
	for k, v := range request.Headers {
		r.Header.Add(k, v)
	}

	_, err := client.Do(r)
	return err
}

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

	// Can access DB because if Spoke enabled an NAT must be set up
	db, dbErr := gorm.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s",
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_PORT"),
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_DB_NAME"),
		os.Getenv("RDS_PASSWORD"),
	))
	if dbErr != nil {
		return events.APIGatewayProxyResponse{}, dbErr
	}
	defer db.Close()

	var conversation chat.Conversation
	isNewContact := db.Model(&chat.Conversation{}).Where("data ->> 'id' = ?", smsWebhook.From).Last(&conversation).RecordNotFound()
	isOptIn := strings.ToLower(strings.TrimSpace(smsWebhook.Body)) == optInStr
	if (isNewContact && isOptIn) || !isNewContact {
		// Proxy the initial opt-in message to Spoke for clarity
		if isOptIn {
			proxyErr := proxyTwilioRequest(request, values)
			if proxyErr != nil {
				return events.APIGatewayProxyResponse{}, err
			}
		}
		err = handleChatSMS(smsWebhook)
	} else {
		err = proxyTwilioRequest(request, values)
	}
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
