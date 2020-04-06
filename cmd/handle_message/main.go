package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
		return err
	}
	db, _ := gorm.Open("postgres", fmt.Sprintf(
		"%s:%s@tcp(%s:5432)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	conversation, _ := directory.GetOrCreateConversationFromMessage(message, db)
	snsClient := svc.NewSNSClient()

	if feed, ok := snsRecord.MessageAttributes["feed"]; ok {
		if feed == svc.ReceivedMessageFeed {
			replies, replyErr := handleReceivedMessage(message, conversation, db)
			if replyErr != nil {
				return replyErr
			}
			repliesJSON, _ := json.Marshal(replies)
			return snsClient.Publish(string(repliesJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SendSMSFeed)
		} else if feed == svc.SentMessageFeed {
			return handleSentMessage(message, conversation, db)
		} else {
			log.Printf("No handler for feed %s", feed)
			return nil
		}
	} else {
		log.Println("Feed not present in SNS message")
		return nil
	}
}

func main() {
	lambda.Start(handler)
}
