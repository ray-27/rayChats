// handlers/chat_handlers.go
package chat

import (
	"fmt"
	"net/http"
	db "raychat/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// HandleGetRoom gets details about a specific room
func HandleGetRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	userID := c.GetString("user_id")

	// Use the exported GetRoom function
	room, exists := GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Check if user is authorized for this room
	if room.IsPrivate {
		if _, ok := room.AuthorizedMembers[userID]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized for this room"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             room.ID,
		"name":           room.Name,
		"is_private":     room.IsPrivate,
		"active_members": len(room.ActiveMembers),
		"total_members":  len(room.AuthorizedMembers),
	})
}

// HandleCreateRoom creates a new chat room
// a room is created between two users (at least)
func HandleCreateRoom(c *gin.Context) {
	userUUID, _ := c.Get("userUUID")
	userUUIDStr := userUUID.(string)

	var req struct {
		RoomName    string   `json:"roomname" binding:"required"`
		GuestEmails []string `json:"guest_emails" binding:"required,dive,email"`
		// GuestPhone	[]string `json:"guest_emails" binding:"required,dive,email"`
		IsPrivate bool `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that at least one guest email is provided
	if len(req.GuestEmails) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one guest email is required"})
		return
	}

	// Create room with creator as admin
	room := CreateRoom(req.RoomName, userUUIDStr, req.IsPrivate)

	// Track successful and failed additions
	var successfulUsers []string
	var failedUsers []string

	// Add each guest user to the room
	for _, email := range req.GuestEmails {
		// Get guest user UUID by email
		guestUUID, err := db.Valkey.GetUserUUIDByEmail(email)
		if err != nil {
			failedUsers = append(failedUsers, email)
			continue
		}

		// Add guest user as authorized member
		if err := db.Valkey.AddUserToRoom(guestUUID, room.ID); err != nil {
			failedUsers = append(failedUsers, email)
			continue
		}

		successfulUsers = append(successfulUsers, email)
	}

	response := gin.H{
		"room_id":          room.ID,
		"name":             room.Name,
		"successful_users": successfulUsers,
	}

	if len(failedUsers) > 0 {
		response["failed_users"] = failedUsers
		response["message"] = "Room created with some users added successfully"
	} else {
		response["message"] = "Room created successfully with all users added"
	}

	c.JSON(http.StatusCreated, response)
}

func HandleAddUsertoRoom(c *gin.Context) {
	var req struct {
		RoomID    string `json:"room_id" binding:"required"`
		UserEmail string `json:"user_email" binding:"required"`
		MakeAdmin bool   `json:"make_admin"` // Optional: make the added user an admin
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get requesting user UUID from auth middleware
	requestingUserUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requestingUserUUIDStr := requestingUserUUID.(string)

	// Check if requesting user is admin of the room using existing function
	isAdmin, err := db.Valkey.IsUserAdmin(requestingUserUUIDStr, req.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add users to this room"})
		return
	}

	// Get user UUID by email using existing function
	userUUID, err := db.Valkey.GetUserUUIDByEmail(req.UserEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found with email: " + req.UserEmail,
		})
		return
	}

	// Add user to room using existing functions
	if req.MakeAdmin {
		err = db.Valkey.AddAdminToRoom(userUUID, req.RoomID)
	} else {
		err = db.Valkey.AddUserToRoom(userUUID, req.RoomID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add user to room",
		})
		return
	}

	role := "member"
	if req.MakeAdmin {
		role = "admin"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("User added to room as %s", role),
		"room_id": req.RoomID,
	})
}

// Update your HandleListRooms function
func HandleListRooms(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUIDStr := userUUID.(string)

	// Get rooms for user from Valkey
	rooms, err := db.Valkey.GetRoomsForUser(userUUIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve rooms",
		})
		return
	}

	// Add user role information and active count to each room
	roomsWithStatus := make([]map[string]interface{}, 0, len(rooms))
	for _, room := range rooms {
		isAdmin := room.Admins != nil && room.Admins[userUUIDStr] //to see of the user is the admin

		// Get active users count
		activeUsers, err := db.Valkey.GetActiveUsers(room.ID)
		activeCount := 0
		if err == nil {
			activeCount = len(activeUsers)
		}

		roomInfo := map[string]interface{}{
			"id":           room.ID,
			"name":         room.Name,
			"creator_id":   room.CreatorID,
			"is_private":   room.IsPrivate,
			"created_at":   room.CreatedAt,
			"is_admin":     isAdmin,
			"member_count": len(room.AuthorizedMembers),
			"active_count": activeCount,
		}

		roomsWithStatus = append(roomsWithStatus, roomInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": roomsWithStatus,
	})
}

// success := AddAuthorizedMember(room.ID, req.SecondUser_id, userID)
// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *gin.Context) {

	userID := c.Query("user_id")
	userName := c.Query("username")

	if userID == "" || userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Use the exported HandleWebSocketConnection function
	HandleWebSocketConnection(userID, userName, conn)
}

// RegisterChatRoutes registers all chat-related routes
func RegisterChatRoutes(router *gin.Engine) {
	chatGroup := router.Group("/chat")
	// chatGroup.Use(auth.AuthRequired())
	{
		chatGroup.GET("/rooms/:roomId", HandleGetRoom)
		chatGroup.GET("/rooms", HandleListRooms)
		chatGroup.POST("/createroom", HandleCreateRoom)
		chatGroup.POST("/addusertoroom", HandleAddUsertoRoom)
		chatGroup.GET("/ws", HandleWebSocket)
	}
}
