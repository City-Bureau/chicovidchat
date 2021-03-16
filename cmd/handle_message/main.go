package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/directory"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

func handleReceivedMessage(message chat.Message, conversation *chat.Conversation, db *gorm.DB) ([]chat.Message, error) {
	var directoryChat directory.DirectoryChat
	err := json.Unmarshal(conversation.Data.RawMessage, &directoryChat)
	if err != nil {
		return []chat.Message{}, err
	}
	replies, replyErr := directoryChat.HandleMessage(message)
	if replyErr != nil {
		return []chat.Message{}, replyErr
	}
	updateErr := directory.UpdateDirectoryChatConversation(&directoryChat, conversation, db)
	if updateErr != nil {
		return []chat.Message{}, updateErr
	}
	return replies, nil
}

func handleSentMessage(message chat.Message, conversation *chat.Conversation, db *gorm.DB) error {
	var directoryChat directory.DirectoryChat
	err := json.Unmarshal(conversation.Data.RawMessage, &directoryChat)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	lastMessage := directoryChat.Messages[len(directoryChat.Messages)-1]
	// TODO: Update this conditional
	if lastMessage.ID != message.ID && (*lastMessage.CreatedAt).Before(*message.CreatedAt) {
		directoryChat.Messages = append(directoryChat.Messages, message)
	}

	return directory.UpdateDirectoryChatConversation(&directoryChat, conversation, db)
}

func handler(request events.SNSEvent) error {
	if len(request.Records) < 1 {
		return nil
	}
	snsRecord := request.Records[0].SNS

	var message chat.Message
	err := json.Unmarshal([]byte(snsRecord.Message), &message)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	db, dbErr := gorm.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s",
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_PORT"),
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_DB_NAME"),
		os.Getenv("RDS_PASSWORD"),
	))
	if dbErr != nil {
		sentry.CaptureException(dbErr)
		return dbErr
	}
	defer db.Close()
	snsClient := svc.NewSNSClient()

	if feed, ok := snsRecord.MessageAttributes["feed"]; ok {
		feedMap, _ := feed.(map[string]interface{})
		feedVal, _ := feedMap["Value"].(string)
		if feedVal == svc.ReceivedMessageFeed {
			conversation, _ := directory.GetOrCreateConversationFromMessage(message.Sender, message, db)
			replies, replyErr := handleReceivedMessage(message, conversation, db)
			if replyErr != nil {
				sentry.CaptureException(replyErr)
				return replyErr
			}
			repliesJSON, _ := json.Marshal(replies)
			return snsClient.Publish(string(repliesJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SendSMSFeed)
		} else if feedVal == svc.SentMessageFeed {
			conversation, _ := directory.GetOrCreateConversationFromMessage(message.Recipient, message, db)
			return handleSentMessage(message, conversation, db)
		} else {
			log.Printf("No handler for feed %s", feedVal)
			return nil
		}
	} else {
		log.Println("Feed not present in SNS message")
		return nil
	}
}

func main() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Transport: &sentry.HTTPSyncTransport{
			Timeout: 5 * time.Second,
		},
	})

	lambda.Start(handler)
}
