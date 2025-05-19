// pkg/models/models.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatorID string    `json:"creator_id"`
	MemberIDs []string  `json:"member_ids"`
	IsPrivate bool      `json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
}

func NewRoom(name string, creatorID string, isPrivate bool) *Room {
	return &Room{
		ID:        uuid.New().String(),
		Name:      name,
		CreatorID: creatorID,
		MemberIDs: []string{creatorID},
		IsPrivate: isPrivate,
		CreatedAt: time.Now(),
	}
}
