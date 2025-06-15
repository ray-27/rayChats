package chat

import (
	"fmt"
	"log"
	db "raychat/database"
	"strings"
	"time"
)

type Room struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	CreatorID         string             `json:"creator_id"`
	AuthorizedMembers map[string]bool    `json:"members"`
	ActiveMembers     map[string]*Client `json:"active_members"`
	Admins            map[string]bool    // Users with admin privileges
	IsPrivate         bool               `json:"is_private"`
	CreatedAt         time.Time
}

func LoadAllRoomsWithMembersFromValkey() ([]*Room, error) {
	//Get all room IDs from the Valkey database
	keys, err := db.Valkey.Client.Keys(db.Valkey.Ctx, "chat:room:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load the rooms from the valkey database.")
	}

	//Extract the roomIDs from the keys
	roomIDs := make([]string, 0)
	for _, key := range keys {
		//Skip the keys that have extra segments like :auth, :active
		if strings.Count(key, ":") == 2 {
			//Extract the id part of fron "chat:room:ID"
			parts := strings.Split(key, ":")
			if len(parts) == 3 {
				roomIDs = append(roomIDs, parts[2])
			}
		}
	}

	log.Printf("Found %d rooms in Valkey", len(roomIDs))

	rooms := make([]*Room, 0, len(roomIDs))
	for _, roomID := range roomIDs {
		roomKey := "chat:room:" + roomID
		roomData, err := db.Valkey.Client.HGetAll(db.Valkey.Ctx, roomKey).Result() //get the room data

		if err != nil {
			log.Printf("Error getting the details for room %s: %v", roomID, err)
			continue //skip that rooom
		}

		//Parse the room data
		room := &Room{
			ID:        roomID,
			Name:      roomData["name"],
			CreatorID: roomData["creator_id"],
			IsPrivate: roomData["is_private"] == "true",
		}

		//Initialize auth, maps
		room.AuthorizedMembers = make(map[string]bool)
		room.ActiveMembers = make(map[string]*Client)
		room.Admins = make(map[string]bool)

		//Get auth membets
		authMembersKey := "chat:room:" + roomID + ":auth"
		authMembers, err := db.Valkey.Client.SMembers(db.Valkey.Ctx, authMembersKey).Result()
		if err != nil {
			log.Printf("Error getting the authorized memebers for the room %s: %v", roomID, err)
		} else {
			for _, member := range authMembers {
				room.AuthorizedMembers[member] = true
			}
		}

		// Get admins
		adminsKey := "chat:room:" + roomID + ":admins"
		admins, err := db.Valkey.Client.SMembers(db.Valkey.Ctx, adminsKey).Result()
		if err != nil {
			log.Printf("Error getting admins for room %s: %v", roomID, err)
		} else {
			for _, admin := range admins {
				room.Admins[admin] = true
			}
		}

		// Always ensure creator is both authorized member and admin
		room.AuthorizedMembers[room.CreatorID] = true
		room.Admins[room.CreatorID] = true

		rooms = append(rooms, room)
		log.Printf("Loaded room: %s, name: %s, authorized members: %d, admins: %d",
			room.ID, room.Name, len(room.AuthorizedMembers), len(room.Admins))
	}

	log.Printf("Successfuly loaded %d rooms from Valkey", len(rooms))
	return rooms, nil
}
