package db

import (
	"encoding/json"
	"fmt"
	"raychat/models"
	"time"
)

// These methods mostly work for CLI app
func (store *ValkeyChatStore) GetUserByUUIDCLI(uuid string) (*models.UserCred, error) {
	// Get user data by UUID
	userData, err := store.Client.Get(store.Ctx, "user:"+uuid).Result()
	if err != nil {
		return nil, err
	}

	var user models.UserCred
	if err := json.Unmarshal([]byte(userData), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *ValkeyChatStore) SaveUserCredentialsCLI(user *models.UserCred) error {
	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Store user by ID
	err = store.Client.Set(store.Ctx, "user:"+user.UUID, userJSON, 0).Err()
	if err != nil {
		return err
	}

	// Create an index by email for login lookups
	return store.Client.Set(store.Ctx, "user:email:"+user.Email, user.UUID, 0).Err()
}

// GetUserByEmail retrieves a user by email address
func (store *ValkeyChatStore) GetUserByEmailCLI(email string) (*models.UserCred, error) {
	// First get the user ID from the email index
	userID, err := store.Client.Get(store.Ctx, "user:email:"+email).Result()
	if err != nil {
		return nil, err
	}

	// Then get the full user data
	userData, err := store.Client.Get(store.Ctx, "user:"+userID).Result()
	if err != nil {
		return nil, err
	}

	var user models.UserCred
	if err := json.Unmarshal([]byte(userData), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

//these function are made for the main app

// Save complete user profile
func (s *ValkeyChatStore) SaveUserProfile(user *models.User) error {
	userKey := "user:" + user.UUID

	// Store main user hash
	err := s.Client.HSet(s.Ctx, userKey,
		"name", user.Name,
		"email", user.Email,
		"phone", user.Phone,
		"uuid", user.UUID,
		"created_at", user.CreatedAt.Format(time.RFC3339),
		"last_login", user.LastLogin.Format(time.RFC3339),
		"status", user.Status,
	).Err()

	if err != nil {
		return err
	}

	// Create lookup keys for email and phone
	emailKey := "user:email:" + user.Email
	phoneKey := "user:phone:" + user.Phone

	// Set lookup keys with expiration (optional)
	if err := s.Client.Set(s.Ctx, emailKey, user.UUID, 0).Err(); err != nil {
		return err
	}

	if err := s.Client.Set(s.Ctx, phoneKey, user.UUID, 0).Err(); err != nil {
		return err
	}

	return nil
}

// Find user by email
func (s *ValkeyChatStore) FindUserByEmail(email string) (*models.User, error) {
	emailKey := "user:email:" + email

	// Get UUID from email lookup
	uuid, err := s.Client.Get(s.Ctx, emailKey).Result()
	if err != nil {
		return nil, fmt.Errorf("user not found by email: %w", err)
	}

	// Get full user profile
	return s.GetUserByUUID(uuid)
}

// Find user by phone
func (s *ValkeyChatStore) FindUserByPhone(phone string) (*models.User, error) {
	phoneKey := "user:phone:" + phone

	// Get UUID from phone lookup
	uuid, err := s.Client.Get(s.Ctx, phoneKey).Result()
	if err != nil {
		return nil, fmt.Errorf("user not found by phone: %w", err)
	}

	// Get full user profile
	return s.GetUserByUUID(uuid)
}

// Get complete user profile by UUID
func (s *ValkeyChatStore) GetUserByUUID(uuid string) (*models.User, error) {
	userKey := "user:" + uuid

	userData, err := s.Client.HGetAll(s.Ctx, userKey).Result()
	if err != nil {
		return nil, err
	}

	if len(userData) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	user := &models.User{
		UUID:   userData["uuid"],
		Name:   userData["name"],
		Email:  userData["email"],
		Phone:  userData["phone"],
		Status: userData["status"],
	}

	// Parse timestamps
	if createdAt, err := time.Parse(time.RFC3339, userData["created_at"]); err == nil {
		user.CreatedAt = createdAt
	}

	if lastLogin, err := time.Parse(time.RFC3339, userData["last_login"]); err == nil {
		user.LastLogin = lastLogin
	}

	return user, nil
}

// Add user to room authorization
func (s *ValkeyChatStore) AddUserToRoomAuth(userUUID, roomID string) error {
	userRoomsKey := "user:" + userUUID + ":rooms"
	return s.Client.SAdd(s.Ctx, userRoomsKey, roomID).Err()
}

// Get all rooms user is authorized for
func (s *ValkeyChatStore) GetUserAuthorizedRooms(userUUID string) ([]string, error) {
	userRoomsKey := "user:" + userUUID + ":rooms"
	return s.Client.SMembers(s.Ctx, userRoomsKey).Result()
}

// Search users by partial email or phone (for user discovery)
func (s *ValkeyChatStore) SearchUsers(query string) ([]*models.User, error) {
	// Use SCAN to find matching lookup keys
	var cursor uint64
	var users []*models.User

	for {
		// Search email patterns
		emailPattern := "user:email:*" + query + "*"
		emailKeys, newCursor, err := s.Client.Scan(s.Ctx, cursor, emailPattern, 10).Result()
		if err != nil {
			break
		}

		// Get UUIDs and fetch user profiles
		for _, key := range emailKeys {
			uuid, err := s.Client.Get(s.Ctx, key).Result()
			if err != nil {
				continue
			}

			user, err := s.GetUserByUUID(uuid)
			if err != nil {
				continue
			}

			users = append(users, user)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return users, nil
}
