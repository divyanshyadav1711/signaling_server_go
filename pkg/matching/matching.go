package matching

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	ID   string
	Conn *websocket.Conn
}

var (
	id=0        
	queue     []*User
	queueLock sync.Mutex

	peerMap   = make(map[string]*User)
	peerLock  sync.Mutex
)

// NewUser creates a new User with a unique ID.
func NewUser(conn *websocket.Conn) *User {
	id++;
	return &User{
		ID:   fmt.Sprintf("user-%d", id),
		Conn: conn,
	}
}

// QueueUser adds a user to the matching queue.
func QueueUser(user *User) {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue = append(queue, user)

	go func() {
		select {
		case <-time.After(5 * time.Minute): // Timeout for waiting in the queue
			RemoveUser(user)
		}
	}()
	matchUsers()
}

// RemoveUser removes a user from the matching queue.
func RemoveUser(user *User) {
	queueLock.Lock()
	defer queueLock.Unlock()
	for i, u := range queue {
		if u == user {
			queue = append(queue[:i], queue[i+1:]...)
			break
		}
	}
}

// matchUsers matches two users from the queue and establishes a peer relationship.
func matchUsers() {
	if len(queue) < 2 {
		return
	}

	user1 := queue[0]
	user2 := queue[1]
	queue = queue[2:] 
	SetPeer(user1, user2)

	notifyUser(user1, user2)
	notifyUser(user2, user1)
}

func notifyUser(user *User, peer *User) {
	if peer == nil {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte("Your peer disconnected."))
		if err != nil {
			fmt.Printf("Failed to notify user %s: %v\n", user.ID, err)
			HandleDisconnect(user)
		}
		return
	}
	// Sending start-call message to the user
	user.Conn.WriteJSON(map[string]interface{}{
		"type":    "start-call",
		"message": "Start your call now!",
	})
}

// GetPeer retrieves the peer of a user.
func GetPeer(user *User) *User {
	peerLock.Lock()
	defer peerLock.Unlock()
	return peerMap[user.ID]
}

// SetPeer sets up a peer relationship between two users.
func SetPeer(user1, user2 *User) {
	peerLock.Lock()
	defer peerLock.Unlock()
	peerMap[user1.ID] = user2
	peerMap[user2.ID] = user1
}

// RemovePeer removes a user's peer relationship.
func RemovePeer(user *User) {
	peerLock.Lock()
	defer peerLock.Unlock()
	peer := peerMap[user.ID]
	delete(peerMap, user.ID)
	if peer != nil {
		delete(peerMap, peer.ID)
	}
}

// HandleDisconnect cleans up resources when a user disconnects.
func HandleDisconnect(user *User) {
	RemoveUser(user)
	HandlePeerDisconnect(user)
	user.Conn.Close()
}

// HandlePeerDisconnect handles a peer's disconnection.
func HandlePeerDisconnect(user *User) {
	peer := GetPeer(user)
	if peer != nil {
		notifyUser(peer, nil) // Notify the peer of disconnection
		RemovePeer(peer)
		QueueUser(peer) // Optionally re-queue the peer
	}
	RemovePeer(user)
}


// StartCallPayload represents the payload structure sent to the client
type StartCallPayload struct {
	Message string `json:"message"` // The message content
}

// sendStartCall sends a "start-call" event with a structured payload
func sendStartCall(conn *websocket.Conn, message string) error {
	// Create the payload
	payload := StartCallPayload{
		Message: message,
	}

	// Serialize the payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Send the "start-call" event with the payload
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send start-call event: %w", err)
	}

	return nil
}