package chat

import (
	"raychat/models"
	"time"

	"github.com/google/uuid"
)

func NewUser(username string) *models.User {
	return &models.User{
		ID:       uuid.New().String(),
		UserName: username,
	}
}

func NewRoom(name string, cretorID string, isPrivate bool) *models.Room {
	return &models.Room{
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
func NewMessage(roomID, senderID, content, msgType string) *models.Message {
	return &models.Message{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now().Unix(),
	}
}
