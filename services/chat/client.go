package chat

import "github.com/gorilla/websocket"

//a client represents a connected chat user, (online user)

type Client struct {
	ID       string
	UserID   string
	UserName string
	Conn     *websocket.Conn
	Manager  *ChatManager
	Send     chan *Message
	Rooms    map[string]bool
}
