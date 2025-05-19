package chat

import "time"

type User struct {
	ID       string `json:"uuid"`
	UserName string `json:"username"`
}

// Room represent a chat room
type Room struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	CreatorID         string          `json:"creator_id"`
	AuthorizedMembers map[string]bool `json:"members"`
	ActiveMembers     map[string]bool `json:"active_members"`
	Admins            map[string]bool // Users with admin privileges
	IsPrivate         bool            `json:"is_private"`
	CreatedAt         time.Time
}

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
