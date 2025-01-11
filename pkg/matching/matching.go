package matching

import (
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
	queue     []*User
	queueLock sync.Mutex

	peerMap   = make(map[string]*User)
	peerLock  sync.Mutex
)

// NewUser creates a new User with a unique ID.
func NewUser(conn *websocket.Conn) *User {
	return &User{
		ID:   fmt.Sprintf("user-%d", len(queue)+1),
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
	queue = queue[2:] // Remove matched users
	SetPeer(user1, user2)

	notifyUser(user1, user2)
	notifyUser(user2, user1)
}

// notifyUser sends a match notification to a user.
func notifyUser(user *User, peer *User) {
	if peer == nil {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte("Your peer disconnected."))
		if err != nil {
			fmt.Printf("Failed to notify user %s: %v\n", user.ID, err)
			HandleDisconnect(user)
		}
		return
	}

	err := user.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Matched with %s", peer.ID)))
	if err != nil {
		fmt.Printf("Failed to notify user %s: %v\n", user.ID, err)
		HandleDisconnect(user)
	}
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
