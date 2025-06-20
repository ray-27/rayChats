package db

import (
	"encoding/json"
	"fmt"
	"raychat/models"
	"time"
)

func (store *ValkeyChatStore) StoreRoomInValkey(roomID string, data models.RoomDataPayload) error {
	// Convert to Valkey format
	valkeyData, err := ConvertToValkeyFormat(data)
	if err != nil {
		return err
	}

	// Store in Valkey with key format "room:id"
	roomKey := "room:" + roomID

	err = store.Client.HSet(store.Ctx, roomKey, valkeyData).Err()
	if err != nil {
		return err
	}

	// Set expiration (optional)
	store.Client.Expire(store.Ctx, roomKey, 24*time.Hour)

	return nil
}

func ConvertToValkeyFormat(data models.RoomDataPayload) (map[string]interface{}, error) {
	// Serialize participants to JSON string
	participantsJSON, err := json.Marshal(data.Participants)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal participants: %w", err)
	}	

	valkeyData := map[string]interface{}{
		"name":         data.Name,
		"creator_id":   data.CreatorID,
		"room_type":    data.RoomType,
		"is_private":   data.IsPrivate,
		"description":  data.Description,
		"participants": string(participantsJSON),
		"count_limit":  data.CountLimit,
		"created_at":   data.Timestamp.Format(time.RFC3339),
		"is_active":    true,
	}

	return valkeyData, nil
}
