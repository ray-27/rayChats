// db/valkey_store.go
package db

import (
	"context"
	"encoding/json"
	"raychat/models"
	"strings"

	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyChatStore struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewValkeyChatStore(addr string, password string, db int) *ValkeyChatStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db, // Use a specific DB to isolate your chat app
	})

	return &ValkeyChatStore{
		Client: client,
		Ctx:    context.Background(),
	}
}

// SaveUser stores user information
func (s *ValkeyChatStore) SaveUser(userID, username string) error {
	key := "chat:user:" + userID
	return s.Client.HSet(s.Ctx, key,
		"username", username,
		"created_at", time.Now().Format(time.RFC3339),
	).Err()
}

// SaveRoom stores room information
func (s *ValkeyChatStore) SaveRoom(room *models.Room) error {
	key := "chat:room:" + room.ID

	// Convert maps to JSON for storage
	authMembers, _ := json.Marshal(room.AuthorizedMembers)
	activeMembers, _ := json.Marshal(room.ActiveMembers)
	admins, _ := json.Marshal(room.Admins)

	return s.Client.HSet(s.Ctx, key,
		"name", room.Name,
		"creator_id", room.CreatorID,
		"is_private", room.IsPrivate,
		"created_at", room.CreatedAt.Format(time.RFC3339),
		"auth_members", string(authMembers),
		"active_members", string(activeMembers),
		"admins", string(admins),
	).Err()
}

// GetRoomDetails retrieves room information
func (s *ValkeyChatStore) GetRoomDetails(roomID string) (map[string]string, error) {
	key := "chat:room:" + roomID
	return s.Client.HGetAll(s.Ctx, key).Result()
}

// AddUserToRoom authorizes a user for a room
func (s *ValkeyChatStore) AddUserToRoom(userID, roomID string) error {
	// Add room to user's room list
	userRoomsKey := "chat:user:" + userID + ":rooms"
	if err := s.Client.SAdd(s.Ctx, userRoomsKey, roomID).Err(); err != nil {
		return err
	}

	// Add user to room's authorized members
	roomMembersKey := "chat:room:" + roomID + ":auth"
	return s.Client.SAdd(s.Ctx, roomMembersKey, userID).Err()
}

// RemoveUserFromRoom removes a user from a room
func (s *ValkeyChatStore) RemoveUserFromRoom(userID, roomID string) error {
	// Remove room from user's room list
	userRoomsKey := "chat:user:" + userID + ":rooms"
	if err := s.Client.SRem(s.Ctx, userRoomsKey, roomID).Err(); err != nil {
		return err
	}

	// Remove user from room's authorized members
	roomMembersKey := "chat:room:" + roomID + ":auth"
	if err := s.Client.SRem(s.Ctx, roomMembersKey, userID).Err(); err != nil {
		return err
	}

	// Remove user from room's active members
	roomActiveKey := "chat:room:" + roomID + ":active"
	return s.Client.SRem(s.Ctx, roomActiveKey, userID).Err()
}

// GetUserRooms retrieves all rooms a user is authorized for
func (s *ValkeyChatStore) GetUserRooms(userID string) ([]string, error) {
	userRoomsKey := "chat:user:" + userID + ":rooms"
	return s.Client.SMembers(s.Ctx, userRoomsKey).Result()
}

// IsUserAuthorizedForRoom checks if a user is authorized for a room
func (s *ValkeyChatStore) IsUserAuthorizedForRoom(userID, roomID string) (bool, error) {
	roomMembersKey := "chat:room:" + roomID + ":auth"
	return s.Client.SIsMember(s.Ctx, roomMembersKey, userID).Result()
}

// SetUserActive marks a user as active in a room
func (s *ValkeyChatStore) SetUserActive(userID, roomID string) error {
	roomActiveKey := "chat:room:" + roomID + ":active"
	return s.Client.SAdd(s.Ctx, roomActiveKey, userID).Err()
}

// SetUserInactive marks a user as inactive in a room
func (s *ValkeyChatStore) SetUserInactive(userID, roomID string) error {
	roomActiveKey := "chat:room:" + roomID + ":active"
	return s.Client.SRem(s.Ctx, roomActiveKey, userID).Err()
}

// GetActiveUsers gets all active users in a room
func (s *ValkeyChatStore) GetActiveUsers(roomID string) ([]string, error) {
	roomActiveKey := "chat:room:" + roomID + ":active"
	return s.Client.SMembers(s.Ctx, roomActiveKey).Result()
}

// GetAllRoomIDs retrieves all room IDs from storage
func (s *ValkeyChatStore) GetAllRoomIDs() ([]string, error) {
	// Use pattern matching to find all room keys
	keys, err := s.Client.Keys(s.Ctx, "chat:room:*").Result()
	if err != nil {
		return nil, err
	}

	// Extract room IDs from keys
	roomIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		// Skip keys that have additional segments (like :auth, :active, etc.)
		if strings.Count(key, ":") == 2 {
			// Extract the ID part from "chat:room:ID"
			parts := strings.Split(key, ":")
			if len(parts) == 3 {
				roomIDs = append(roomIDs, parts[2])
			}
		}
	}

	return roomIDs, nil
}

// GetRoomAuthorizedMembers gets all authorized members for a room
func (s *ValkeyChatStore) GetRoomAuthorizedMembers(roomID string) ([]string, error) {
	key := "chat:room:" + roomID + ":auth"
	return s.Client.SMembers(s.Ctx, key).Result()
}

// GetRoomAdmins gets all admins for a room
func (s *ValkeyChatStore) GetRoomAdmins(roomID string) ([]string, error) {
	key := "chat:room:" + roomID + ":admins"
	return s.Client.SMembers(s.Ctx, key).Result()
}
