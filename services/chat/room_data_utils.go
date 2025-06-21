package chat

import (
	"fmt"
	"log"
	db "raychat/database"
	"raychat/models"
	"time"
)

func StoreRoomInValkey(roomID string, roomInfo models.RoomInfo) error {
	// Prepare keys
	roomKey := fmt.Sprintf("chat:room:%s", roomID)
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)
	adminsKey := fmt.Sprintf("chat:room:%s:admins", roomID)

	// Store main room data as hash
	roomHash := map[string]interface{}{
		"name":        roomInfo.Name,
		"creator_id":  roomInfo.CreatorID,
		"is_private":  roomInfo.IsPrivate,
		"room_type":   roomInfo.RoomType,
		"description": roomInfo.Description,
		"created_at":  roomInfo.Timestamp.Format(time.RFC3339),
	}

	err := db.Valkey.Client.HSet(db.Valkey.Ctx, roomKey, roomHash).Err()
	if err != nil {
		return fmt.Errorf("failed to store room data: %w", err)
	}

	// Store authorized members as set
	err = db.Valkey.Client.SAdd(db.Valkey.Ctx, authKey, roomInfo.CreatorID).Err()
	if err != nil {
		return fmt.Errorf("failed to add creator to auth members: %w", err)
	}

	// Add participants as authorized members if private
	if roomInfo.IsPrivate {
		for _, participant := range roomInfo.Participants {
			userID := participant.Name
			if participant.Email != nil && *participant.Email != "" {
				userID = *participant.Email
			}
			db.Valkey.Client.SAdd(db.Valkey.Ctx, authKey, userID)
		}
	}

	// Store admins as set
	err = db.Valkey.Client.SAdd(db.Valkey.Ctx, adminsKey, roomInfo.CreatorID).Err()
	if err != nil {
		return fmt.Errorf("failed to add creator to admins: %w", err)
	}

	// Set expiration on all keys (24 hours)
	db.Valkey.Client.Expire(db.Valkey.Ctx, roomKey, 24*time.Hour)
	db.Valkey.Client.Expire(db.Valkey.Ctx, authKey, 24*time.Hour)
	db.Valkey.Client.Expire(db.Valkey.Ctx, adminsKey, 24*time.Hour)

	return nil
}

func LoadRoomFromValkey(roomID string) (*Room, error) {
	roomKey := fmt.Sprintf("chat:room:%s", roomID)
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)
	adminsKey := fmt.Sprintf("chat:room:%s:admins", roomID)

	// Get main room data
	roomData, err := db.Valkey.Client.HGetAll(db.Valkey.Ctx, roomKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room data: %w", err)
	}

	if len(roomData) == 0 {
		return nil, fmt.Errorf("room not found")
	}

	// Parse created_at
	var createdAt time.Time
	if createdAtStr, exists := roomData["created_at"]; exists {
		createdAt, _ = time.Parse(time.RFC3339, createdAtStr)
	}

	// Parse is_private
	isPrivate := roomData["is_private"] == "true"

	// Create room struct
	room := &Room{
		ID:                roomID,
		Name:              roomData["name"],
		CreatorID:         roomData["creator_id"],
		AuthorizedMembers: make(map[string]bool),
		ActiveMembers:     make(map[string]*Client),
		Admins:            make(map[string]bool),
		IsPrivate:         isPrivate,
		CreatedAt:         createdAt,
	}

	// Get authorized members
	authMembers, err := db.Valkey.Client.SMembers(db.Valkey.Ctx, authKey).Result()
	if err != nil {
		log.Printf("Error getting authorized members for room %s: %v", roomID, err)
	} else {
		for _, member := range authMembers {
			room.AuthorizedMembers[member] = true
		}
	}

	// Get admins
	adminMembers, err := db.Valkey.Client.SMembers(db.Valkey.Ctx, adminsKey).Result()
	if err != nil {
		log.Printf("Error getting admins for room %s: %v", roomID, err)
	} else {
		for _, admin := range adminMembers {
			room.Admins[admin] = true
		}
	}

	// Always ensure creator is both authorized member and admin
	room.AuthorizedMembers[room.CreatorID] = true
	room.Admins[room.CreatorID] = true

	return room, nil
}

func AddUserToRoomAuthMembers(roomID, userID string) error {
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)

	// Add user to authorized members set
	err := db.Valkey.Client.SAdd(db.Valkey.Ctx, authKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to authorized members: %w", err)
	}

	// Reset expiration to maintain consistency
	db.Valkey.Client.Expire(db.Valkey.Ctx, authKey, 24*time.Hour)

	return nil
}

// Remove user from authorized members
func RemoveUserFromRoomAuthMembers(roomID, userID string) error {
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)

	err := db.Valkey.Client.SRem(db.Valkey.Ctx, authKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove user from authorized members: %w", err)
	}

	return nil
}

// Check if user is authorized member
func IsUserAuthorizedMember(roomID, userID string) (bool, error) {
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)

	exists, err := db.Valkey.Client.SIsMember(db.Valkey.Ctx, authKey, userID).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user authorization: %w", err)
	}

	return exists, nil
}

// Get all authorized members of a room
func GetRoomAuthMembers(roomID string) ([]string, error) {
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)

	members, err := db.Valkey.Client.SMembers(db.Valkey.Ctx, authKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get authorized members: %w", err)
	}

	return members, nil
}

// Add user as admin (also adds to authorized members)
func AddUserAsRoomAdmin(roomID, userID string) error {
	authKey := fmt.Sprintf("chat:room:%s:auth", roomID)
	adminsKey := fmt.Sprintf("chat:room:%s:admins", roomID)

	// Add to authorized members first
	err := db.Valkey.Client.SAdd(db.Valkey.Ctx, authKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to authorized members: %w", err)
	}

	// Add to admins
	err = db.Valkey.Client.SAdd(db.Valkey.Ctx, adminsKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to admins: %w", err)
	}

	// Reset expiration
	db.Valkey.Client.Expire(db.Valkey.Ctx, authKey, 24*time.Hour)
	db.Valkey.Client.Expire(db.Valkey.Ctx, adminsKey, 24*time.Hour)

	return nil
}
