package db

import (
	"encoding/json"
	"raychat/models"
)

func (store *ValkeyChatStore) SaveUserCredentials(user *models.UserCred) error {
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
func (store *ValkeyChatStore) GetUserByEmail(email string) (*models.UserCred, error) {
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
