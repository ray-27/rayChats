package models

import "time"

type User struct {
	UUID      string `json:"uuid"`
	Name      string `json:"username"`
	Email     string
	Phone     string
	CreatedAt time.Time
	LastLogin time.Time
	Status    string

	Rooms map[string]bool `json:"rooms"` // Map of room IDs the user belongs to
}

// Room represent a chat room
// type Room struct {
// 	ID                string             `json:"id"`
// 	Name              string             `json:"name"`
// 	CreatorID         string             `json:"creator_id"`
// 	AuthorizedMembers map[string]bool    `json:"members"`
// 	ActiveMembers     map[string]*Client `json:"active_members"`
// 	Admins            map[string]bool    // Users with admin privileges
// 	IsPrivate         bool               `json:"is_private"`
// 	CreatedAt         time.Time
// }

type Message struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp,omitempty"`
}
