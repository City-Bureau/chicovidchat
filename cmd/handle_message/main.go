package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/directory"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

func getOrCreateConversationFromMessage(message chat.Message, db *gorm.DB) (*chat.Conversation, bool) {
	var conversation chat.Conversation
	if db.Where("data ->> 'id' = ? AND (data ->> 'active')::boolean IS TRUE", message.Sender).Take(&conversation).RecordNotFound() {
		chat := directory.NewDirectoryChat(message.Sender)
		updateConversationChat(chat, &conversation, db)
		return &conversation, true
	}
	return &conversation, false
}

func updateConversationChat(directoryChat *directory.DirectoryChat, conversation *chat.Conversation, db *gorm.DB) error {
	chatJSON, _ := json.Marshal(directoryChat)
	conversation.Data = postgres.Jsonb{
		RawMessage: json.RawMessage(chatJSON),
	}

	db.Save(conversation)
	return nil
}

func handleReceivedMessage(message chat.Message, conversation *chat.Conversation, db *gorm.DB) (chat.Message, error) {
	var directoryChat directory.DirectoryChat
	err := json.Unmarshal(conversation.Data.RawMessage, &directoryChat)
	if err != nil {
		return chat.Message{}, err
	}
	reply, replyErr := directoryChat.HandleMessage(message)
	if replyErr != nil {
		return chat.Message{}, replyErr
	}
	updateErr := updateConversationChat(&directoryChat, conversation, db)
	if updateErr != nil {
		return chat.Message{}, updateErr
	}
	return reply, nil
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

	return updateConversationChat(&directoryChat, conversation, db)
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
	db, _ := gorm.Open("postgresql", fmt.Sprintf(
		"%s:%s@tcp(%s:5432)/%s",
		os.Getenv("RDS_USERNAME"),
		os.Getenv("RDS_PASSWORD"),
		os.Getenv("RDS_HOST"),
		os.Getenv("RDS_DB_NAME"),
	))
	// TODO: Use created
	conversation, _ := getOrCreateConversationFromMessage(message, db)
	snsClient := svc.NewSNSClient()

	if feed, ok := snsRecord.MessageAttributes[""]; ok {
		if feed == svc.ReceivedMessageFeed {
			reply, replyErr := handleReceivedMessage(message, conversation, db)
			if replyErr != nil {
				return replyErr
			}
			replyJSON, _ := json.Marshal(reply)
			return snsClient.Publish(string(replyJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SendSMSFeed)
		} else if feed == svc.SentMessageFeed {
			return handleSentMessage(message, conversation, db)
		} else {
			// TODO: Should throw some kind of error
			log.Println(feed)
			return nil
		}
	} else {
		// TODO: Should be error parsing
		return nil
	}
}

func main() {
	lambda.Start(handler)
}
