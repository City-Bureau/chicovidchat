package svc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// ReceivedMessageFeed is the feed name for handling received messages
const ReceivedMessageFeed = "handle_received_message"

// SentMessageFeed is the feed name for handling sent messages
const SentMessageFeed = "handle_sent_message"

// SendSMSFeed is the feed name for sending a Twilio SMS message
const SendSMSFeed = "send_twilio_sms"

// SNS is an interface for the SNSClient and associated mock
type SNS interface {
	Publish(string, string, string) error
}

// SNSClient implements SNS for a generic way of managing the SNS service
type SNSClient struct {
	Client *sns.SNS
}

// NewSNSClient creates an SNSClient object
func NewSNSClient() *SNSClient {
	client := sns.New(session.New())
	return &SNSClient{Client: client}
}

// Publish sends a message to a given topic and feed
func (c *SNSClient) Publish(message string, topicArn string, feed string) error {
	_, err := c.Client.Publish(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicArn),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"feed": &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(feed),
			},
		},
	})
	return err
}
