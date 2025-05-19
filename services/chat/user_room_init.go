package chat

import (
	"time"

	"github.com/google/uuid"
)

func NewUser(username string) *User {
	return &User{
		ID:       uuid.New().String(),
		UserName: username,
	}
}

func NewRoom(name string, cretorID string, isPrivate bool) *Room {
	return &Room{
		ID:                uuid.New().String(),
		Name:              name,
		CreatorID:         cretorID,
		AuthorizedMembers: map[string]bool{cretorID: true}, //Add cretor as the first member
		ActiveMembers:     make(map[string]bool),           //Initially empty
		Admins:            map[string]bool{cretorID: true}, //Creator is automatically an admin
		IsPrivate:         isPrivate,
		CreatedAt:         time.Now(),
	}
}

// NewMessage creates a new message, though this should not be needed in use,
func NewMessage(roomID, senderID, content, msgType string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
	}
}
