// models/message.go
package models

// import (
// 	"time"

// 	"github.com/google/uuid"
// )

// type Message struct {
// 	ID           string    `json:"id"`
// 	SenderUUID   string    `json:"sender_uuid"`
// 	ReceiverUUID string    `json:"receiver_uuid"`
// 	Content      string    `json:"content"`
// 	Timestamp    time.Time `json:"timestamp"`
// }

// func NewMessage(senderUUID, receiverUUID, content string) *Message {
// 	return &Message{
// 		ID:           uuid.New().String(),
// 		SenderUUID:   senderUUID,
// 		ReceiverUUID: receiverUUID,
// 		Content:      content,
// 		Timestamp:    time.Now(),
// 	}
// }
