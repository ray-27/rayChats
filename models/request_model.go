package models

import "time"

type CreateRoomRequest struct {
	RoomCode string   `json:"roomCode" binding:"required"`
	Roominfo RoomInfo `json:"roominfo" binding:"required"`
}

type RoomInfo struct {
	Name        string `json:"name" binding:"required"`
	CreatorID   string `json:"creator_id" binding:"required"`
	RoomType    string `json:"room_type" binding:"required"`
	IsPrivate   bool   `json:"is_private"`
	Description string `json:"room_description"`
	// CountLimit   int           `json:"count_limit"`
	Participants []ContactInfo `json:"participants"`
	Timestamp    time.Time     `json:"timestamp"`
}

type RoomDataPayload struct {
	Name         string        `json:"name"`
	CreatorID    string        `json:"creator_id"`
	RoomType     string        `json:"room_type"`
	IsPrivate    bool          `json:"is_private"`
	Description  string        `json:"room_description"`
	CountLimit   int           `json:"count_limit"`
	Participants []ContactInfo `json:"participants"`
	Timestamp    time.Time     `json:"timestamp"`
}

type CreateRoomResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	RoomID  string `json:"room_id,omitempty"`
}

type ContactInfo struct {
	Name        string  `json:"name" binding:"required"`
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	ContactType string  `json:"contact_type" binding:"required"`
}

type AddUserToRoomPayload struct {
	UserID   string `json:"user_id" binding:"required"`
	RoomCode string `json:"room_code" bidinding:"required"`
	JoinTime string `json:"join_time"`
	Type     string `json:"type"`
}
